package mqtt

type OTAFirmwarePullMqttMsg struct {
	Headers        OTAFirmwarePullHeaders `json:"headers"`
	DeviceID       string                 `json:"deviceId"`
	Timestamp      int64                  `json:"timestamp"`
	MessageID      string                 `json:"messageId"`
	RequestVersion string                 `json:"requestVersion"`
}

type OTAFirmwarePullHeaders struct {
	Force  bool `json:"force"`
	Latest bool `json:"latest"`
}

// FirmwareReply 表示接收到的固件回复消息
type FirmwareReplyMqttMsg struct {
	Success     bool                   `json:"success"`
	MessageID   string                 `json:"messageId"`
	DeviceID    string                 `json:"deviceId"`
	Timestamp   int64                  `json:"timestamp"`
	Headers     FirmwareReplyHeaders   `json:"headers"`
	URL         string                 `json:"url"`
	Version     string                 `json:"version"`
	Parameters  map[string]interface{} `json:"parameters"`
	Sign        string                 `json:"sign"`
	SignMethod  string                 `json:"signMethod"`
	FirmwareID  string                 `json:"firmwareId"`
	Size        int64                  `json:"size"`
	MessageType string                 `json:"messageType"`
}

// FirmwareHeaders 内部 headers 结构
type FirmwareReplyHeaders struct {
	Async    bool   `json:"async"`
	SendFrom string `json:"send-from"`
}
