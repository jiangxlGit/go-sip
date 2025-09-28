package grpc_client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	capi "go-sip/api/c"
	db "go-sip/db/sqlite"
	"go-sip/grpc_api"
	. "go-sip/logger"
	"go-sip/m"
	pb "go-sip/signaling"
	sipapi "go-sip/sip"
	"go-sip/sipwebrtc"
	"go-sip/utils"
	"go-sip/zlm_api"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var SipCli *SipClient

type SipClient struct {
	clientID  string
	conn      *grpc.ClientConn
	stream    pb.SipService_StreamChannelClient
	client    pb.SipServiceClient
	AudioDone context.Context
}

func NewSipClient(clientID string) *SipClient {

	client := &SipClient{
		clientID: clientID,
	}

	sipapi.NotifyFunc = client.IpcEventReq
	sipapi.InviteFunc = client.IpcInviteReq
	sipapi.NotifyAiEventFunc = client.AiEventReq
	sipapi.NotifyOTAUpgradeFunc = client.OTAUpgradeReq
	return client
}

func GetSipClient() *SipClient {
	return SipCli
}

func (c *SipClient) Connect(addr string) error {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second, // PING发送间隔从默认15秒改为30秒
			Timeout:             20 * time.Second, // 等待PING响应的超时时间
			PermitWithoutStream: true,             // 允许无活跃流时发送PING
		}),
	)
	if err != nil {
		return err
	}

	client := pb.NewSipServiceClient(conn)

	stream, err := client.StreamChannel(context.Background())
	if err != nil {
		conn.Close()
		return err
	}

	// 发送注册信息
	if err := stream.Send(&pb.ClientMessage{
		Content: &pb.ClientMessage_Register{
			Register: &pb.ClientRegister{
				ClientId:   c.clientID,
				Version:    "1.0.0",
				DeviceType: GetPlatform(),
			},
		},
	}); err != nil {
		conn.Close()
		return err
	}

	c.conn = conn
	c.stream = stream
	c.client = client
	return nil
}

func GetPlatform() string {

	if model, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		if strings.Contains(string(model), "rk3568") {
			return "rk3568"
		} else if strings.Contains(string(model), "rk3588") {
			return "rk3588"
		} else if strings.Contains(string(model), "rk3576") {
			return "rk3576"
		}
	}

	return m.CMConfig.DeviceType
}

func (c *SipClient) IpcEventReq(notifyData *sipapi.Notify, msg_type string) (*pb.IpcEventAck, error) {

	if c.client != nil {
		return c.client.IpcEventReq(context.Background(), &pb.IpcEventRequest{
			ClientId:     c.clientID,
			IpcId:        notifyData.IpcId,
			IpcIp:        notifyData.IpcIP,
			Event:        msg_type,
			ChannelId:    notifyData.ChannelId,
			IpcName:      notifyData.IpcName,
			Status:       notifyData.Status,
			ActiveTime:   notifyData.ActiveTime,
			Manufacturer: notifyData.Manufacturer,
			Transport:    notifyData.Transport,
			Streamtype:   notifyData.StreamType,
		})
	}

	return nil, fmt.Errorf("client已断开")

}

func (c *SipClient) IpcInviteReq(ipc_id string) (*pb.IpcInviteAck, error) {

	if c.client != nil {
		return c.client.IpcInviteReq(context.Background(), &pb.IpcInviteRequest{
			ClientId: c.clientID,
			IpcId:    ipc_id,
		})
	}

	return nil, fmt.Errorf("client已断开")

}

func (c *SipClient) AiEventReq(device_id, rk_platform, stream_id string, event_id int64, class_name string, max_score float64, count int64) (*pb.AIEventAck, error) {
	Logger.Debug("==== AIEventReq ====", zap.Any("streamId", stream_id), zap.Any("class_name", class_name), zap.Any("max_score", max_score), zap.Any("count", count))
	if c.client != nil {
		return c.client.AiEventReq(context.Background(), &pb.AIEventRequest{
			DeviceId:   device_id,
			RkPlatform: rk_platform,
			StreamId:   stream_id,
			ClientId:   c.clientID,
			EventId:    event_id,
			ClassName:  class_name,
			MaxScore:   max_score,
			Count:      count,
		})
	}

	return nil, fmt.Errorf("client已断开")
}

func (c *SipClient) OTAUpgradeReq(device_id, firmware_id, firmware_version, upgrade_complete, upgrade_progress, upgrade_error string) (*pb.OTAUpgradeAck, error) {
	Logger.Debug("==== OTAUpgradeReq ====", zap.Any("device_id", device_id), zap.Any("firmware_id", firmware_id), zap.Any("firmware_version", firmware_version),
		zap.Any("upgrade_progress", upgrade_progress), zap.Any("upgrade_error", upgrade_error))
	if c.client != nil {
		return c.client.OTAUpgradeReq(context.Background(), &pb.OTAUpgradeRequest{
			DeviceId:        device_id,
			FirmwareId:      firmware_id,
			FirmwareVersion: firmware_version,
			UpgradeComplete: upgrade_complete,
			UpgradeProgress: upgrade_progress,
			UpgradeError:    upgrade_error,
		})
	}

	return nil, fmt.Errorf("client已断开")
}

func (c *SipClient) Run() {
	defer c.conn.Close()

	// 用一个 channel 缓冲发送结果，避免多个 goroutine 并发写 stream
	resultChan := make(chan *pb.ClientMessage, 100)

	// 独立 goroutine 负责串行发送
	go func() {
		for msg := range resultChan {
			if err := c.stream.Send(msg); err != nil {
				Logger.Error("发送结果失败", zap.Error(err))
				return
			}
		}
	}()

	// 命令处理循环
	for {
		cmd, err := c.stream.Recv()
		if err != nil {
			Logger.Error("连接断开: ", zap.Error(err))
			close(resultChan) // 关闭发送协程
			return
		}
		if cmd == nil {
			Logger.Error("连接断开, cmd为空")
			continue
		}

		go func(cmd *pb.ServerCommand) {
			// 处理服务端命令
			Logger.Info("收到服务端命令", zap.Any("method", cmd.Method), zap.Any("MsgID", cmd.MsgID))
			result := c.executeCommand(cmd)

			if result != nil {

				// 返回执行结果
				resultChan <- &pb.ClientMessage{
					Content: &pb.ClientMessage_Result{
						Result: &pb.CommandResult{
							MsgID:   cmd.MsgID,
							Success: result.Success,
							Payload: result.Msg,
						},
					},
				}
			}

		}(cmd)
	}
}

type CommandResult struct {
	Success bool
	Msg     []byte
}

func (c *SipClient) executeCommand(cmd *pb.ServerCommand) *CommandResult {

	rsp := &CommandResult{}
	rsp.Success = true
	rsp.Msg = []byte("执行成功")
	switch cmd.Method {
	case m.Ping:

		return rsp

	case m.Play:
		d := &grpc_api.Sip_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("实时流点播执行失败: %v", err))
			return rsp
		}

		stream_arr := strings.Split(d.ChannelID, "_")
		if len(stream_arr) <= 1 {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		var zlmGetMediaListReq = zlm_api.ZlmGetMediaListReq{}
		zlmGetMediaListReq.Vhost = "__defaultVhost__"
		zlmGetMediaListReq.Schema = "rtsp"
		zlmGetMediaListReq.App = "rtp"
		zlmGetMediaListReq.StreamID = d.ChannelID

		var ssrc string
		var resolution int
		if strings.HasPrefix(d.ChannelID, "IPC") {
			err = capi.IpcPushStreamReset(d.DeviceID, stream_arr[0], nil)
			if err != nil {
				rsp.Success = false
				rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
				return rsp
			}
			ssrc = fmt.Sprintf("1%s%s", stream_arr[0][len(stream_arr[0])-4:], d.DeviceID[len(d.DeviceID)-3:])
		} else {
			ssrc = fmt.Sprintf("%s%s", stream_arr[0][:5], d.DeviceID[len(d.DeviceID)-3:])
			var destZlmHost string
			var destZlmSecret string
			var destZlmIp string
			// 标清流
			if stream_arr[1] == "0" {
				resolution = 0
				destZlmHost = sipapi.Local_ZLM_Host
				destZlmSecret = m.CMConfig.ZlmSecret
				destZlmIp = m.CMConfig.ZlmInnerIp
			} else if stream_arr[1] == "1" {
				resolution = 1
				destZlmHost = d.ZlmDomain
				destZlmSecret = d.ZlmSecret
				destZlmIp = d.ZLMIP
			}
			// 查询zlm是否存在此流
			resp := zlm_api.ZlmGetMediaList(destZlmHost, destZlmSecret, zlmGetMediaListReq)
			if resp.Code == 0 && len(resp.Data) > 0 {
				Logger.Info("zlm流已存在", zap.Any("destZlmHost", destZlmHost), zap.Any("streamId", d.ChannelID))
			} else {
				Logger.Info("zlm流不存在", zap.Any("destZlmHost", destZlmHost), zap.Any("streamId", d.ChannelID))
				rtpPort := d.ZLMPort
				if rtpPort == 0 {
					// rtpPort为0，不存在，调用openRtpServer让zlm打开一个收流udp端口
					rtp_info := zlm_api.ZlmStartRtpServer(destZlmHost, destZlmSecret, d.ChannelID, zlmGetMediaListReq.App, 0)
					if rtp_info.Code != 0 || rtp_info.Port == 0 {
						Logger.Error("open rtp server fail", zap.Any("destZlmHost", destZlmHost), zap.Any("streamId", d.ChannelID), zap.Int("ZlmStartRtpServer resp code", rtp_info.Code))
						rsp.Success = false
						rsp.Msg = []byte(fmt.Sprintf("实时流点播执行失败: %v", err))
						return rsp
					}
					rtpPort = rtp_info.Port
				}
				// 向摄像头发送信令请求推实时流到本地zlm
				pm := &sipapi.Streams{ChannelID: stream_arr[0], StreamID: d.ChannelID,
					ZlmIP: destZlmIp, ZlmPort: rtpPort, T: 0, Resolution: resolution,
					Mode: d.Mode, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: ssrc}
				_, err = sipapi.SipPlay(pm)
				if err != nil {
					Logger.Error("向摄像头发送信令请求实时流推标清流到zlm失败", zap.Any("ipcId", stream_arr[0]), zap.Error(err))
					rsp.Success = false
					rsp.Msg = []byte(fmt.Sprintf("实时流点播执行失败: %v", err))
					return rsp
				}
			}
		}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		count := 0
		for range ticker.C {
			count++
			if count >= 15 {
				break
			}
			if !strings.HasPrefix(d.ChannelID, "IPC") && stream_arr[1] == "1" {
				// 发送信令指令给摄像头，推高清流到远程zlm，现在查询远程zlm是否存在高清流
				resp3 := zlm_api.ZlmGetMediaList(d.ZlmDomain, d.ZlmSecret, zlmGetMediaListReq)
				if resp3.Code == 0 && len(resp3.Data) > 0 {
					rsp.Success = true
					rsp.Msg = []byte("远程zlm流已存在高清流")
					return rsp
				}
			} else {
				// 如果是国标标清流或非国标流，则先查询本地zlm是否存在该流，不存在则直接break
				resp2 := zlm_api.ZlmGetMediaList(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, zlmGetMediaListReq)
				if resp2.Code != 0 || len(resp2.Data) == 0 {
					break
				}
				// 查询远程zlm是否存在该流
				resp3 := zlm_api.ZlmGetMediaList(d.ZlmDomain, d.ZlmSecret, zlmGetMediaListReq)
				if resp3.Code == 0 && len(resp3.Data) > 0 {
					rsp.Success = true
					rsp.Msg = []byte("远程zlm流已存在")
					return rsp
				}
				Logger.Info("远程zlm不存在此流，开始推本地zlm标清流到远程zlm", zap.String("stream id", d.ChannelID))
				// 调用板端zlm的active接口，发送给远程zlm
				var req = zlm_api.ZlmStartSendRtpReq{}
				if d.App == "" {
					req.App = "rtp"
				} else {
					req.App = zlmGetMediaListReq.App
				}
				req.StreamID = d.ChannelID
				req.Vhost = "__defaultVhost__"
				req.DstUrl = d.ZLMIP
				req.DstPort = strconv.Itoa(d.ZLMPort)
				req.Ssrc = ssrc
				if stream_arr[1] == "1" {
					req.Ssrc = "2"
				}
				if d.Mode == 1 {
					req.IsUdp = "0"
				} else {
					req.IsUdp = "1"
				}
				Logger.Info("ZlmStartSendRtp req", zap.Any("app", zlmGetMediaListReq.App), zap.Any("req", req))
				_resp := zlm_api.ZlmStartSendRtp(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, req)
				if _resp.Code == 0 && _resp.LocalPort > 0 {
					Logger.Info("实时流点播发送RTP成功", zap.Any("stream id", d.ChannelID), zap.Any("local_port", _resp.LocalPort))
					rsp.Success = true
					rsp.Msg = []byte("实时流点播发送RTP成功")
					return rsp
				}
			}
		}
	case m.StopPlay:
		d := &grpc_api.Sip_Stop_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		// 判断流是否存在zlm
		var zlmGetMediaListReq = zlm_api.ZlmGetMediaListReq{}
		zlmGetMediaListReq.App = "rtp"
		zlmGetMediaListReq.Vhost = "__defaultVhost__"
		zlmGetMediaListReq.Schema = "rtsp"
		zlmGetMediaListReq.StreamID = d.StreamID

		resp := zlm_api.ZlmGetMediaList(d.ZlmDomain, d.ZlmSecret, zlmGetMediaListReq)
		if resp.Code == 0 && len(resp.Data) > 0 {
			Logger.Info("StopPlay远程zlm流存在,开始调用stopSendRtp", zap.Any("streamId", d.StreamID))
			// stopSendRtp 本地zlm
			var req = zlm_api.ZlmStopSendRtpReq{}
			if d.App == "" {
				req.App = "rtp"
			} else {
				req.App = d.App
			}
			req.StreamID = d.StreamID
			req.Vhost = "__defaultVhost__"
			stream_arr := strings.Split(d.StreamID, "_")
			req.Ssrc = "1"
			if stream_arr[1] == "1" {
				req.Ssrc = "2"
			}
			_resp := zlm_api.ZlmStopSendRtp(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret, req)
			if _resp.Code == 0 {
				rsp.Success = true
				rsp.Msg = []byte("停止RTP推流成功")
				return rsp
			}
			Logger.Info("调用stopSendRtp成功，关闭远程zlm的rtp端口", zap.Any("streamId", d.StreamID))
			closeRtpRsp := zlm_api.ZlmCloseRtpServer(d.ZlmDomain, d.ZlmSecret, d.StreamID)
			if closeRtpRsp.Code == 0 && closeRtpRsp.Hit >= 1 {
				rsp.Success = true
				rsp.Msg = []byte("关闭远程zlm的rtp端口成功")
				return rsp
			} else {
				rsp.Success = true
				rsp.Msg = []byte("关闭远程zlm的rtp端口成功失败")
				return rsp
			}
		}

		// err = sipapi.SipStopPlay(d.StreamID)
		// if err != nil {
		// 	rsp.Success = false
		// 	rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
		// 	return rsp
		// }

	case m.PlayBack:

		d := &grpc_api.Sip_Play_Back_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
		s := time.Unix(d.StartTime, 0)
		e := time.Unix(d.EndTime, 0)
		pm := &sipapi.Streams{ChannelID: d.ChannelID, StreamID: fmt.Sprintf("%s_%d_%d", d.ChannelID,
			d.StartTime, d.EndTime), ZlmIP: d.ZLMIP, ZlmPort: d.ZLMPort, S: s, E: e, T: 1, Resolution: d.Resolution,
			Mode: d.Mode, Ttag: db.M{}, Ftag: db.M{}, OnlyAudio: false, Ssrc: d.ChannelID[len(d.ChannelID)-5:]}
		_, err = sipapi.SipPlay(pm)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

	case m.ResumePlay:
		d := &grpc_api.Sip_Resume_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipResumePlay(d.StreamID)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.PausePlay:
		d := &grpc_api.Sip_Pause_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipPausePlay(d.StreamID)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

	case m.SeekPlay:
		d := &grpc_api.Sip_Seek_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipSeekPlay(d.StreamID, d.SubTime)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.SpeedPlay:
		d := &grpc_api.Sip_Speed_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipSpeedPlay(d.StreamID, d.Speed)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

	case m.RecordList:

		{

			d := &grpc_api.Sip_Play_Back_Recocd_List_Req{}
			err := utils.JSONDecode(cmd.Payload, d)
			if err != nil {
				Logger.Error("Unmarshal failed ", zap.Error(err))
				rsp.Success = false
				rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
				return rsp
			}

			channel := &sipapi.Channels{ChannelID: d.ChannelID}

			if err := db.Get(db.DBClient, channel); err != nil {
				rsp.Success = false
				rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
				return rsp
			}
			res, err := sipapi.SipRecordList(channel, d.StartTime, d.EndTime)
			if err != nil {
				rsp.Success = false
				rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
				return rsp
			}

			record := utils.JSONEncode(res)

			rsp.Msg = record
			return rsp
		}

	case m.Broadcast:

		d := &grpc_api.Sip_Ipc_BroadCast_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipIpcBroadCast(d.ChannelID)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.PlayIPCAudio:
		d := &grpc_api.Sip_Play_IPC_Audio_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.SipPlayAudio(d.ChannelID, d.ZLMPort, d.ZLMIP)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

	case m.PlayAudio:

		d := &grpc_api.Sip_Audio_Play_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		bluez_index, err := GetBluezSinkIndex()

		if err == nil {
			// 挂起音频设备
			if err := exec.Command("pacmd", " suspend-sink", bluez_index, "1").Run(); err != nil {
				Logger.Warn("挂起音频设备失败", zap.Error(err))
			} else {
				Logger.Info("挂起音频设备失败成功")
			}
		}

		go func() {

			AudioDone, cancel := context.WithCancel(context.Background())

			client, err := sipwebrtc.NewWebRtcPlayClient(cancel)
			if err != nil {
				Logger.Error("new webrtc client failed", zap.Error(err))
			}
			defer client.ClosePlay()
			SinalUrl := fmt.Sprintf("http://%s:%s/index/api/webrtc?app=audio&stream=%s&type=play&sign=%s", d.ZLMIP, d.ZLMPort, d.StreamID, d.Token)

			// 连接到信令服务器
			if err := client.Connect(client.PeerConnection_play, SinalUrl); err != nil {
				Logger.Error("connect 信令服务器失败", zap.Error(err))
				return
			}

			<-AudioDone.Done()

			Logger.Info("语音播放退出")

		}()

		Logger.Info("语音播放开始", zap.String("streamID", d.StreamID), zap.String("ZLMIP", d.ZLMIP), zap.String("ZLMPort", d.ZLMPort))

	case m.PushAudio:

		d := &grpc_api.Sip_Audio_Push_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
		go func() {

			SinalUrl := fmt.Sprintf("http://%s:%s/index/api/webrtc?app=audio&stream=%s&type=push&sign=%s", d.ZLMIP, d.ZLMPort, d.StreamID, d.Token)

			AudioDone, cancel := context.WithCancel(context.Background())

			client, err := sipwebrtc.NewWebrtcPushClient()
			if err != nil {
				Logger.Error("new webrtc client failed", zap.Error(err))
			}
			defer client.ClosePush()

			// 连接到信令服务器
			if err := client.Connect(client.PeerConnection_push, SinalUrl); err != nil {
				Logger.Error("connect 信令服务器失败", zap.Error(err))
			}

			client.HandlePushStream(cancel)

			<-AudioDone.Done()

			Logger.Info("语音推流退出")

		}()

		Logger.Info("语音推流开始", zap.String("streamID", d.StreamID), zap.String("ZLMIP", d.ZLMIP), zap.String("ZLMPort", d.ZLMPort))

	case m.SetVolume:
		d := &grpc_api.Sip_Set_Volume_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
		err = sipwebrtc.SetSourceVolume(d.Volume)
		if err != nil {
			Logger.Error("设置中控语音对讲音量失败", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

	case m.CloseAudio:
		Logger.Info("关闭语音对讲")

	case m.DeviceControl:
		d := &grpc_api.Sip_IPC_Control_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = sipapi.DeviceControl(d.DeviceID, d.LeftRight, d.UpDown, d.InOut, d.MoveSpeed, d.ZoomSpeed)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.IpcPushStreamReset:
		d := &grpc_api.Sip_Ipc_Push_Stream_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = capi.IpcPushStreamReset(d.DeviceID, d.IpcId, nil)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.IpcStreamReset:
		d := &grpc_api.Sip_Ipc_Push_Stream_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}

		err = capi.IpcStreamReset(d.DeviceID, d.IpcId)
		if err != nil {
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
	case m.DeviceOTA:
		d := &grpc_api.OTA_Event_Req{}
		err := utils.JSONDecode(cmd.Payload, d)
		if err != nil {
			Logger.Error("Unmarshal failed ", zap.Error(err))
			rsp.Success = false
			rsp.Msg = []byte(fmt.Sprintf("执行失败: %v", err))
			return rsp
		}
		go capi.DeviceOTA(d.DeviceID, d.FirmwareID, d.FirmwareDownloadURL, d.FirmwareMD5, d.FirmwareVersion)
	}

	return rsp
}

func GetBluezSinkIndex() (string, error) {
	// 执行pactl命令
	cmd := exec.Command("pactl", "list", "short", "sinks")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行pactl失败: %v", err)
	}

	// 解析输出
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "bluez") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				return fields[0], nil
			}
		}
	}

	return "", fmt.Errorf("未找到蓝牙音频设备")
}
