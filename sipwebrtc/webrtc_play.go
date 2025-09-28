package sipwebrtc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	. "go-sip/logger"

	"github.com/mesilliac/pulse-simple"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

// RealtimeClient 包含 WebRTC 连接和音频播放相关的功能
type RealtimeClient struct {
	PeerConnection_play *webrtc.PeerConnection
	PeerConnection_push *webrtc.PeerConnection

	AudioTrack *webrtc.TrackLocalStaticSample

	isPush bool // 是否为推流模式
	isPlay bool // 是否为拉流模式
}

// NewRealtimeClient 创建并返回一个新的 RealtimeClient 实例
func NewWebRtcPlayClient(ctx context.CancelFunc) (*RealtimeClient, error) {
	// 配置媒体引擎
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, fmt.Errorf("注册默认编解码器失败: %v", err)
	}

	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypePCMA,
			ClockRate:   sampleRate,
			Channels:    channelCount,
			SDPFmtpLine: "minptime=10;useinbandfec=1",
		},
		PayloadType: 8,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, fmt.Errorf("注册G711编解码器失败: %v", err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// 创建 PeerConnection
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("创建PeerConnection失败: %v", err)
	}

	// 6. 连接状态变化处理
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {

		switch s {
		case webrtc.PeerConnectionStateConnected:
			Logger.Info("webrtc 连接已建立")
		case webrtc.PeerConnectionStateDisconnected:
			if peerConnection != nil {
				peerConnection.Close()
			}
			Logger.Info("webrtc 连接已断开")
		default:
			Logger.Info("连接状态: ", zap.Any("连接状态", s.String()))
		}
	})

	client := &RealtimeClient{
		PeerConnection_play: peerConnection,
	}
	// 处理远端音频轨道
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			go client.HandleRemoteTrack(ctx, track)
		}
	})

	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return nil, fmt.Errorf("添加音频接收器失败: %v", err)
	}

	return client, nil
}

// 处理远端音频轨道
func (c *RealtimeClient) HandleRemoteTrack(audioctx context.CancelFunc, track *webrtc.TrackRemote) {

	defer audioctx()
	defer c.ClosePlay()
	ss := pulse.SampleSpec{pulse.SAMPLE_ALAW, 8000, 1}

	playback, err := pulse.NewStream(
		"",                        // serverName
		"audio playback",          // clientName
		pulse.STREAM_PLAYBACK,     // record mode
		"echo_cancel_sink",        // <<< 指定采集源为回声消除后的
		"Echo-Cancel Sink Stream", // stream name
		&ss,                       // sample spec
		nil, nil,                  // channel map, buffer attr
	)
	if playback != nil {
		defer playback.Free()
		defer playback.Drain()
	}

	if err != nil {
		fmt.Printf("Could not create playback stream: %s\n", err)
		return
	}

	for {
		// 读取 RTP 包
		rtp, _, err := track.ReadRTP()
		if err != nil {
			Logger.Error("读取 RTP 包失败: ", zap.Error(err))
			break
		}
		//pcma := g711.DecodeAlaw(rtp.Payload)

		_, err = playback.Write(rtp.Payload)
		if err != nil {
			Logger.Error("写入失败:", zap.Error(err))
			continue
		}
	}

}

// Connect 连接到信令服务器
func (c *RealtimeClient) Connect(peerConnection *webrtc.PeerConnection, signalingURL string) error {

	// 创建Offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("创建Offer失败: %v", err)
	}
	// 设置本地描述
	if err := peerConnection.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("设置本地描述失败: %v", err)
	}

	// 发送Offer到信令服务器
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", signalingURL, bytes.NewBufferString(offer.SDP))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var answer RTCSessionDescription
	if err := json.Unmarshal(body, &answer); err != nil {
		return fmt.Errorf("解析Answer失败: %v", err)
	}

	// 设置远程描述
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answer.SDP,
	}); err != nil {
		return fmt.Errorf("设置远程描述失败: %v", err)
	}
	Logger.Info("WebRTC 连接已建立", zap.String("SDP", answer.SDP))
	return nil
}

// Close 关闭连接并清理资源
func (c *RealtimeClient) ClosePlay() {
	if c.PeerConnection_play != nil {
		_ = c.PeerConnection_play.Close()
	}
}

func (c *RealtimeClient) ClosePush() {
	if c.PeerConnection_push != nil {
		_ = c.PeerConnection_push.Close()
	}
}
