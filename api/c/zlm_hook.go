package api

import (
	"fmt"
	"net/http"
	"time"

	. "go-sip/db/alioss"
	db "go-sip/db/sqlite"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	sipapi "go-sip/sip"
	"go-sip/utils"
	"go-sip/zlm_api"

	"go-sip/yolo"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var YoloProcessorMap = make(map[string]*yolo.YoloProcessorStruct)

// ZLM WebHook
func ZLMWebHook(c *gin.Context) {
	method := c.Param("method")
	// Logger.Info("client sip ZLMWebHook", zap.String("method", method))
	switch method {
	case "on_server_started":
		// zlm启动
		zlmServerStart(c)
	case "on_http_access":
		// http请求鉴权
	case "on_play":
		// 点播业务
		zlmStreamOnPlay(c)
	case "on_publish":
		// 推流业务
		zlmStreamPublish(c)
	case "on_stream_none_reader":
		// 无人阅读通知 关闭流
		zlmStreamNoneReader(c)
	case "on_stream_not_found":
		// 请求播放时，流不存在时触发
		zlmStreamNotFound(c)
	case "on_record_mp4":
		//  mp4录制完成
		zlmRecordMp4(c)
	case "on_stream_changed":
		// 流注册和注销通知
		zlmStreamChanged(c)
	default:
		m.ZlmWebHookResponse(c, 0, "success")
	}

}

func zlmServerStart(c *gin.Context) {
	m.ZlmWebHookResponse(c, 0, "zlm server start success")
}

// zlm点播业务
func zlmStreamOnPlay(c *gin.Context) {
	m.ZlmWebHookResponse(c, 0, "on play success")
}

func zlmStreamChanged(c *gin.Context) {

	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZLMStreamChangedData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	if req.Regist {
		if req.Schema == "rtsp" && (req.APP == "rtp" || req.APP == "live") {
			if strings.HasPrefix(req.Stream, "IPC") {
				streamArr := strings.Split(req.Stream, "_")
				IpcNotGbInfoUpdate(streamArr[0], "ON")
			}
			Logger.Info("流注册 ", zap.Any("req", req))
			ZlmStreamAiModelHandler(c, req.Stream, req.APP)
		}

	} else {
		if req.Schema == "rtsp" && (req.APP == "rtp" || req.APP == "live") {
			Logger.Info("流注销 :", zap.Any("req", req))
			if _, ok := YoloProcessorMap[req.Stream]; ok {
				yoloProcessor := YoloProcessorMap[req.Stream]
				yoloProcessor.StopYoloProcessor()
			}
			yolo.GlobalModelTriggerMonitor.AiEventTriggerStop(req.Stream)
			delete(YoloProcessorMap, req.Stream)

			if strings.HasPrefix(req.Stream, "IPC") {
				streamArr := strings.Split(req.Stream, "_")
				IpcNotGbInfoUpdate(streamArr[0], "OFFLINE")
				// 停止ffmpeg
				KillFfmpegIfExist(req.Stream)
			} else {
				err = sipapi.SipStopPlay(req.Stream)
				if err != nil {
					Logger.Error("摄像头停止错误 ：", zap.Error(err))
					return
				}
			}
		}
	}
	m.ZlmWebHookResponse(c, 0, "success")
}

func zlmStreamPublish(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"code":         0,
		"enable_audio": true,
		"enable_MP4":   false,
		"msg":          "success",
	})
}

func zlmRecordMp4(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("zlmRecordMp4 body error", zap.Error(err))
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var recordMp4Data = model.ZLMRecordMp4Data{}
	if err := utils.JSONDecode(data, &recordMp4Data); err != nil {
		Logger.Error("zlmRecordMp4 body error", zap.Error(err))
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	m.ZlmWebHookResponse(c, 0, "success")
	go func() {
		var maxRetry = 5
		var j = 1
		var fileOssDownloadUrl = ""
		for ; j <= maxRetry; j++ {
			// 上传到oss
			filePath := recordMp4Data.FilePath
			objectKey := "ipc_video_playback/" + recordMp4Data.Stream + "/" + recordMp4Data.FileName
			fileUrl, err := UploadFile(objectKey, filePath)
			if err != nil || fileUrl == "" {
				Logger.Warn("zlmRecordMp4 upload file error", zap.Error(err))
				// 延迟再试，防止接口频繁调用
				time.Sleep(1 * time.Second)
			} else {
				fileOssDownloadUrl = fileUrl
				// 上传成功，则删除本地文件
				utils.DeleteFile(filePath)
				Logger.Info("zlmRecordMp4 upload file success", zap.String("fileUrl", fileUrl))
				break
			}
		}
		if j > maxRetry {
			Logger.Error("zlmRecordMp4 upload file error", zap.Any("maxRetry", maxRetry))
		} else if fileOssDownloadUrl != "" {
			var i = 1
			for ; i <= maxRetry; i++ {
				recordMp4Data.FileOssDownloadUrl = fileOssDownloadUrl
				// 调用 gateway 接口，存储录像信息
				result := IpcPlaybackRecord(recordMp4Data)
				if result.Code == model.CodeSucc {
					Logger.Info("录像信息上报成功", zap.Any("result", result.Result))
					break
				} else {
					Logger.Warn("录像信息上报失败，将重试")
					// 延迟再试，防止接口频繁调用
					time.Sleep(3 * time.Second)
				}
			}
			if i > maxRetry {
				Logger.Error("录像信息上报重试达到上限，仍然失败", zap.Any("recordMp4Data", recordMp4Data))
			}
		}

	}()

}

func zlmStreamNotFound(c *gin.Context) {

	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZLMStreamNotFoundData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}

	stream_arr := strings.Split(req.Stream, "_")
	// 判断req.Stream是否以IPC开头，不以IPC开头则表示为国标设备
	if !strings.HasPrefix(stream_arr[0], "IPC") {
		sd_stream_id := stream_arr[0] + "_0"
		rtp_info := zlm_api.ZlmStartRtpServer(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, sd_stream_id, "rtp", 0)
		if rtp_info.Code != 0 || rtp_info.Port == 0 {
			Logger.Error("open rtp server fail", zap.Int("code", rtp_info.Code))
			return
		}
		// 向摄像头发送信令请求推实时标清流到zlm
		pm := &sipapi.Streams{ChannelID: stream_arr[0], StreamID: sd_stream_id,
			ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 1,
			Mode: 0, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: fmt.Sprintf("%s0", stream_arr[0][len(stream_arr[0])-5:])}
		_, err = sipapi.SipPlay(pm)
		if err != nil {
			Logger.Error("向摄像头发送信令请求实时标清流推流到zlm失败", zap.Any("ipcId", stream_arr[0]), zap.Error(err))
			return
		}

		hd_stream_id := stream_arr[0] + "_1"
		rtp_info2 := zlm_api.ZlmStartRtpServer(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, hd_stream_id, "rtp", 0)
		if rtp_info2.Code != 0 || rtp_info2.Port == 0 {
			Logger.Error("open rtp server fail", zap.Int("code", rtp_info2.Code))
			return
		}
		// 向摄像头发送信令请求推实时高清清流到zlm
		pm2 := &sipapi.Streams{ChannelID: stream_arr[0], StreamID: hd_stream_id,
			ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 1,
			Mode: 0, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: fmt.Sprintf("%s1", stream_arr[0][len(stream_arr[0])-5:])}
		_, err = sipapi.SipPlay(pm2)
		if err != nil {
			Logger.Error("向摄像头发送信令请求实时高清流推流到zlm失败", zap.Any("ipcId", stream_arr[0]), zap.Error(err))
			return
		}
	}

	c.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"close": true,
	})
}

func zlmStreamNoneReader(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZLMStreamNoneReaderData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}

	s_size := strings.Split(req.Stream, "_")
	if len(s_size) == 3 {
		c.JSON(http.StatusOK, map[string]any{
			"code":  0,
			"close": true,
		})
	} else {
		c.JSON(http.StatusOK, map[string]any{
			"code":  0,
			"close": false,
		})
	}

}
