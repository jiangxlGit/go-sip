package api

import (
	"encoding/json"
	"fmt"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_server_util"
	"go-sip/grpc_api"
	grpc_server "go-sip/grpc_api/s"
	. "go-sip/logger"
	"go-sip/m"
	pb "go-sip/signaling"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 非国标推流重置
// @Router /ipc/nogbPushStreamReset [get]
func IpcNoGbPushStreamReset(c *gin.Context) {
	device_id := c.Query("device_id")
	ipc_id := c.Query("ipc_id")
	if device_id == "" {
		m.JsonResponse(c, m.StatusParamsERR, "参数错误")
		return
	}
	sip_server := grpc_server.GetSipServer()
	data := &grpc_api.Sip_Ipc_Push_Stream_Req{
		DeviceID: device_id,
		IpcId:    ipc_id,
	}
	d, err := json.Marshal(data)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "参数格式错误，json序列化失败")
		return
	}

	_, err = redis_util.HGet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), ipc_id)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, "ipc_id未注册，请检查摄像头是否正常")
		return
	}
	Logger.Info("非国标推流重置", zap.Any("data", data))
	result, err := sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
		MsgID:   m.MsgID_IpcPushStreamReset,
		Method:  m.IpcPushStreamReset,
		Payload: d,
	})
	if err != nil {
		m.JsonResponse(c, m.StatusSysERR, "中控请求错误，请检查是否掉线")
		return
	}
	m.JsonResponse(c, m.StatusSucc, string(result.Payload))

}
