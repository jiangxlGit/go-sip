package model

type IpcInfo struct {
	// 摄像头ipcid
	IpcId string `json:"ipcId" gorm:"column:ipcId"`
	// 摄像头ip
	IpcIP string `json:"ipcIP"`
	// Name 设备名称
	IpcName string `json:"ipcName" gorm:"column:ipcName" `
	// DeviceID 设备id
	DeviceID string `json:"deviceId" gorm:"column:deviceId"`
	// 摄像头channelId
	ChannelId string `json:"channelId" gorm:"column:channelId"`
	// 制造厂商
	Manufacturer string `json:"manufacturer"  gorm:"column:manufacturer"`
	// 信令传输协议
	Transport string `json:"transport" gorm:"column:transport"`
	// 流传输模式
	StreamType string `json:"streamtype"  gorm:"column:streamtype"`
	// Status 状态  ON 在线, OFFLINE 离线
	Status string `json:"status"  gorm:"column:status"`
	// Active 最后活跃时间
	ActiveTime int64 `json:"activeTime"  gorm:"column:activeTime"`
	// ipc关联的sip服务id
	SipId string `json:"sipId" gorm:"column:sipId"`
	// 上一次心跳时间
	LastHeartbeatTime int64 `json:"lastHeartbeatTime" gorm:"column:lastHeartbeatTime"`
	// 内网ip
	InnerIP string `json:"innerIp"`
	// 非国标摄像头账号
	NogbUsername string `json:"nogbUsername"`
	// 非国标摄像头密码
	NogbPassword string `json:"nogbPassword"`
}

type IotNotGbIpcInfo struct {
	// 摄像头ipcid
	IpcId string `json:"ipcId"`
	// ipc名称
	IpcName string `json:"ipcName"`
	// 内网ip
	InnerIP string `json:"innerIp"`
	// 配置名称
	Manufacturer string `json:"manufacturer"`
	// 用户名
	Username string `json:"username"`
	// 密码
	Password string `json:"password"`
	// rtsp子码流后缀
	RtspSubSuffix string `json:"rtspSubSuffix"`
	// rtsp主码流后缀
	RtspMainSuffix string `json:"rtspMainSuffix"`
	// Status 状态  ON 在线, OFFLINE 离线
	Status string `json:"status"`
}

type NotGBConfig struct {
	ID             int64  `json:"id"`
	Manufacturer   string `json:"manufacturer"`
	RtspSubSuffix  string `json:"rtspSubSuffix"`
	RtspMainSuffix string `json:"rtspMainSuffix"`
}

type NotGbIpcInfoUpdateReq struct {
	IpcId string `json:"ipcId" binding:"required"`
	// ipc名称
	IpcName string `json:"ipcName"`
	// 内网ip
	InnerIP string `json:"innerIp"`
	// 非国标摄像头账号
	NogbUsername string `json:"nogbUsername"`
	// 非国标摄像头密码
	NogbPassword string `json:"nogbPassword"`
}

type NotGBConfigAddOrUpdateReq struct {
	IpcId          string `json:"ipcId"`
	Manufacturer   string `json:"manufacturer" binding:"required"`    // 配置名称
	RtspSubSuffix  string `json:"rtspSubSuffix" binding:"required"`   // 子流后缀
	RtspMainSuffix string `json:"rtspMainSuffix"  binding:"required"` // 主流后缀
}

func FormNotGbConfig(req *NotGBConfigAddOrUpdateReq) *NotGBConfig {
	return &NotGBConfig{
		Manufacturer:   req.Manufacturer,
		RtspMainSuffix: req.RtspMainSuffix,
		RtspSubSuffix:  req.RtspSubSuffix,
	}
}

type IpcHeartbeatInfo struct {
	IpcId             string `json:"ipcId"`
	LastHeartbeatTime int64  `json:"lastHeartbeat"`
	Status            string `json:"status"`
}
