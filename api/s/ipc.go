package api

import (
	"encoding/json"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_server_util"
	"go-sip/grpc_api"
	grpc_server "go-sip/grpc_api/s"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/mq/kafka"
	pb "go-sip/signaling"
	"strings"

	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func IpcStatusSync() {
	ctx := context.Background()
	go func() {
		timer := time.NewTicker(time.Second * 30)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				Logger.Info("IpcStatusSync exit")
				return
			case <-timer.C:
				ok, _ := redis_util.SetNX(redis.IPC_STATUS_SYNC_LOCK_KEY, "ok", time.Second*10)
				if !ok {
					continue
				}
				Logger.Info("开始同步ipc状态")
				// 查询redis获取ipc列表
				device_ipc_info_map, err := redis_util.HGetAll_2(redis.DEVICE_IPC_INFO_KEY)
				if err != nil || device_ipc_info_map == nil || len(device_ipc_info_map) == 0 {
					Logger.Debug("未查询到ipc列表")
				} else {
					// 遍历device_ipc_info_map
					for _, v := range device_ipc_info_map {
						ipc_info := model.IpcInfo{}
						// 反序列化
						err := json.Unmarshal([]byte(v), &ipc_info)
						if err != nil {
							Logger.Error("json反序列化失败", zap.Error(err))
							continue
						}
						// 查询ipc状态
						ipcStatus, err := redis_util.Get_2(fmt.Sprintf(redis.IPC_STATUS_KEY, ipc_info.IpcId))
						if err != nil || ipcStatus == "" {
							ipc_info.Status = "OFFLINE"
						} else {
							ipc_info.Status = ipcStatus
						}
						if ipc_info.Status == "OFFLINE" {
							// 获取门店id
							storeNo, err := redis_util.HGet_4(redis.IOT_DEVICE_STORE_KEY, ipc_info.DeviceID)
							if err != nil || storeNo == "" {
								Logger.Error("未找到关联的门店", zap.Error(err))
							} else {
								// 去除storeNo中的双引号
								storeNo = strings.ReplaceAll(storeNo, "\"", "")
								kafkaMsg := kafka.DeviceStateKafkaMsg{
									DeviceType:   "9",
									DeviceSerial: ipc_info.IpcId,
									StoreNo:      storeNo,
									OnlineState:  0,
									Timestamp:    time.Now().UnixMilli(),
								}
								Logger.Info("发送ipc离线kafka消息", zap.Any("kafkaMsg", kafkaMsg))
								err = kafka.SendKafkaMessageByTopic(kafka.HTY_IOT_DEVICE_ONLINE_TOPIC, ipc_info.DeviceID, kafkaMsg)
								if err != nil {
									Logger.Error("发送ipc离线kafka消息失败", zap.Any("kafkaMsg", kafkaMsg), zap.Error(err))
								}
							}
							// 如果ipc_info.LastHeartbeatTime距离当前时间超过24小时, 则置为错误状态
							if time.Now().Unix()-ipc_info.LastHeartbeatTime > 60 { // 改成1分钟为了测试
								ipc_info.Status = "ERROR"
							}
						}
						redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, ipc_info.IpcId, ipc_info)
					}
				}
			}
		}
	}()

}

// @Summary 国标推流重置
// @Router /ipc/streamReset [get]
func IpcStreamReset(c *gin.Context) {
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
	Logger.Info("国标推流重置", zap.Any("data", data))
	result, err := sip_server.ExecuteCommand(device_id, &pb.ServerCommand{
		MsgID:   m.MsgID_IpcStreamReset,
		Method:  m.IpcStreamReset,
		Payload: d,
	})
	if err != nil {
		m.JsonResponse(c, m.StatusSysERR, "中控请求错误，请检查是否掉线")
		return
	}
	m.JsonResponse(c, m.StatusSucc, string(result.Payload))

}
