package yolo

import (
	"context"
	. "go-sip/logger"
	. "go-sip/sip"
	"time"

	"go.uber.org/zap"
	"gocv.io/x/gocv"

	"github.com/swdee/go-rknnlite"
	"github.com/swdee/go-rknnlite/preprocess"
	"github.com/swdee/go-rknnlite/render"
)

const (
	// 模型置信度阈值
	AI_MODEL_PROBABILITY = 0.7
)

type StreamProcessor struct {
	DeviceId       string
	RkPlatform     string
	StreamId       string
	VedioStreamUrl string
	Rts            []*YoloRuntime
	CancelFunc     context.CancelFunc
	YoloResultCh   chan map[string]YoloDetResultStat
}

type YoloRuntime struct {
	Rt             *rknnlite.Runtime
	StreamId       string
	RkModelName    string
	RkModelPath    string
	RkPlatform     string
	ClassNames     []string
	ClassNameScore map[string]float64 // 存储类别名称和对应的分数
	Confidence     float32            // 置信度
	frame_interval int                // 帧间隔
	RtClosed       chan string        // 新增: 每个rt关闭的通知
}

// YOLO识别结构统计
type YoloDetResultStat struct {
	ClassId  int     // 类别ID
	Count    int     // 类别个数
	MaxScore float32 // 最高分数
}

type YoloLastInfo struct {
	ClassName      string    // 类别ID
	LastDetTime    time.Time // 上一次输出时间
	LastNotifyTime time.Time // 上一次通知时间
}

// 视频流处理
func (sp *StreamProcessor) VedioStreamProcess(ctx context.Context) {
	defer close(sp.YoloResultCh)

	// 使用cv2打开视频流（或视频文件），并创建一个可用于后续逐帧读取的对象
	cap, err := gocv.OpenVideoCapture(sp.VedioStreamUrl)
	if err != nil {
		Logger.Error("无法打开视频流: ", zap.Any("error", err))
		return
	}
	defer cap.Close()
	// 判断是否成功打开视频流
	if !cap.IsOpened() {
		Logger.Error("无法打开视频流")
		return
	}
	// 设置cap buffer size
	cap.Set(gocv.VideoCaptureBufferSize, 1)

	// 帧计数器
	frameCount := 0

	yoloPostProcesser := YoloPostProcess()

	rgbImg := gocv.NewMat()
	defer rgbImg.Close()

	cropImg := gocv.NewMat()
	defer cropImg.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := cap.Read(&img); !ok || img.Empty() {
		Logger.Error("摄像头流读取失败", zap.Any("stream url ", sp.VedioStreamUrl), zap.Error(err))
		return
	}

	resizer := preprocess.NewResizer(img.Cols(), img.Rows(), 640, 640)
	modelTriggerMonitor := GetAiModelTriggerMonitor(sp.StreamId, 20*time.Second)

	for {
		select {
		case <-ctx.Done():
			Logger.Info("VedioStreamProcess已退出")
			return
		default:
			// 读取视频帧
			if ok := cap.Read(&img); !ok || img.Empty() {
				Logger.Debug("读取失败，2秒后重试...", zap.Any("streamId", sp.StreamId))
				time.Sleep(2 * time.Second)
				continue
			}
			frameCount++
			yoloDetResultStatMap := make(map[string]YoloDetResultStat, 0)
			// 遍历ai模型，进行推理
			for _, rt := range sp.Rts {
				select {
				case <-rt.RtClosed:
					Logger.Info("worker[%s] 模型已关闭，退出循环", zap.Any("streamId", sp.StreamId))
					return
				default:
					if rt == nil || rt.Rt == nil {
						continue
					}
					if rt.ClassNames == nil || len(rt.ClassNames) == 0 {
						continue
					}
					if frameCount%rt.frame_interval != 0 {
						continue
					}
					gocv.CvtColor(img, &rgbImg, gocv.ColorBGRToRGB)
					resizer.LetterBoxResize(rgbImg, &cropImg, render.Black)
					// 对图像文件执行推理
					outputs, err := rt.Rt.Inference([]gocv.Mat{cropImg})
					if err != nil || outputs == nil {
						Logger.Error("运行时推理异常", zap.Any("模型名称", rt.RkModelName), zap.Any("模型平台", rt.RkPlatform), zap.Any("error", err))
						continue
					}
					if err == nil && outputs != nil {
						if rt.Confidence < AI_MODEL_PROBABILITY || rt.Confidence > 1 {
							rt.Confidence = 0.85
						}
						detectObjs := yoloPostProcesser.DetectObjects(outputs, resizer)
						detectResults := detectObjs.GetDetectResults()
						for _, detResult := range detectResults {
							className := rt.ClassNames[detResult.Class]
							if detResult.Probability > rt.Confidence && className != "" {
								var maxScore float32
								if detResult.Probability > yoloDetResultStatMap[className].MaxScore {
									maxScore = detResult.Probability
								} else {
									maxScore = yoloDetResultStatMap[className].MaxScore
								}
								yoloDetResultStatMap[className] = YoloDetResultStat{
									ClassId:  detResult.Class + 1,
									Count:    yoloDetResultStatMap[className].Count + 1,
									MaxScore: maxScore,
								}
								// 模型触发业务
								modelTriggerMonitor.AiEventTrigger(sp.StreamId, className, rt.RkPlatform)
							}
						}
						outputs.Free()
					}
				}
			}
			if len(yoloDetResultStatMap) > 0 {
				// 发到 channel
				select {
				case sp.YoloResultCh <- yoloDetResultStatMap:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (sp *StreamProcessor) YoloResultProcess() {
	yoloLastInfoMap := make(map[string]*YoloLastInfo)
	for result := range sp.YoloResultCh {
		for className, stat := range result {
			yoloLastInfo := yoloLastInfoMap[className]
			if yoloLastInfo == nil {
				yoloLastInfo = &YoloLastInfo{
					ClassName:      className,
					LastDetTime:    time.Now(),
					LastNotifyTime: time.Now(),
				}
			} else {
				if yoloLastInfo.LastDetTime.Add(time.Duration(5) * time.Second).Before(time.Now()) {
					delete(yoloLastInfoMap, className)
					continue
				} else {
					yoloLastInfo.LastDetTime = time.Now()
					// 5秒通知一次
					if yoloLastInfo.LastNotifyTime.Add(time.Duration(5) * time.Second).Before(time.Now()) {
						// 更新上一次通知时间
						yoloLastInfo.LastNotifyTime = time.Now()
						NotifyAiEventFunc(sp.DeviceId, sp.RkPlatform, sp.StreamId, int64(stat.ClassId), className, float64(stat.MaxScore), int64(stat.Count))
					}
				}
			}
			yoloLastInfoMap[className] = yoloLastInfo
		}
	}
}

// 线程安全的关闭方法
func (sp *StreamProcessor) Stop() {
	// 关闭运行时并释放资源
	sp.CancelFunc()
	for _, rt := range sp.Rts {
		if rt != nil && rt.Rt != nil {
			err := rt.Rt.Close()
			if err != nil {
				Logger.Error("关闭RKNN运行时失败: ", zap.String("streamId", rt.StreamId), zap.Any("error", err))
			} else {
				Logger.Info("RKNN运行时已关闭", zap.String("streamId", rt.StreamId), zap.String("modelName", rt.RkModelName), zap.String("platform", rt.RkPlatform))
				close(rt.RtClosed)
			}
		} else {
			Logger.Warn("RKNN运行时为空，无法关闭")
		}
	}
	Logger.Info("停止视频流完成", zap.String("streamId", sp.StreamId))

}
