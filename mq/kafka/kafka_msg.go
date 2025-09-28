package kafka

type KafkaMsg struct {
	DeviceSerial string `json:"deviceSerial"`
	StoreNo      string `json:"storeNo"`
	Text         string `json:"text"`
	Timestamp    int64  `json:"timestamp"`
}

type AiHumanEventKafkaMsg struct {
	DeviceSerial string `json:"deviceSerial"`
	StoreNo      string `json:"storeNo"`
	AiEvent      string `json:"aiEvent"`
	HumanCount   int64  `json:"humanCount"`
	Timestamp    int64  `json:"timestamp"`
}

type DeviceStateKafkaMsg struct {
	// 设备类型: 9: 汉唐云三代中控摄像头, a: 汉唐云三代中控门禁
	DeviceType   string `json:"deviceType"`
	DeviceSerial string `json:"deviceSerial"` // 设备序列号
	StoreNo      string `json:"storeNo"`      // 门店编号
	OnlineState  int64  `json:"onlineState"`  // "1 - 在线, 0 - 离线"
	Timestamp    int64  `json:"timestamp"`    // 时间戳
}
