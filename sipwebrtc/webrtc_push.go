package sipwebrtc

import (
	"context"
	"fmt"
	. "go-sip/logger"
	"os/exec"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"go.uber.org/zap"

	"github.com/mesilliac/pulse-simple"
)

func NewWebrtcPushClient() (*RealtimeClient, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("创建PeerConnection失败: %v", err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypePCMA,
			ClockRate: 8000,
			Channels:  1,
		},
		"audio",
		"pion-pcma-audio",
	)
	if err != nil {
		return nil, fmt.Errorf("创建audio track失败: %v", err)
	}

	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return nil, fmt.Errorf("创建add track失败: %v", err)
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
		case webrtc.PeerConnectionStateFailed:
			Logger.Info("webrtc 连接失败")
		case webrtc.PeerConnectionStateClosed:
			if peerConnection != nil {
				peerConnection.Close()
			}
			Logger.Info("webrtc 连接已关闭")
		default:
			Logger.Info("连接状态: ", zap.Any("连接状态", s.String()))
		}
	})

	return &RealtimeClient{
		PeerConnection_push: peerConnection,
		AudioTrack:          audioTrack,
	}, nil
}

func (c *RealtimeClient) HandlePushStream(ctx context.CancelFunc) {

	var ss = pulse.SampleSpec{pulse.SAMPLE_ALAW, 8000, 1}

	bufferAttr := pulse.BufferAttr{
		Maxlength: 160 * 4, // 最大缓冲区长度（字节）
		Tlength:   160 * 2, // 目标缓冲长度（播放时有效）
		Prebuf:    0,       // 播放时预缓冲（可设为0）
		Minreq:    160,     // 播放时请求的最小数据
		Fragsize:  160,     // 录音时每次读取的数据块大小
	}

	// create the capture device and audio storage buffers
	capture, err := pulse.NewStream(
		"",                          // serverName
		"audio record",              // clientName
		pulse.STREAM_RECORD,         // record mode
		"echo_cancel_source",        // <<< 指定采集源为回声消除后的
		"Echo-Cancel Source Stream", // stream name
		&ss,                         // sample spec
		nil, &bufferAttr,            // channel map, buffer attr
	)
	defer capture.Free()
	if err != nil {
		fmt.Printf("Could not create capture stream: %s\n", err)
		return
	}
	SetSourceVolume(150) // 设置音量为 100%
	// 9. 开始发送 PCMA 音频数据

	defer ctx()
	defer c.ClosePush()

	buffer := make([]byte, 160) // 20ms 的音频数据缓冲区
	for {
		_, err := capture.Read(buffer)
		if err != nil {
			// Logger.Error("读音频帧失败:", zap.Error(err))
			continue
		}
		// 发送音频帧
		if err := c.AudioTrack.WriteSample(media.Sample{
			Data:      buffer,
			Duration:  20 * time.Millisecond, // 关键：指定帧持续时间
			Timestamp: time.Now(),            // 可选：设置时间戳
		}); err != nil {
			Logger.Error("发送音频帧失败: ", zap.Error(err))
			break
		}
	}

}

func SetSourceVolume(volume int) error {

	v := fmt.Sprintf("%d%%", volume)
	cmd := exec.Command("pactl", "set-source-volume", "echo_cancel_source", v)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("设置音量失败: %v\n输出: %s", err, output)
	}
	return nil
}
