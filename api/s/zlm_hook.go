package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_server_util"
	"go-sip/grpc_api"
	grpc_server "go-sip/grpc_api/s"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	pb "go-sip/signaling"
	"go-sip/utils"
	"go-sip/zlm_api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var audio_pull_map = make(map[string]string) // audio 板端推流是否成功注册
var ZLM_Node = make(map[string]model.ZlmInfo)

var stream_hd = make(map[string]int)

var StreamWait = make(map[string]chan struct{})

func ZLMWebHook(c *gin.Context) {
	method := c.Param("method")
	Logger.Info("server sip ZLMWebHook", zap.String("method", method))
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
	case "on_rtp_server_timeout":
		// rtp server 长时间未收到数据
		zlmRtpServerTimeout(c)
	default:
		m.ZlmWebHookResponse(c, 0, "success")
	}

}

func zlmServerStart(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}

	var req model.ZlmServerStartDate
	if err := utils.JSONDecode(data, &req); err != nil {
		fmt.Println("JSON解析失败：", err)
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	// 获取redis数据, 没有则存入zlm返回的信息到redis
	zlmInfoRedisStr, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerId)
	if err != nil || zlmInfoRedisStr == "" {
		var zlmRedisDTO model.ZlmInfo
		zlmRedisDTO.ZlmDomain = req.MediaServerId
		zlmRedisDTO.ZlmIp = req.ExternIP
		zlmRedisDTO.ZlmSecret = req.APISecret
		zlmRedisDTO.ZlmPort = req.HTTPPort
		err = redis_util.HSetStruct_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerId, zlmRedisDTO)
		if err != nil {
			m.ZlmWebHookResponse(c, -1, "redis error")
			return
		}
	}

	m.ZlmWebHookResponse(c, 0, "zlm server start success")
}

// zlm点播业务
func zlmStreamOnPlay(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZlmStreamOnPlayData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "参数格式错误，json反序列化失败")
		return
	}
	Logger.Info("server sip zlmStreamOnPlay", zap.Any("req", req))
	// 获取参数列表
	// paramsMap := make(map[string]string)
	// paramsArray := strings.Split(req.Params, "&")

	// for _, params := range paramsArray {
	// 	param := strings.Split(params, "=")
	// 	if len(param) != 2 {
	// 		m.ZlmWebHookResponse(c, 401, "Unauthorized")
	// 		return
	// 	}
	// 	paramsMap[param[0]] = param[1]
	// }
	// redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerID)
	// if err != nil {
	// 	m.ZlmWebHookResponse(c, -1, "redis error")
	// 	return
	// }

	// // 反序列化 JSON 字符串
	// var zlmInfo model.ZlmInfo
	// err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
	// if err != nil {
	// 	m.ZlmWebHookResponse(c, -1, "参数格式错误，json反序列化失败")
	// 	return
	// }

	var stream_id_arr []string
	if req.Stream != "" {
		stream_id_arr = strings.Split(req.Stream, "_")
	} else {
		m.ZlmWebHookResponse(c, -1, "参数格式错误")
		return
	}

	// 实时流
	if len(stream_id_arr) == 1 || len(stream_id_arr) == 2 {
		Logger.Info("ZlmStreamOnPlay 实时流", zap.Any("req.Stream", req.Stream))
	} else {
		m.ZlmWebHookResponse(c, -1, "stream参数格式错误")
		return
	}
	//视频播放触发鉴权
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
	paramsMap := make(map[string]string)
	paramsArray := strings.Split(req.Params, "&")
	for _, params := range paramsArray {
		param := strings.Split(params, "=")
		if len(param) != 2 {
			m.ZlmWebHookResponse(c, 401, "Unauthorized")
			return
		}
		paramsMap[param[0]] = param[1]
	}
	if req.Regist {
		Logger.Info("流注册 ", zap.Any("req", req))
		if req.APP == "audio" && req.Schema == "rtsp" {
			if device_id, ok := paramsMap["device_id"]; ok {
				if zlm_node, ok := ZLM_Node[device_id]; ok {
					sip_req := &grpc_api.Sip_Audio_Play_Req{
						DeviceID: device_id,
						ZLMIP:    zlm_node.ZlmIp,
						ZLMPort:  zlm_node.ZlmPort,
						StreamID: req.Stream,
						Token:    m.SMConfig.Sign,
					}
					d, err := json.Marshal(sip_req)
					if err != nil {
						m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
						return
					}
					sip_id, err := redis_util.HGet_2(redis.DEVICE_SIP_KEY, device_id)
					if err != nil || sip_id == "" {
						m.ZlmWebHookResponse(c, -1, "sip_id未找到， 或者不在该sip服务")
						return
					}
					go func() {
						sip_server := grpc_server.GetSipServer()
						_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
							MsgID:   m.MsgID_Device_Play_Audio,
							Method:  m.PlayAudio,
							Payload: d,
						})
						if err != nil {
							Logger.Error("执行远程播放命令失败", zap.Any("device_id", device_id), zap.Error(err))
							return
						}
					}()
				}
			} else {
				audio_pull_map[req.Stream] = device_id
			}
		}
		if req.Schema == "rtsp" { // 如果是国标流注册
			if st, ok := StreamWait[req.Stream]; ok {
				st <- struct{}{}
			}
		}
	} else {
		Logger.Info("流注销 :", zap.Any("req", req))
		if req.APP == "rtp" && req.Schema == "rtsp" {
			s_size := strings.Split(req.Stream, "_")

			redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerId)
			if err != nil {
				m.ZlmWebHookResponse(c, -1, "查询redis错误")
				return
			}

			// 反序列化 JSON 字符串
			var zlmInfo model.ZlmInfo
			err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
			if err != nil {
				m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
				return
			}

			if device_id, ok := paramsMap["device_id"]; ok {
				sip_server := grpc_server.GetSipServer()
				sip_req := &grpc_api.Sip_Stop_Play_Req{
					App:         req.APP,
					DeviceID:    device_id,
					StreamID:    req.Stream,
					ZlmIP:       zlmInfo.ZlmIp,
					ZlmDomain:   zlmInfo.ZlmDomain,
					ZlmHttpPort: zlmInfo.ZlmPort,
					ZlmSecret:   zlmInfo.ZlmSecret,
				}
				d, err := json.Marshal(sip_req)
				if err != nil {
					m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
					return
				}
				device_id, err := redis_util.HGet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), s_size[0])
				if err != nil || device_id == "" {
					m.ZlmWebHookResponse(c, -1, "ipc_id未注册，请检查摄像头是否正常")
					return
				}
				_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
					MsgID:   m.MsgID_StopPlay,
					Method:  m.StopPlay,
					Payload: d,
				})
				if err != nil {
					m.ZlmWebHookResponse(c, -1, "终端请求错误，请检查是否掉线")
					return
				}
			}

		}
		if req.APP == "audio" && req.Schema == "rtsp" {
			delete(audio_pull_map, req.Stream)
			delete(ZLM_Node, req.Stream)
			delete(stream_hd, req.Stream)
		}
	}
	m.ZlmWebHookResponse(c, 0, "success")
}

func zlmStreamPublish(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZlmStreamPublishData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}

	// 获取参数列表
	paramsMap := make(map[string]string)
	paramsArray := strings.Split(req.Params, "&")

	if req.Params != "" {
		for _, params := range paramsArray {
			param := strings.Split(params, "=")
			if len(param) != 2 {
				m.ZlmWebHookResponse(c, 401, "Unauthorized")
				return
			}
			paramsMap[param[0]] = param[1]
		}
	}

	if req.App == "audio" {
		if tp, ok := paramsMap["type"]; ok {
			if tp == "push" { // 通知板端拉流
				if device_id, ok := paramsMap["device_id"]; ok {
					redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerID)
					if err != nil {
						m.ZlmWebHookResponse(c, -1, "redis error")
						return
					}

					// 反序列化 JSON 字符串
					var zlmInfo model.ZlmInfo
					err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
					if err != nil {
						m.ZlmWebHookResponse(c, -1, "参数格式错误，json反序列化失败")
						return
					}

					ZLM_Node[device_id] = zlmInfo
				}
			}
		}
	} else if req.App == "rtp" { // 国标推流 不用鉴权
		Logger.Info("rtp推流, 不用鉴权")
	} else if req.App == "broadcast" { // 摄像头广播推流
		sip_req := &grpc_api.Sip_Ipc_BroadCast_Req{
			ChannelID: req.Stream,
		}
		d, err := json.Marshal(sip_req)
		if err != nil {
			m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
		}
		device_id, err := redis_util.HGet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), req.Stream)
		if err != nil || device_id == "" {
			m.ZlmWebHookResponse(c, -1, "ipc_id未注册，请检查摄像头是否正常")
			return
		}

		sip_server := grpc_server.GetSipServer()

		sip_server.StreamMap[req.Stream] = req.MediaServerID
		go func() {

			_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
				MsgID:   m.MsgID_BroadCast,
				Method:  m.Broadcast,
				Payload: d,
			})
			if err != nil {
				return
			}
		}()

	} else {

		if sign, ok := paramsMap["sign"]; ok {
			token := utils.GetMD5(sign)
			if token != utils.GetMD5(m.SMConfig.Sign) {
				m.ZlmWebHookResponse(c, 401, "Unauthorized")
				return
			}
		} else {
			m.ZlmWebHookResponse(c, 401, "Unauthorized")
			return
		}
	}

	c.JSON(http.StatusOK, map[string]any{
		"code":         0,
		"enable_audio": true,
		"enable_MP4":   false,
		"msg":          "success",
	})
}

func zlmRtpServerTimeout(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZlmRtpServerTimeoutData{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	Logger.Info("=== rtp服务超时", zap.Any("req", req))
}

func zlmRecordMp4(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	var req = model.ZLMRecordMp4Data{}
	if err := utils.JSONDecode(data, &req); err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}

	m.ZlmWebHookResponse(c, 0, "success")
}

// mode 0 UDP 1 Tcp被动
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
	Logger.Info("server sip zlmStreamNotFound", zap.Any("req", req))

	// 获取参数列表
	paramsMap := make(map[string]string)
	paramsArray := strings.Split(req.Params, "&")

	for _, params := range paramsArray {
		param := strings.Split(params, "=")
		if len(param) != 2 {
			m.ZlmWebHookResponse(c, -1, "传参格式错误")
			return
		}
		paramsMap[param[0]] = param[1]
	}
	redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerID)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "查询redis错误")
		return
	}

	// 反序列化 JSON 字符串
	var zlmInfo model.ZlmInfo
	err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
		return
	}
	sip_server := grpc_server.GetSipServer()

	var stream_id_arr []string
	if req.Stream != "" {
		stream_id_arr = strings.Split(req.Stream, "_")
	} else {
		m.ZlmWebHookResponse(c, -1, "参数格式错误")
		return
	}

	if req.APP == "rtp" || req.APP == "live" {
		mode, err := strconv.Atoi(paramsMap["mode"]) // 返回 (int, error)
		if err != nil || mode < 0 || mode > 1 {
			m.ZlmWebHookResponse(c, -1, "参数格式错误，mode参数错误")
			return
		}
		var device_id string

		// 请求zlm进行合屏
		if sub_ipc, ok := paramsMap["sub_ipc"]; ok {
			device_id = stream_id_arr[0]
			// 对sub_ipc进行分割
			sub_stream_id_arr := strings.Split(sub_ipc, "_")
			stream_id_list := make([]string, len(sub_stream_id_arr))
			// 点播所有拼接的摄像头
			for i, ipc_id := range sub_stream_id_arr {
				// 默认使用标清流
				stream_id := fmt.Sprintf("%s_0", ipc_id)
				stream_id_list[i] = stream_id
				rtpinfo := zlm_api.ZlmGetMediaInfo(zlmInfo.ZlmDomain, zlmInfo.ZlmSecret, stream_id)
				if rtpinfo.Code == 0 && !rtpinfo.Exist {
					rtp_info := zlm_api.ZlmStartRtpServer("http://"+zlmInfo.ZlmIp+":"+zlmInfo.ZlmPort, zlmInfo.ZlmSecret, stream_id, req.APP, mode)
					sip_req := &grpc_api.Sip_Play_Req{
						DeviceID:    device_id,
						ChannelID:   stream_id,
						ZLMIP:       zlmInfo.ZlmIp,
						ZlmDomain:   zlmInfo.ZlmDomain,
						ZlmSecret:   zlmInfo.ZlmSecret,
						ZlmHttpPort: zlmInfo.ZlmPort,
						ZLMPort:     rtp_info.Port,
						Resolution:  0, // 废弃
						Mode:        mode,
						App:         req.APP,
					}
					d, err := json.Marshal(sip_req)
					if err != nil {
						Logger.Error("json序列化失败")
						break
					}
					_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
						MsgID:   m.MsgID_Play,
						Method:  m.Play,
						Payload: d,
					})
					if err != nil {
						Logger.Error("ipc点播失败", zap.Any("ipcId", ipc_id))
					}
				}
			}

			// 进行合屏
			dto := model.StreamMergeInfoDTO{
				DeviceId:  device_id,
				IpcIdList: stream_id_list,
				StreamId:  device_id, // 默认使用设备ID
				Type:      1,
			}
			resp := zlm_api.ZlmMergeStream(dto, &zlmInfo)
			if resp.Code != 0 {
				m.ZlmWebHookResponse(c, -1, "参数格式错误，合屏失败")
				return
			}
			redis_util.HSet_2(redis.MERGE_VIDEO_STREAM_IPC_LIST_KEY, device_id, sub_ipc)
		} else { // 非合屏
			ipc_id := stream_id_arr[0]
			// 如果ipc_id以IPC开头，则表示为非国标摄像头
			if !strings.HasPrefix(ipc_id, "IPC") {
				device_id, err = redis_util.HGet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), ipc_id)
				if err != nil || device_id == "" {
					m.ZlmWebHookResponse(c, -1, "ipc_id未注册，请检查摄像头是否正常")
					return
				}
			} else {
				device_id, err = redis_util.HGet_2(redis.NOT_GB_IPC_DEVICE, ipc_id)
				if err != nil || device_id == "" {
					m.ZlmWebHookResponse(c, -1, "ipc_id未注册，请检查摄像头是否正常")
					return
				}
			}
			rtp_info := zlm_api.ZlmStartRtpServer(zlmInfo.ZlmDomain, zlmInfo.ZlmSecret, req.Stream, req.APP, mode)

			// rtp实时流
			if len(stream_id_arr) == 1 || len(stream_id_arr) == 2 {
				// 点播ipc
				sip_req := &grpc_api.Sip_Play_Req{
					DeviceID:    device_id,
					ChannelID:   req.Stream,
					ZLMIP:       zlmInfo.ZlmIp,
					ZlmDomain:   zlmInfo.ZlmDomain,
					ZlmSecret:   zlmInfo.ZlmSecret,
					ZlmHttpPort: zlmInfo.ZlmPort,
					ZLMPort:     rtp_info.Port,
					Resolution:  0, // 废弃
					Mode:        mode,
					App:         req.APP,
				}
				d, err := json.Marshal(sip_req)
				if err != nil {
					Logger.Error("ipc的json序列化失败")
				} else {
					_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
						MsgID:   m.MsgID_Play,
						Method:  m.Play,
						Payload: d,
					})
					if err != nil {
						Logger.Error("ipc点播失败", zap.Any("stream_id", req.Stream))
					}
				}

			} else {
				m.ZlmWebHookResponse(c, -1, "stream参数格式错误")
				return
			}
		}

	} else if req.APP == "audio" { // 通知板端推流
		if _, ok := audio_pull_map[req.Stream]; !ok { // 如果该流已注册 ，则不需要在通知板端推流
			if tp, ok := paramsMap["type"]; ok {
				if tp == "play" {
					if device_id, ok := paramsMap["device_id"]; ok {
						if sign, ok := paramsMap["sign"]; ok {
							token := utils.GetMD5(sign)
							if token != utils.GetMD5(m.SMConfig.Sign) {
								m.ZlmWebHookResponse(c, 401, "参数格式错误，sign错误")
								return
							}
						} else {
							m.ZlmWebHookResponse(c, 401, "参数格式错误，sign错误")
							return
						}

						sip_req := &grpc_api.Sip_Audio_Push_Req{
							DeviceID: device_id,
							ZLMIP:    zlmInfo.ZlmIp,
							ZLMPort:  zlmInfo.ZlmPort,
							StreamID: req.Stream,
							Token:    m.SMConfig.Sign,
						}
						d, err := json.Marshal(sip_req)
						if err != nil {
							m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
							return
						}
						sip_id, err := redis_util.HGet_2(redis.DEVICE_SIP_KEY, device_id)
						if err != nil || sip_id == "" {
							m.ZlmWebHookResponse(c, -1, "sip_id未找到， 或者不在该sip服务")
							return
						}

						_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
							MsgID:   m.MsgID_Device_Push_Audio,
							Method:  m.PushAudio,
							Payload: d,
						})
						if err != nil {
							Logger.Error("执行远程推送命令失败", zap.Any("device_id", device_id), zap.Error(err))
							m.ZlmWebHookResponse(c, -1, "执行远程推送命令失败")
							return
						}

					}
				}
			}
		}
	}
	StreamWait[req.Stream] = make(chan struct{})

	tick := time.NewTicker(5 * time.Second)
	select {
	case <-StreamWait[req.Stream]:
		close(StreamWait[req.Stream])
		delete(StreamWait, req.Stream)
		break
	case <-tick.C:
		Logger.Warn("等待流超时", zap.Any("stream", req.Stream))
		break
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

	if req.APP == "rtp" {
		redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.MediaServerID)
		if err != nil {
			m.ZlmWebHookResponse(c, -1, "查询redis错误")
			return
		}

		// 反序列化 JSON 字符串
		var zlmInfo model.ZlmInfo
		err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
		if err != nil {
			m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
			return
		}
		s_size := strings.Split(req.Stream, "_")
		if len(s_size) == 3 {
			req.Stream = s_size[0]
		}
		// sip_server := grpc_server.GetSipServer()
		// sip_req := &grpc_api.Sip_Stop_Play_Req{
		// 	App:         req.APP,
		// 	StreamID:    req.Stream,
		// 	ZlmIP:       zlmInfo.ZlmIp,
		// 	ZlmDomain:   zlmInfo.ZlmDomain,
		// 	ZlmHttpPort: zlmInfo.ZlmPort,
		// 	ZlmSecret:   zlmInfo.ZlmSecret,
		// }
		// d, err := json.Marshal(sip_req)
		// if err != nil {
		// 	m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
		// 	return
		// }
		// device_id, err := redis_util.HGet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), req.Stream)
		// if err != nil {
		// 	m.ZlmWebHookResponse(c, -1, "ipc_id未注册，请检查摄像头是否正常")
		// 	return
		// }
		// _, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
		// 	MsgID:   m.MsgID_StopPlay,
		// 	Method:  m.StopPlay,
		// 	Payload: d,
		// })
		// if err != nil {
		// 	m.ZlmWebHookResponse(c, -1, "终端请求错误，请检查是否掉线")
		// 	return
		// }
		// closeRtpServer 服务端
		closeRtpRsp := zlm_api.ZlmCloseRtpServer("http://"+zlmInfo.ZlmIp+":"+zlmInfo.ZlmPort, zlmInfo.ZlmSecret, req.Stream)
		if closeRtpRsp.Code != 0 {
			m.ZlmWebHookResponse(c, -1, "关闭RtpServer失败")
			return
		}
		if closeRtpRsp.Hit >= 1 {
			m.ZlmWebHookResponse(c, 0, "成功命中并关闭RtpServer")
		} else {
			m.ZlmWebHookResponse(c, 0, "未命中RtpServer")
		}

	} else if req.APP == "audio" {

		sip_server := grpc_server.GetSipServer()
		sip_req := &grpc_api.Sip_Close_Audio_Req{}
		d, err := json.Marshal(sip_req)
		if err != nil {
			m.ZlmWebHookResponse(c, -1, "参数格式错误，json序列化失败")
		}

		if device_id, ok := audio_pull_map[req.Stream]; ok {

			_, err = sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
				MsgID:   m.MsgID_Device_Close_Audio,
				Method:  m.CloseAudio,
				Payload: d,
			})
			if err != nil {
				m.ZlmWebHookResponse(c, -1, "终端请求错误，请检查是否掉线")
			}

		}

	}
	c.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"close": true,
	})
}
