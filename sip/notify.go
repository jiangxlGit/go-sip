package sip

import (
	db "go-sip/db/sqlite"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/signaling"
	"go-sip/zlm_api"

	"go.uber.org/zap"
)

const (
	// NotifyMethodUserActive ipc活跃状态通知
	NotifyMethodIpcActive = "devices.active"
	// NotifyMethodUserRegister ipc注册通知
	NotifyMethodIpcRegister = "devices.regiester"
	// NotifyMethodDeviceActive ipc通道活跃通知
	NotifyMethodIpcChannelsActive = "channels.active"
	// INVITE_METHOD 通知方法
	NotifyMethodInvite = "invite"

	Local_ZLM_Host = "http://127.0.0.1:9092"
)

var NotifyFunc func(notifyData *Notify, msg_type string) (*signaling.IpcEventAck, error)
var InviteFunc func(ipc_id string) (*signaling.IpcInviteAck, error)
var NotifyAiEventFunc func(device_id, rk_platform, stream_id string, event_id int64, class_name string, max_score float64, count int64) (*signaling.AIEventAck, error)
var NotifyOTAUpgradeFunc func(device_id, firmware_id, firmware_version, upgrade_complete, upgrade_progress, upgrade_error string) (*signaling.OTAUpgradeAck, error)

// Notify 消息通知结构
type Notify struct {
	Method       string `json:"method"`
	DeviceID     string `json:"data"`
	DeviceType   string `json:"device_type"`
	IpcId        string `json:"ipc_id"`
	IpcIP        string `json:"ipc_ip"`
	IpcName      string `json:"ipc_name"`
	ChannelId    string `json:"channel_id"`
	Manufacturer string `json:"manufacturer"`
	Transport    string `json:"transport"`
	StreamType   string `json:"stream_type"`
	Status       string `json:"status"`
	ActiveTime   int64  `json:"active_time"`
}

func notify(notifyData *Notify) {

	defer func() {
		if r := recover(); r != nil {
			Logger.Error("notify panic recovered", zap.Any("error", r))
		}
	}()

	if notifyData == nil {
		Logger.Error("notify received nil data")
		return
	}

	var ack *signaling.IpcEventAck
	var err error

	switch notifyData.Method {
	case NotifyMethodIpcActive: // ipc活跃通知
		ack, err = NotifyFunc(notifyData, notifyData.Method)
		if err != nil {
			Logger.Error("设备活跃通知服务端错误", zap.Any("data.DeviceID", notifyData.DeviceID), zap.Error(err))
		}
		if ack == nil {
			Logger.Error("通知服务端返回ack为nil", zap.Any("data.DeviceID", notifyData.DeviceID), zap.String("method", notifyData.Method))
			return
		}
	case NotifyMethodIpcRegister: // ipc注册通知
		ack, err = NotifyFunc(notifyData, notifyData.Method)
		if err != nil {
			Logger.Error("设备注册通知服务端错误", zap.Any("data.DeviceID", notifyData.DeviceID), zap.Error(err))
		}
		if ack == nil {
			Logger.Error("通知服务端返回ack为nil", zap.Any("data.DeviceID", notifyData.DeviceID), zap.String("method", notifyData.Method))
			return
		}
	case NotifyMethodIpcChannelsActive: // ipc通道活跃通知
		ack, err = NotifyFunc(notifyData, notifyData.Method)
		if err != nil {
			Logger.Error("通道活跃通知服务端错误", zap.Any("data.DeviceID", notifyData.DeviceID), zap.Error(err))
		}
		if ack == nil {
			Logger.Error("通知服务端返回ack为nil", zap.Any("data.DeviceID", notifyData.DeviceID), zap.String("method", notifyData.Method))
			return
		}
		var req = zlm_api.ZlmGetRtpInfoReq{}
		req.App = "rtp"
		req.Vhost = "__defaultVhost__"
		// 推送标清流到本地zlm
		req.StreamID = notifyData.DeviceID + "_0"
		resp := zlm_api.ZlmGetRtpInfo(Local_ZLM_Host, m.CMConfig.ZlmSecret, req)
		if resp.Code == 0 && !resp.Exist {

			SipStopPlay(req.StreamID)

			Logger.Info("本地ZLM不存在流，开始推送本地标清流", zap.Any("deviceID", notifyData.DeviceID))
			// 不存在，调用openRtpServer让本地zlm打开一个收流端口
			rtp_info := zlm_api.ZlmStartRtpServer(Local_ZLM_Host, m.CMConfig.ZlmSecret, req.StreamID, req.App, 0)
			if rtp_info.Code != 0 || rtp_info.Port == 0 {
				Logger.Error("open rtp server fail", zap.Int("code", rtp_info.Code))
			} else {
				// 向摄像头发送信令请求推实时流到zlm
				pm := &Streams{ChannelID: notifyData.DeviceID, StreamID: req.StreamID,
					ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 0,
					Mode: 0, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: req.StreamID[len(req.StreamID)-5:]}
				_, err = SipPlay(pm)
				if err != nil {
					Logger.Error("向摄像头发送信令请求实时流推标清流到zlm失败", zap.Any("deviceId", notifyData.DeviceID), zap.Error(err))
				}
			}
		}
		// 推送高清流到本地zlm
		// req.StreamID = notifyData.DeviceID + "_1"
		// resp2 := zlm_api.ZlmGetRtpInfo(Local_ZLM_Host, m.CMConfig.ZlmSecret, req)
		// if resp2.Code == 0 && !resp2.Exist {
		// 	Logger.Info("本地ZLM不存在流，开始推送本地高清流", zap.Any("deviceID", notifyData.DeviceID))

		// 	SipStopPlay(req.StreamID)
		// 	// 不存在，调用openRtpServer让本地zlm打开一个收流端口
		// 	rtp_info := zlm_api.ZlmStartRtpServer(Local_ZLM_Host, m.CMConfig.ZlmSecret, req.StreamID, 0)
		// 	if rtp_info.Code != 0 || rtp_info.Port == 0 {
		// 		Logger.Error("open rtp server fail", zap.Int("code", rtp_info.Code))
		// 	} else {
		// 		// 向摄像头发送信令请求推实时流到zlm
		// 		pm := &Streams{ChannelID: notifyData.DeviceID, StreamID: req.StreamID,
		// 			ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 1,
		// 			Mode: 0, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false}
		// 		_, err = SipPlay(pm)
		// 		if err != nil {
		// 			Logger.Error("向摄像头发送信令请求实时流推高清流到zlm失败", zap.Any("deviceId", notifyData.DeviceID), zap.Error(err))
		// 		}
		// 	}
		// }
	case NotifyMethodInvite: // 邀请通知
		ack, err = NotifyFunc(notifyData, notifyData.Method)
		if err != nil {
			Logger.Error("邀请通知服务端错误", zap.Any("data.DeviceID", notifyData.DeviceID), zap.Error(err))
		}
		if ack == nil {
			Logger.Error("通知服务端返回ack为nil", zap.Any("data.DeviceID", notifyData.DeviceID), zap.String("method", notifyData.Method))
			return
		}
	default:
		Logger.Error("notify config not found", zap.Any("data.DeviceID", notifyData.DeviceID), zap.Any("method", notifyData.Method))
	}

}

func notifyDevicesAcitve(device_id string) *Notify {
	return &Notify{
		Method:   NotifyMethodIpcActive,
		DeviceID: device_id,
	}
}
func notifyDevicesRegister(device_id string) *Notify {
	return &Notify{
		Method:   NotifyMethodIpcRegister,
		DeviceID: device_id,
	}
}

func notifyChannelsActive(device *Devices, channel *Channels) *Notify {
	return &Notify{
		Method:       NotifyMethodIpcChannelsActive,
		DeviceID:     channel.ChannelID,
		IpcId:        device.DeviceID,
		IpcIP:        device.Host,
		IpcName:      channel.Name,
		ChannelId:    channel.ChannelID,
		Manufacturer: channel.Manufacturer,
		Transport:    device.TransPort,
		StreamType:   channel.StreamType,
		Status:       channel.Status,
		ActiveTime:   channel.Active,
	}
}

func notifyInvite(channelid string) *Notify {
	return &Notify{
		Method:   NotifyMethodInvite,
		DeviceID: channelid,
	}
}
