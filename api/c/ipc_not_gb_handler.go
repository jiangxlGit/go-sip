package api

import (
	"fmt"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"
	"go-sip/zlm_api"
	"log"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// 推流控制参数
const (
	MaxRetryCount = 10
)

type IpcPushStreamData struct {
	// 摄像头ipcid
	StreamId string
	// 内网ip
	InnerIP string
	// 用户名
	Username string
	// 密码
	Password string
	// rtsp流后缀
	RtspSuffix string
	// 是否是主码流，true 主码流，false 辅码流
	IsMainStream bool
	// zlmIp
	ZlmIp string
	// zlmSecret
	ZlmSecret string
}

func IpcPushStreamHandler(deviceId string) {
	notGbIpcList, err := GetNotGbIpcList(deviceId)
	if err != nil {
		Logger.Error("GetNotGbIpcList error", zap.Any("deviceId", deviceId), zap.Error(err))
		return
	}
	for _, ipcInfo := range notGbIpcList {
		ok := utils.CheckPort(ipcInfo.InnerIP, "554", 3*time.Second)
		if !ok {
			IpcNotGbInfoUpdate(ipcInfo.IpcId, "OFFLINE")
			continue
		}

		subData := &IpcPushStreamData{
			StreamId:     fmt.Sprintf("%s_%s", ipcInfo.IpcId, "0"),
			InnerIP:      ipcInfo.InnerIP,
			Username:     ipcInfo.Username,
			Password:     ipcInfo.Password,
			RtspSuffix:   ipcInfo.RtspSubSuffix,
			IsMainStream: false,
			ZlmIp:        m.CMConfig.ZlmInnerIp,
			ZlmSecret:    m.CMConfig.ZlmSecret,
		}
		startStreamDaemonWithRetry(subData)

		mainData := &IpcPushStreamData{
			StreamId:     fmt.Sprintf("%s_%s", ipcInfo.IpcId, "1"),
			InnerIP:      ipcInfo.InnerIP,
			Username:     ipcInfo.Username,
			Password:     ipcInfo.Password,
			RtspSuffix:   ipcInfo.RtspMainSuffix,
			IsMainStream: true,
			ZlmIp:        m.CMConfig.ZlmInnerIp,
			ZlmSecret:    m.CMConfig.ZlmSecret,
		}
		startStreamDaemonWithRetry(mainData)
	}
}

func IpcPushStreamReset(deviceId, ipcId string, zlmInfo *model.ZlmInfo) error {
	if deviceId == "" {
		return fmt.Errorf("deviceId is empty")
	}
	notGbIpcList, err := GetNotGbIpcList(deviceId)
	if err != nil {
		Logger.Error("IpcPushStreamReset error", zap.Any("deviceId", deviceId), zap.Error(err))
		return fmt.Errorf("查询非国标ipc列表失败")
	}
	for _, ipcInfo := range notGbIpcList {
		if ipcId != "" && ipcInfo.IpcId != ipcId {
			continue
		}
		ip := ipcInfo.InnerIP
		if ip == "" {
			continue
		}

		ok := utils.CheckPort(ip, "554", 3*time.Second)
		if !ok {
			IpcNotGbInfoUpdate(ipcInfo.IpcId, "ERROR")
			continue
		}

		// 判断流是否存在zlm
		var zlmGetMediaListReq = zlm_api.ZlmGetMediaListReq{}
		zlmGetMediaListReq.App = "rtp"
		zlmGetMediaListReq.Vhost = "__defaultVhost__"
		zlmGetMediaListReq.Schema = "rtsp"
		zlmGetMediaListReq.StreamID = fmt.Sprintf("%s_%s", ipcInfo.IpcId, "0")
		resp := zlm_api.ZlmGetMediaList(fmt.Sprintf("http://%s:9092", m.CMConfig.ZlmInnerIp), m.CMConfig.ZlmSecret, zlmGetMediaListReq)
		if resp.Code == 0 && len(resp.Data) > 0 {
			Logger.Debug("推流重置子码流存在", zap.Any("streamId", zlmGetMediaListReq.StreamID))
			IpcNotGbInfoUpdate(ipcInfo.IpcId, "ON")
		} else {
			Logger.Error("推流重置子码流不存在", zap.Any("streamId", zlmGetMediaListReq.StreamID))
			subData := &IpcPushStreamData{
				StreamId:     fmt.Sprintf("%s_%s", ipcInfo.IpcId, "0"),
				InnerIP:      ip,
				Username:     ipcInfo.Username,
				Password:     ipcInfo.Password,
				RtspSuffix:   ipcInfo.RtspSubSuffix,
				IsMainStream: false,
			}
			// 先杀死旧的进程
			err := KillFfmpegIfExist(subData.StreamId)
			if err == nil {
				err = ffmpegPushStreamAsync(subData)
				if err != nil {
					Logger.Error("子码流重新推流失败", zap.Any("streamId", subData.StreamId), zap.Error(err))
					IpcNotGbInfoUpdate(ipcInfo.IpcId, "ERROR")
				}
			}
		}

		var zlmIp string
		var zlmHost string
		var zlmSecret string
		if zlmInfo == nil {
			zlmIp = m.CMConfig.ZlmInnerIp
			zlmHost = fmt.Sprintf("http://%s:9092", m.CMConfig.ZlmInnerIp)
			zlmSecret = m.CMConfig.ZlmSecret
		} else {
			zlmIp = zlmInfo.ZlmIp
			zlmHost = zlmInfo.ZlmDomain
			zlmSecret = zlmInfo.ZlmSecret
		}
		zlmGetMediaListReq.StreamID = fmt.Sprintf("%s_%s", ipcInfo.IpcId, "1")
		resp2 := zlm_api.ZlmGetMediaList(zlmHost, zlmSecret, zlmGetMediaListReq)
		if resp2.Code == 0 && len(resp2.Data) > 0 {
			Logger.Debug("推流重置主码流存在", zap.Any("streamId", zlmGetMediaListReq.StreamID))
			IpcNotGbInfoUpdate(ipcInfo.IpcId, "ON")
		} else {
			Logger.Error("推流重置主码流不存在", zap.Any("streamId", zlmGetMediaListReq.StreamID))
			mainData := &IpcPushStreamData{
				StreamId:     fmt.Sprintf("%s_%s", ipcInfo.IpcId, "1"),
				InnerIP:      ip,
				Username:     ipcInfo.Username,
				Password:     ipcInfo.Password,
				RtspSuffix:   ipcInfo.RtspMainSuffix,
				IsMainStream: true,
				ZlmIp:        zlmIp,
				ZlmSecret:    zlmSecret,
			}
			// 先杀死旧的进程
			err = KillFfmpegIfExist(mainData.StreamId)
			if err == nil {
				err = ffmpegPushStreamAsync(mainData)
				if err != nil {
					Logger.Error("主码流重新推流失败", zap.Any("streamId", mainData.StreamId), zap.Error(err))
					IpcNotGbInfoUpdate(ipcInfo.IpcId, "ERROR")
				}
			}
		}

	}
	return nil
}

func KillFfmpegIfExist(streamId string) error {
	match := fmt.Sprintf("ffmpeg.*%s", streamId)

	for i := 1; i <= 3; i++ {
		// 检查是否存在匹配进程
		checkCmd := exec.Command("pgrep", "-f", match)
		output, err := checkCmd.CombinedOutput()

		if err != nil || len(strings.TrimSpace(string(output))) == 0 {
			Logger.Debug("没有匹配的进程", zap.Any("streamId", streamId))
			// 没有匹配进程，退出重试
			return nil
		}
		// 有匹配进程，尝试 kill
		killCmd := exec.Command("pkill", "-f", match)
		err = killCmd.Run()
		if err != nil {
			Logger.Debug("尝试杀死 ffmpeg 进程失败", zap.String("streamId", streamId), zap.Error(err))
		} else {
			return nil
		}
		// 等待 1 秒后下一轮
		time.Sleep(1 * time.Second)
	}

	// 最后再确认是否杀成功
	finalCheck := exec.Command("pgrep", "-f", match)
	finalOutput, _ := finalCheck.CombinedOutput()
	if len(strings.TrimSpace(string(finalOutput))) != 0 {
		Logger.Error("ffmpeg 进程仍存在，杀死失败", zap.String("streamId", streamId), zap.String("pids", string(finalOutput)))
		return fmt.Errorf("ffmpeg 杀死失败")
	} else {
		Logger.Debug("ffmpeg 杀死成功", zap.String("streamId", streamId))
		return nil
	}
}

func startStreamDaemonWithRetry(s *IpcPushStreamData) {
	// 先杀死旧的进程
	KillFfmpegIfExist(s.StreamId)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				Logger.Error("推流协程 panic", zap.Any("recover", r), zap.String("streamId", s.StreamId))
			}
		}()
		var zlmGetMediaListReq = zlm_api.ZlmGetMediaListReq{}
		zlmGetMediaListReq.App = "rtp"
		zlmGetMediaListReq.Vhost = "__defaultVhost__"
		zlmGetMediaListReq.Schema = "rtsp"
		zlmGetMediaListReq.StreamID = s.StreamId
		for i := 1; i <= MaxRetryCount; i++ {
			ffmpegPushStreamSync(s)
			// 最后一次失败就直接退出，不再 sleep
			if i == MaxRetryCount {
				break
			}
			Logger.Warn("推流失败，", zap.Int("次数：", i), zap.String("streamId", s.StreamId))
			time.Sleep(20 * time.Second)
			resp := zlm_api.ZlmGetMediaList(fmt.Sprintf("http://%s:9092", m.CMConfig.ZlmInnerIp), m.CMConfig.ZlmSecret, zlmGetMediaListReq)
			if resp.Code == 0 && len(resp.Data) > 0 {
				Logger.Debug("推流到本地zlm成功", zap.String("streamId", zlmGetMediaListReq.StreamID))
				return
			}
		}
		Logger.Error("推流最终失败，超过最大重试次数", zap.String("streamId", s.StreamId), zap.Int("maxRetry", MaxRetryCount))
	}()
}

func ffmpegPushStream(req *IpcPushStreamData) *exec.Cmd {
	if req.IsMainStream {
		return ffmpegPushMainStream(req)
	} else {
		return ffmpegPushSubStream(req)
	}
}

// 推送主码流
func ffmpegPushMainStream(req *IpcPushStreamData) *exec.Cmd {
	// 源 RTSP 地址
	input := fmt.Sprintf("rtsp://%s:%s@%s:554%s", req.Username, req.Password, req.InnerIP, req.RtspSuffix)
	// 推送目标 RTSP 地址
	output := fmt.Sprintf("rtsp://%s:554/rtp/%s", req.ZlmIp, req.StreamId)
	// 构造 cmd
	cmd := exec.Command("ffmpeg",
		// "-loglevel", "info", // 只输出错误日志
		"-rtsp_transport", "tcp",
		// "-rtsp_transport", "udp", // 内网低延迟可用 UDP
		// "-timeout", "30000000", // 连接/读超时（微秒）
		"-i", input, // 输入流
		// "-fflags", "+nobuffer+discardcorrupt", // 不缓冲 + 遇到坏帧直接丢
		// "-flags", "+low_delay", // 开启低延迟模式
		// "-analyzeduration", "500000", // 快速分析流
		// "-probesize", "100000", // 降低探测时间
		// "-flush_packets", "1", // 每帧立刻推送
		// "-use_wallclock_as_timestamps", "1", // 使用系统时间戳
		// "-max_delay", "200000", // 限制缓冲最大延迟
		// "-tune", "zerolatency", // 低延迟优化
		"-c", "copy",
		// "-c:v", "libx264", // 视频设置成h264编码
		// "-b:v", "512k", // 平均视频码率
		// "-maxrate", "512k", // 最大视频码率
		// "-bufsize", "1024k", // VBV缓冲区（一般设为2倍）
		// "-vf", "scale=1920:1080,fps=15", // 统一分辨率和帧率（可调）
		"-c:a", "pcm_alaw", // 音频编码为 AAC
		"-ar", "8000", // 音频采样率
		"-f", "rtsp", // 输出为 RTSP
		output,
	)
	Logger.Debug("ffmpeg命令: ", zap.String("cmd", cmd.String()))
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd
}

// 推送子码流
func ffmpegPushSubStream(req *IpcPushStreamData) *exec.Cmd {
	// 源 RTSP 地址
	input := fmt.Sprintf("rtsp://%s:%s@%s:554%s", req.Username, req.Password, req.InnerIP, req.RtspSuffix)
	// 推送目标 RTSP 地址
	output := fmt.Sprintf("rtsp://%s:554/rtp/%s", req.ZlmIp, req.StreamId)
	// 构造 cmd
	cmd := exec.Command("ffmpeg",
		// "-loglevel", "info", // 只输出错误日志
		"-rtsp_transport", "tcp",
		// "-rtsp_transport", "udp", // 内网低延迟可用 UDP
		// "-timeout", "30000000", // 连接/读超时（微秒）
		"-i", input, // 输入流
		// "-fflags", "+nobuffer+discardcorrupt", // 不缓冲 + 遇到坏帧直接丢
		// "-flags", "+low_delay", // 开启低延迟模式
		// "-analyzeduration", "500000", // 快速分析流
		// "-probesize", "100000", // 降低探测时间
		// "-flush_packets", "1", // 每帧立刻推送
		// "-use_wallclock_as_timestamps", "1", // 使用系统时间戳
		// "-max_delay", "200000", // 限制缓冲最大延迟
		// "-tune", "zerolatency", // 低延迟优化
		"-c", "copy",
		// "-c:v", "libx264", // 视频设置成h264编码
		// "-b:v", "256k", // 平均视频码率
		// "-maxrate", "256k", // 最大视频码率
		// "-bufsize", "512k", // VBV缓冲区（一般设为2倍）
		// "-vf", "scale=640:360,fps=15", // 统一分辨率和帧率（可调）
		"-c:a", "pcm_alaw", // 音频编码为 AAC
		"-ar", "8000", // 音频采样率
		"-f", "rtsp", // 输出为 RTSP
		output,
	)
	Logger.Debug("ffmpeg命令: ", zap.String("cmd", cmd.String()))
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd
}

func ffmpegPushStreamSync(req *IpcPushStreamData) error {
	Logger.Debug("同步开始推流", zap.String("streamId", req.StreamId))
	cmd := ffmpegPushStream(req)
	err := cmd.Run()
	if err != nil {
		Logger.Debug("同步推流函数退出", zap.String("streamId", req.StreamId), zap.Error(err))
	}
	return err
}

func ffmpegPushStreamAsync(req *IpcPushStreamData) error {
	Logger.Info("异步开始推流", zap.String("streamId", req.StreamId))
	cmd := ffmpegPushStream(req)
	// 非阻塞启动
	if err := cmd.Start(); err != nil {
		Logger.Error("推流进程启动失败", zap.String("streamId", req.StreamId), zap.Error(err))
		return err
	}
	// 异步等待退出，避免僵尸进程
	go func(streamId string) {
		err := cmd.Wait()
		Logger.Debug("异步推流进程退出", zap.String("streamId", streamId), zap.Error(err))
	}(req.StreamId)
	return nil
}
