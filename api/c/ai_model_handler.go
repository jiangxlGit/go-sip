package api

import (
	"context"
	"errors"
	"fmt"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	sipapi "go-sip/sip"
	"go-sip/utils"
	"go-sip/yolo"
	. "go-sip/zlm_api"
	"reflect"
	"strings"

	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	IpcYoloModelsMap       = make(map[string][]*model.AiModelInfo)
	AiModelTriggerTimerMap = make(map[string]*time.Timer)
	devicePlatformMap      = make(map[string]string)
)

func ZlmStreamAiModelHandler(ctx context.Context, streamId, app string) error {
	parts := strings.Split(streamId, "_")
	if len(parts) == 2 {
		ipcId := parts[0]
		idx := parts[1]
		yoloModelList := make([]*model.AiModelInfo, 0)
		if idx == "0" {
			// 标清流下只加载默认模型
			defaultYoloModels := IpcYoloModelsMap["default_ai_models"]
			yoloModelList = append(yoloModelList, defaultYoloModels...)
		} else if idx == "1" {
			// 高清流加载非默认模型
			yoloModelList = append(yoloModelList, IpcYoloModelsMap[ipcId]...)
		}
		if len(yoloModelList) == 0 {
			return fmt.Errorf("未配置模型")
		}
		yoloProcessor := &yolo.YoloProcessorStruct{
			VedioStreamUrl: fmt.Sprintf("rtsp://%s:554/%s/%s", m.CMConfig.ZlmInnerIp, app, streamId),
			StreamId:       streamId,
		}
		for deviceId, rkPlatform := range devicePlatformMap {
			yoloProcessor.DeviceId = deviceId
			yoloProcessor.RkPlatform = rkPlatform
		}

		yoloModelBatch := make([]*yolo.YoloModelInfo, 0)

		for _, model := range yoloModelList {
			_labelList := make([]string, 0)
			_labelscore := make(map[string]float64)
			for _, label := range model.ModelLabelList {
				_labelList = append(_labelList, label.LabelName)
				_labelscore[label.LabelName] = label.Score
			}

			_yoloModel := &yolo.YoloModelInfo{
				RkModelName:    model.ModelFileName,
				RkModelPath:    model.ModelLocalFilePath,
				RkPlatform:     model.ModelPlatform,
				ClassNames:     _labelList,
				ClassNameScore: _labelscore,
				Confidence:     model.AiModelConfidence,
			}
			yoloModelBatch = append(yoloModelBatch, _yoloModel)
		}
		yoloProcessor.YoloModelInfoList = yoloModelBatch

		err := yoloProcessor.StartYoloProcessor()
		if err != nil {
			Logger.Error("启动YoloProcessor失败", zap.Error(err))
		} else {
			YoloProcessorMap[streamId] = yoloProcessor
		}
	}
	return nil
}

// 启动模型触发器
func StartAiModelTriggerHandler() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		modelTriggerMonitor := yolo.GlobalModelTriggerMonitor
		if modelTriggerMonitor == nil {
			continue
		}
		cases, keys := modelTriggerMonitor.BuildSelectCases()
		chosen, recv, ok := reflect.Select(cases)
		if chosen == 0 { // updateCh 被触发，重新构建 cases
			continue
		}
		streamId := keys[chosen-1] // 因为 cases[0] 是 updateCh
		if !ok {
			continue
		}
		event := recv.Interface().(yolo.TriggerEvent)
		switch event.Event {
		case "start":
			AiModelStartRecord(ctx, streamId, event.ClassName)
		case "stop":
			AiModelStopRecord(ctx, streamId)
		default:
			Logger.Error("未知事件类型", zap.String("event", event.Event))
		}
	}
}

type RecordManager struct {
	startChanMap sync.Map // map[string]chan bool
	stopChanMap  sync.Map // map[string]chan bool
	mutexMap     sync.Map // map[string]*sync.Mutex
	maxRetry     int
	retryDelay   time.Duration
}

func NewRecordManager() *RecordManager {
	return &RecordManager{
		maxRetry:   5,
		retryDelay: 2 * time.Second,
	}
}

// 获取 streamId 对应的锁（防止并发冲突）
func (rm *RecordManager) getLock(streamId string) *sync.Mutex {
	lock, _ := rm.mutexMap.LoadOrStore(streamId, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func AiModelStartRecord(c context.Context, streamId, className string) {
	rm := NewRecordManager()
	startCh, err := rm.startRecord(c, streamId, className)
	if err != nil {
		Logger.Error("start record failed:", zap.Error(err))
		return
	}
	go func() {
		select {
		case ok := <-startCh:
			if ok {
				Logger.Info("AI模型录制【启动】成功", zap.Any("streamId", streamId), zap.Any("className", className))
			} else {
				Logger.Error("AI模型录制【启动】失败", zap.Any("streamId", streamId), zap.Any("className", className))
				return
			}
		case <-c.Done():
			Logger.Error("AI模型录制【启动】超时或取消", zap.Any("streamId", streamId), zap.Any("className", className))
			return
		}
	}()
}

func AiModelStopRecord(c context.Context, streamId string) {
	rm := NewRecordManager()
	stopCh, err := rm.stopRecord(c, streamId)
	if err != nil {
		Logger.Error("stop record failed:", zap.Error(err))
		return
	}
	go func() {
		select {
		case ok := <-stopCh:
			if ok {
				Logger.Info("AI模型录制【停止】成功", zap.Any("streamId", streamId))
			} else {
				Logger.Error("AI模型录制【停止】失败", zap.Any("streamId", streamId))
				return
			}
		case <-c.Done():
			Logger.Error("AI模型录制停止超时或取消", zap.Any("streamId", streamId))
			return
		}
	}()
}

// 启动录制
func (rm *RecordManager) startRecord(ctx context.Context, streamId, streamType string) (<-chan bool, error) {
	if _, exists := rm.startChanMap.Load(streamId); exists {
		return nil, errors.New("startRecord already in progress")
	}

	resultChan := make(chan bool, 1)
	rm.startChanMap.Store(streamId, resultChan)

	go func() {
		defer func() {
			close(resultChan)
			rm.startChanMap.Delete(streamId)
		}()

		lock := rm.getLock(streamId)
		lock.Lock()
		defer lock.Unlock()

		for i := 0; i < rm.maxRetry; i++ {
			select {
			case <-ctx.Done():
				Logger.Warn("zlm start record canceled:", zap.Any("streamId", streamId))
				resultChan <- false
				return
			default:
				status := ZlmGetRecordStatus(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, streamId)
				if !status.Status {
					resp := ZlmStartRecord(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, streamId, streamType)
					if resp.Code == 0 && resp.Result {
						resultChan <- true
						return
					}
					Logger.Warn("zlm start record failed:", zap.Any("streamId", streamId), zap.Any("resp", resp))
				} else {
					resultChan <- false
					return
				}
				time.Sleep(rm.retryDelay)
			}
		}
		Logger.Error("zlm start record retry exceeded:", zap.Any("streamId", streamId))
		resultChan <- false
	}()

	return resultChan, nil
}

// 停止录制
func (rm *RecordManager) stopRecord(ctx context.Context, streamId string) (<-chan bool, error) {
	if _, exists := rm.stopChanMap.Load(streamId); exists {
		return nil, errors.New("stopRecord already in progress")
	}

	resultChan := make(chan bool, 1)
	rm.stopChanMap.Store(streamId, resultChan)

	go func() {
		defer func() {
			close(resultChan)
			rm.stopChanMap.Delete(streamId)
		}()

		lock := rm.getLock(streamId)
		lock.Lock()
		defer lock.Unlock()

		for i := 0; i < rm.maxRetry; i++ {
			select {
			case <-ctx.Done():
				Logger.Warn("zlm stop record canceled:", zap.Any("streamId", streamId))
				resultChan <- false
				return
			default:
				status := ZlmGetRecordStatus(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, streamId)
				if status.Status {
					resp := ZlmStopRecord(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, streamId)
					if resp.Code == 0 && resp.Result {
						resultChan <- true
						return
					}
					Logger.Warn("zlm stop record failed:", zap.Any("streamId", streamId), zap.Any("resp", resp))
				} else {
					resultChan <- false
					return
				}
				time.Sleep(rm.retryDelay)
			}
		}
		Logger.Error("zlm stop record retry exceeded:", zap.Any("streamId", streamId))
		resultChan <- false
	}()

	return resultChan, nil
}

// 下载设备关联的模型列表
func DownloadDeviceRelationAIModel(deviceId string, deviceType string) {
	if deviceId == "" {
		Logger.Error("设备ID不能为空")
		return
	}
	devicePlatformMap[deviceId] = deviceType
	list, err := GetDeviceRelationAiModelList(deviceId)
	if err != nil {
		Logger.Error("查询失败", zap.Any("err", err))
		return
	}
	if list == nil || len(list) == 0 {
		Logger.Info("没有关联的AI模型")
		return
	}

	// 项目根目录
	rootPath, err := os.Getwd()
	if err != nil {
		Logger.Error("获取当前目录失败", zap.Any("err", err))
		return
	}

	// 指定下载目录
	saveDir, err := utils.FileDirHandler(rootPath, "ai_model")
	if err != nil || saveDir == "" {
		Logger.Error("创建目录失败", zap.Any("err", err))
		return
	}
	for ipcId, modelList := range list {

		IpcDir, err := utils.FileDirHandler(saveDir, ipcId)

		files, err := os.ReadDir(IpcDir)
		if err != nil {
			continue
		}
		AiModelFileMd5Map := make(map[string]string)
		AiModelFilePath := make(map[string]string)
		for _, file := range files {
			if file.IsDir() {
				continue // 跳过目录
			}

			fullPath := filepath.Join(IpcDir, file.Name())
			md5Str, err := utils.ComputeFileMD5(fullPath)
			if err != nil {
				Logger.Error("计算文件MD5失败", zap.Any("fileName", file.Name()), zap.Error(err))
				return
			}
			AiModelFileMd5Map[file.Name()] = md5Str
			AiModelFilePath[file.Name()] = fullPath
		}

		ai_models := make([]*model.AiModelInfo, 0)

		for _, model := range modelList {

			if !strings.EqualFold(model.ModelPlatform, deviceType) {
				Logger.Error("模型平台不匹配，跳过", zap.Any("modelPlatform", model.ModelPlatform), zap.Any("deviceType", deviceType))
				continue
			}

			path := filepath.Join(IpcDir, model.ModelFileName)
			if md5, ok := AiModelFileMd5Map[model.ModelFileName]; ok {
				if md5 != model.ModelFileMd5 {
					Logger.Info("模型文件MD5不匹配,重新下载", zap.Any("ModelFileName", model.ModelFileName), zap.Any("MD5", md5))
					utils.DownloadFileTo(path, model.ModelFileURL)
				}
			} else {
				Logger.Info("模型文件不存在，开始下载", zap.Any("ModelFileName", model.ModelFileName), zap.Any("ModelFileKey", model.ModelFileKey))
				utils.DownloadFileTo(path, model.ModelFileURL)
			}
			_model := model
			_model.ModelLocalFilePath = path
			ai_models = append(ai_models, _model)

		}
		IpcYoloModelsMap[ipcId] = ai_models
	}
	Logger.Info("成功下载已关联的AI模型", zap.Any("deviceId", deviceId), zap.Any("IpcYoloModelsMap", IpcYoloModelsMap))
}
