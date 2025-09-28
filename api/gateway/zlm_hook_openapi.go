package gateway

import (
	. "go-sip/common"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_gateway_util"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		hook
// @Description	zlm 启动，具体业务自行实现
func ZLMWebHook(c *gin.Context) {
	// 鉴权
	sign := c.Query("sign")
	if sign != "" {
		token := utils.GetMD5(sign)
		if token != utils.GetMD5(m.GatewayConfig.Sign) {
			m.ZlmWebHookResponse(c, 401, "Unauthorized")
			return
		}
	} else {
		m.ZlmWebHookResponse(c, 401, "Unauthorized")
		return
	}
	method := c.Param("method")
	if method == "" {
		m.ZlmWebHookResponse(c, -1, "method不能为空")
		return
	}
	// 获取参数列表
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "body error")
		return
	}
	paramsMap := getHookBodyParams(c, bodyBytes)
	if paramsMap == nil {
		m.ZlmWebHookResponse(c, -1, "参数错误")
	}
	app := c.Param("app")
	Logger.Info("gateway ZLMWebHook params", zap.Any("method", method), zap.Any("app", app), zap.Any("paramsMap", paramsMap))

	// 获取sip服务host
	device_id := paramsMap["device_id"]
	stream_id := paramsMap["stream"]
	// 合屏的话，流id就是设备id
	sub_ipc := paramsMap["sub_ipc"]
	if sub_ipc != "" {
		device_id = stream_id
	}
	sipServerId := getSipServerId(app, device_id, stream_id)
	sipServerHost, _, err := getSipServerHost(sipServerId)
	if err != nil {
		Logger.Error("获取SIP服务器失败", zap.Error(err))
		m.ZlmWebHookResponse(c, -1, "获取SIP服务器失败")
		return
	}

	// 调用sip接口
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s%s", sipServerHost, ZLMWebHookBaseURL+"/"+method), io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		m.ZlmWebHookResponse(c, -1, "json marshal error")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		m.ZlmWebHookResponse(c, -1, "调用sip hook接口失败")
		return
	}
	defer resp.Body.Close()
	switch method {
	case "on_http_access":
		c.JSON(http.StatusOK, map[string]any{
			"code":   0,
			"second": 86400})
	case "on_play":
		var zlmStreamOnPlayData = model.ZlmStreamOnPlayData{}
		if err := utils.JSONDecode(bodyBytes, &zlmStreamOnPlayData); err != nil {
			m.ZlmWebHookResponse(c, -1, "json marshal error")
			return
		}
		m.ZlmWebHookResponse(c, 0, "success")
	case "on_publish":
		// 推流鉴权
		c.JSON(http.StatusOK, map[string]any{
			"code":       0,
			"enableHls":  1,
			"enableMP4":  false,
			"enableRtxp": 1,
			"msg":        "success",
		})
	case "on_stream_not_found":
		time.Sleep(time.Second * 1)
		c.JSON(http.StatusOK, map[string]any{
			"code":  0,
			"close": true,
		})
	case "on_stream_none_reader":
		c.JSON(http.StatusOK, map[string]any{
			"code":  0,
			"close": true,
		})
	case "on_stream_changed":
		// 流注册和注销通知
		var zLMStreamChangedData = model.ZLMStreamChangedData{}
		if err := utils.JSONDecode(bodyBytes, &zLMStreamChangedData); err != nil {
			m.ZlmWebHookResponse(c, -1, "json marshal error")
			return
		}
		m.ZlmWebHookResponse(c, 0, "success")
	default:
		m.ZlmWebHookResponse(c, 0, "success")
	}

}

// 根据hook方法获取body参数列表
func getHookBodyParams(c *gin.Context, bodyBytes []byte) map[string]string {
	paramsMap := make(map[string]string)
	paramsArray := []string{}
	method := c.Param("method")
	switch method {
	case "on_server_started":
		// zlm 启动
	case "on_http_access":
		// http请求鉴权
	case "on_play":
		var zlmStreamOnPlayData = model.ZlmStreamOnPlayData{}
		if err := utils.JSONDecode(bodyBytes, &zlmStreamOnPlayData); err != nil {
			Logger.Warn("GetHookBodyParams body error")
			return nil
		}
		paramsArray = strings.Split(zlmStreamOnPlayData.Params, "&")
	case "on_publish":
		// 推流业务
		var zlmStreamPublishData = model.ZlmStreamPublishData{}
		if err := utils.JSONDecode(bodyBytes, &zlmStreamPublishData); err != nil {
			Logger.Warn("GetHookBodyParams body error")
			return nil
		}
		paramsArray = strings.Split(zlmStreamPublishData.Params, "&")
	case "on_stream_none_reader":
		// 无人观看视频流业务
	case "on_stream_not_found":
		// 请求播放时，流不存在时触发
		var zLMStreamNotFoundData = model.ZLMStreamNotFoundData{}
		if err := utils.JSONDecode(bodyBytes, &zLMStreamNotFoundData); err != nil {
			Logger.Warn("GetHookBodyParams body error")
			return nil
		}
		paramsArray = strings.Split(zLMStreamNotFoundData.Params, "&")
	case "on_record_mp4":
		//  mp4 录制完成
	case "on_stream_changed":
		// 流注册和注销通知
		var zLMStreamChangedData = model.ZLMStreamChangedData{}
		if err := utils.JSONDecode(bodyBytes, &zLMStreamChangedData); err != nil {
			Logger.Warn("GetHookBodyParams body error")
			return nil
		}
		paramsArray = strings.Split(zLMStreamChangedData.Params, "&")
	default:
		break
	}
	if len(paramsArray) != 0 {
		for _, params := range paramsArray {
			if params == "" {
				continue
			}
			param := strings.Split(params, "=")
			if len(param) != 2 {
				Logger.Warn("参数格式错误")
				return nil
			}
			paramsMap[param[0]] = param[1]
		}
	}
	return paramsMap
}

func getSipServerId(app, device_id, stream_id string) string {
	var sip_server_id string
	if device_id != "" {
		redis_sip_server_id, err := redis_util.HGet_2(redis.DEVICE_SIP_KEY, device_id)
		if err != nil {
			Logger.Error("没有关联的sip服务", zap.Any("redis.DEVICE_SIP_KEY", redis.DEVICE_SIP_KEY), zap.Any("device_id", device_id))
		}
		sip_server_id = redis_sip_server_id
	} else if stream_id != "" {
		if app == "rtp" {
			stream_id_arr := strings.Split(stream_id, "_")
			// 查询redis获取ipc列表
			device_ipc_info_str, err := redis_util.HGet_2(redis.DEVICE_IPC_INFO_KEY, stream_id_arr[0])
			if err != nil || device_ipc_info_str == "" {
				Logger.Error("未找到任何ipc", zap.Error(err))
			} else {
				ipc_info := model.IpcInfo{}
				// 反序列化
				err = json.Unmarshal([]byte(device_ipc_info_str), &ipc_info)
				if err != nil {
					Logger.Error("json反序列化失败", zap.Error(err))
				} else {
					sip_server_id = ipc_info.SipId
				}
			}
		}
	}
	return sip_server_id
}

func getSipServerHost(sipServerId string) (string, string, error) {
	var sip_server_host string
	var sip_server_id string
	if sipServerId != "" {
		redis_sip_server_host, err := redis_util.HGet_2(redis.SIP_SERVER_HOST, sipServerId)
		if err != nil {
			return "", "", fmt.Errorf("未找到任何sip服务")
		}
		sip_server_host = redis_sip_server_host
	}

	if sip_server_host == "" {
		// 查出所有sip服务, hash取模获取一个
		sip_server_map, err := redis_util.HGetAll_2(redis.SIP_SERVER_HOST)
		if err != nil || sip_server_map == nil || len(sip_server_map) == 0 {
			return "", "", fmt.Errorf("未找到任何sip服务")
		}

		selectPollSipServerId, selectPollSipServerHost, err := SelectPollMapValue("gateway_invoke_server", FilterLoopbackSipServers(sip_server_map))
		if err != nil || selectPollSipServerHost == "" {
			return "", "", fmt.Errorf("未找到任何sip服务")
		}
		sip_server_host = selectPollSipServerHost
		sip_server_id = selectPollSipServerId
		return selectPollSipServerHost, selectPollSipServerId, nil
	}
	return sip_server_host, sip_server_id, nil
}
