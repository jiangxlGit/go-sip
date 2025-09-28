package gateway

import (
	"encoding/json"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_gateway_util"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"

	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		获取zlm服务信息接口
// @Router			/open/zlm/info [post]
func ZLMInfo(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("read body error", zap.Error(err))
		model.JsonResponseSysERR(c, "read body error")
		return
	}
	req := &ZLMInfoBody{}
	if err := utils.JSONDecode(bodyBytes, &req); err != nil {
		Logger.Error("bind json error", zap.Error(err))
		model.JsonResponseSysERR(c, "invalid request")
		return
	}
	ipc_id := req.IPCId
	device_id := req.DeviceId
	// 不能全部为空
	if ipc_id == "" && device_id == "" {
		Logger.Error("ipc_id和device_id不能同时为空")
		model.JsonResponseSysERR(c, "ipc_id和device_id不能同时为空")
		return
	}
	var deviceId = ""
	if device_id != "" {
		deviceId = device_id
	} else {
		// 根据ipc_id查询device_id
		device_ipc_info_str, err := redis_util.HGet_2(redis.DEVICE_IPC_INFO_KEY, ipc_id)
		if err != nil || device_ipc_info_str == "" {
			Logger.Error("根据ipc_id查询device_id失败", zap.Error(err))
			model.JsonResponseSysERR(c, "ipc_id查询device_id失败")
			return
		}

		ipcInfo := model.IpcInfo{}
		// 反序列化
		err = json.Unmarshal([]byte(device_ipc_info_str), &ipcInfo)
		if err != nil || ipcInfo.DeviceID == "" {
			Logger.Error("根据ipc_id查询device_id失败", zap.Error(err))
			model.JsonResponseSysERR(c, "ipc_id查询device_id失败")
			return
		}
		deviceId = ipcInfo.DeviceID
	}
	zlm_info, err := GatewayGetZlmInfo(deviceId)
	if err != nil || zlm_info == nil {
		Logger.Error("根据ipc_id查询zlmInfo失败", zap.Error(err))
		model.JsonResponseSysERR(c, "根据ipc_id查询zlmInfo失败")
		return
	}

	// 返回zlm_info
	model.JsonResponseSucc(c, zlm_info)
}

type ZLMInfoBody struct {
	IPCId    string `json:"ipcId"`
	DeviceId string `json:"deviceId"`
}
