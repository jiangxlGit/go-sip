package sipwebrtc

const (
	sampleRate      = 8000 // G.711 标准采样率
	channelCount    = 1    // 单声道
	frameDuration   = 20   // 20ms 帧
	samplesPerFrame = 160
)

type RTCSessionDescription struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
	Code int    `json:"code"`
	ID   string `json:"id"`
}
