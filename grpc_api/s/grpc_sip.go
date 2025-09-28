package grpc_server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_server_util"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/mq/kafka"
	pb "go-sip/signaling"
	sipapi "go-sip/sip"
	"go-sip/zlm_api"

	"sync"
	"time"

	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

var SipSrv *SipServer

type SipServer struct {
	pb.UnimplementedSipServiceServer
	clients   sync.Map // 使用 sync.Map 管理客户端连接
	StreamMap map[string]string
}

func GetSipServer() *SipServer {

	if SipSrv == nil {
		SipSrv = &SipServer{
			StreamMap: make(map[string]string),
			clients:   sync.Map{},
		}
	}
	return SipSrv
}

// 双向流处理
func (s *SipServer) StreamChannel(stream pb.SipService_StreamChannelServer) error {
	// 接收初始注册信息
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	reg := firstMsg.GetRegister()
	if reg == nil {
		return status.Error(codes.InvalidArgument, "需要先注册客户端")
	}

	// 记录客户端连接
	clientCtx := &ClientContext{
		ID:     reg.ClientId,
		Stream: stream,
	}

	// 设备和rk平台关联存入redis
	redis_util.HSetIfNotExist_2(redis.AI_MODEL_DEVICE_RK_PLATFORM_KEY, reg.ClientId, reg.DeviceType)

	// 获取门店id,并发送中控设备在线状态
	storeNo, err := redis_util.HGet_4(redis.IOT_DEVICE_STORE_KEY, reg.ClientId)
	if err != nil || storeNo == "" {
		Logger.Error("未找到关联的门店", zap.Error(err))
	} else {
		// 去除storeNo中的双引号
		storeNo = strings.ReplaceAll(storeNo, "\"", "")
		kafkaMsg := kafka.DeviceStateKafkaMsg{
			DeviceType:   "a",
			DeviceSerial: reg.ClientId,
			StoreNo:      storeNo,
			OnlineState:  1,
			Timestamp:    time.Now().UnixMilli(),
		}
		err = kafka.SendKafkaMessageByTopic(kafka.HTY_IOT_DEVICE_ONLINE_TOPIC, reg.ClientId, kafkaMsg)
		if err != nil {
			Logger.Error("发送kafka消息失败", zap.Any("kafkaMsg", kafkaMsg), zap.Error(err))
		}
	}

	redis_util.HSet_2(redis.DEVICE_SIP_KEY, reg.ClientId, m.SMConfig.SipID)

	s.clients.Store(reg.ClientId, clientCtx)
	defer s.clients.Delete(reg.ClientId)
	defer redis_util.HDel_2(redis.DEVICE_SIP_KEY, reg.ClientId)

	for {
		msg, err := stream.Recv()
		if err != nil {
			Logger.Error("stream recv failed", zap.Error(err))
			redis_util.Del_2(fmt.Sprintf(redis.DEVICE_STATUS_KEY, reg.ClientId))
			return err
		}
		if msg != nil {
			Logger.Info("收到客户端消息", zap.Any("client id", reg.ClientId), zap.Any("msg", msg))

			// 如果是响应
			if res := msg.GetResult(); res != nil {
				if chVal, ok := clientCtx.ResponseChans.Load(res.MsgID); ok {
					ch := chVal.(chan *pb.CommandResult)
					ch <- res
					close(ch)
					clientCtx.ResponseChans.Delete(res.MsgID)
				}
				continue
			}

			// 处理其他非响应消息
		}
	}

}

// 主动调用客户端方法
func (s *SipServer) ExecuteCommand(clientID string, cmd *pb.ServerCommand) (*pb.CommandResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	val, ok := s.clients.Load(clientID)
	if !ok {
		return nil, status.Error(codes.NotFound, "客户端未连接")
	}
	client := val.(*ClientContext)

	resultChan := make(chan *pb.CommandResult, 1)
	client.ResponseChans.Store(cmd.MsgID, resultChan)
	defer client.ResponseChans.Delete(cmd.MsgID)

	if err := client.Stream.Send(cmd); err != nil {
		return nil, err
	}

	select {
	case res := <-resultChan:
		return res, nil
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "等待响应超时")
	}
}

type ClientContext struct {
	ID            string
	Stream        pb.SipService_StreamChannelServer
	LastActive    time.Time
	ClientCtx     context.Context
	ResponseChans sync.Map // key: MsgID(string), value: chan *pb.CommandResult
}

func (s *SipServer) IpcEventReq(ctx context.Context, req *pb.IpcEventRequest) (*pb.IpcEventAck, error) {
	switch req.Event {
	case sipapi.NotifyMethodIpcActive:
		Logger.Info("设备活跃状态通知", zap.Any("device_id", req.ClientId))
	case sipapi.NotifyMethodIpcRegister:
		Logger.Info("添加新的摄像头", zap.Any("device_id", req.ClientId), zap.Any("ipc_id", req.IpcId))
		// ipc状态初始化
		if req.IpcId != "" {
			redis_util.Set_2(fmt.Sprintf(redis.IPC_STATUS_KEY, req.IpcId), "ON", time.Second*120)
		}
	case sipapi.NotifyMethodIpcChannelsActive:
		Logger.Info("收到通道活跃通知 ", zap.Any("device_id", req.ClientId), zap.Any("channel_id", req.IpcId))
		// 添加新的ipc信息保存到redis
		ipcInfo := toIpcInfo(req, m.SMConfig.SipID)
		if ipcInfo.IpcId != "" && !redis_util.HExists_2(redis.DEVICE_IPC_INFO_KEY, req.IpcId) {
			Logger.Debug("IpcEventReq set redis:", zap.Any("ipcInfo: ", ipcInfo))
			redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, req.IpcId, ipcInfo)
			redis_util.SAdd_2(fmt.Sprintf(redis.DEVICE_IPC_KEY, req.ClientId), req.IpcId)
		}
		// 设备状态更新
		redis_util.Set_2(fmt.Sprintf(redis.DEVICE_STATUS_KEY, req.ClientId), "online", time.Second*180)
		// ipc状态更新
		redis_util.Set_2(fmt.Sprintf(redis.IPC_STATUS_KEY, req.IpcId), "ON", time.Second*120)
		// ipcId对应deviceId存入redis
		redis_util.HSet_2(fmt.Sprintf(redis.SIP_IPC, m.SMConfig.SipID), req.IpcId, req.ClientId)
		// 获取门店id
		storeNo, err := redis_util.HGet_4(redis.IOT_DEVICE_STORE_KEY, req.ClientId)
		if err != nil || storeNo == "" {
			Logger.Error("未找到关联的门店", zap.Error(err))
		} else {
			// 去除storeNo中的双引号
			storeNo = strings.ReplaceAll(storeNo, "\"", "")
			kafkaMsg := kafka.DeviceStateKafkaMsg{
				DeviceType:   "9",
				DeviceSerial: req.IpcId,
				StoreNo:      storeNo,
				OnlineState:  1,
				Timestamp:    time.Now().UnixMilli(),
			}
			err = kafka.SendKafkaMessageByTopic(kafka.HTY_IOT_DEVICE_ONLINE_TOPIC, req.ClientId, kafkaMsg)
			if err != nil {
				Logger.Error("发送kafka消息失败", zap.Any("kafkaMsg", kafkaMsg), zap.Error(err))
			}
		}
	}
	return &pb.IpcEventAck{Success: true, Msg: "success"}, nil
}

func (s *SipServer) IpcInviteReq(ctx context.Context, req *pb.IpcInviteRequest) (*pb.IpcInviteAck, error) {

	zlm_id := s.StreamMap[req.IpcId]
	redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, zlm_id)
	if err != nil {
		return nil, err
	}
	// 反序列化 JSON 字符串
	var zlmInfo model.ZlmInfo
	err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
	if err != nil {
		return nil, err
	}
	Logger.Info("IpcInviteReq", zap.Any("zlmInfo", zlmInfo))
	rtp_info := zlm_api.ZlmStartSendRtpPassive(zlmInfo.ZlmDomain, zlmInfo.ZlmSecret, req.IpcId)
	return &pb.IpcInviteAck{Success: true, ZlmIp: zlmInfo.ZlmIp, ZlmPort: int64(rtp_info.LocalPort)}, nil
}

func (s *SipServer) AiEventReq(ctx context.Context, req *pb.AIEventRequest) (*pb.AIEventAck, error) {
	Logger.Info("收到AI事件请求", zap.Any("req", req))
	// 设备和rk平台关联存入redis
	redis_util.HSetIfNotExist_2(redis.AI_MODEL_DEVICE_RK_PLATFORM_KEY, req.DeviceId, req.RkPlatform)
	// 记录流id和ai模型关联关系到redis
	redis_util.HSetIfNotExist_2(redis.AI_MODEL_STREAM_CLASSNAME_KEY, req.StreamId, req.ClassName)
	// 获取门店id
	storeNo, err := redis_util.HGet_4(redis.IOT_DEVICE_STORE_KEY, req.DeviceId)
	if err != nil || storeNo == "" {
		Logger.Error("未找到关联的门店", zap.Error(err))
	} else {
		// 去除storeNo中的双引号
		storeNo = strings.ReplaceAll(storeNo, "\"", "")
		if req.ClassName == "person" {
			// 人形模型识别消息发送给kafka
			kafkaMsg := kafka.AiHumanEventKafkaMsg{
				DeviceSerial: req.DeviceId,
				StoreNo:      storeNo,
				AiEvent:      "1",
				HumanCount:   req.Count,
				Timestamp:    time.Now().UnixMilli(),
			}
			err = kafka.SendKafkaMessageByTopic(kafka.HTY_IOT_DOOR_STATE_TOPIC, req.DeviceId, kafkaMsg)
			if err != nil {
				Logger.Error("发送kafka消息失败", zap.Any("kafkaMsg", kafkaMsg), zap.Error(err))
			}
		}
	}
	return &pb.AIEventAck{Success: true, Msg: "success"}, nil
}

func toIpcInfo(req *pb.IpcEventRequest, sipId string) model.IpcInfo {
	// 新增ipcinfo
	return model.IpcInfo{
		IpcId:             req.IpcId,
		IpcIP:             req.IpcIp,
		IpcName:           req.IpcName,
		DeviceID:          req.ClientId,
		ChannelId:         req.ChannelId,
		Manufacturer:      req.Manufacturer,
		Transport:         req.Transport,
		StreamType:        req.Streamtype,
		Status:            req.Status,
		ActiveTime:        req.ActiveTime,
		SipId:             sipId,
		LastHeartbeatTime: time.Now().Unix(),
	}
}
