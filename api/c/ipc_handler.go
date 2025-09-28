package api

import (
	"fmt"
	db "go-sip/db/sqlite"
	. "go-sip/logger"
	"go-sip/m"
	sipapi "go-sip/sip"
	"go-sip/zlm_api"

	"go.uber.org/zap"
)

func IpcStreamReset(deviceId, ipcId string) error {
	if deviceId == "" {
		return fmt.Errorf("deviceId is empty")
	}

	sd_stream_id := ipcId + "_0"
	rtp_info := zlm_api.ZlmStartRtpServer(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, sd_stream_id, "rtp", 1)
	if rtp_info.Code != 0 || rtp_info.Port == 0 {
		Logger.Error("open rtp server fail", zap.Int("code", rtp_info.Code))
		return fmt.Errorf("open rtp server fail")
	}
	// 向摄像头发送信令请求推实时标清流到zlm
	pm := &sipapi.Streams{ChannelID: ipcId, StreamID: sd_stream_id,
		ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 0,
		Mode: 1, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: fmt.Sprintf("%s0", ipcId[len(ipcId)-5:])}
	_, err := sipapi.SipPlay(pm)
	if err != nil {
		Logger.Error("向摄像头发送信令请求实时标清流推流到zlm失败", zap.Any("deviceId", ipcId), zap.Error(err))
		return err
	}

	hd_stream_id := ipcId + "_1"
	rtp_info2 := zlm_api.ZlmStartRtpServer(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, hd_stream_id, "rtp", 0)
	if rtp_info2.Code != 0 || rtp_info2.Port == 0 {
		Logger.Error("open rtp server fail", zap.Int("code", rtp_info2.Code))
		return fmt.Errorf("open rtp server fail")
	}
	// 向摄像头发送信令请求推实时高清清流到zlm
	pm2 := &sipapi.Streams{ChannelID: ipcId, StreamID: hd_stream_id,
		ZlmIP: m.CMConfig.ZlmInnerIp, ZlmPort: rtp_info.Port, T: 0, Resolution: 1,
		Mode: 0, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: fmt.Sprintf("%s1", ipcId[len(ipcId)-5:])}
	_, err = sipapi.SipPlay(pm2)
	if err != nil {
		Logger.Error("向摄像头发送信令请求实时高清流推流到zlm失败", zap.Any("deviceId", ipcId), zap.Error(err))
		return err
	}

	return nil

}
