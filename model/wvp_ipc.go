package model

// 非国标ipc新增请求结构体
type IpcInfoNotGbAddReq struct {
	DeviceId     string `json:"deviceId" binding:"required"`     // 设备id
	IpcName      string `json:"ipcName" binding:"required"`      // ipc的名称
	InnerIP      string `json:"innerIp" binding:"required"`      // ipc的内网ip
	NogbUsername string `json:"nogbUsername" binding:"required"` // 非国标摄像头账号
	NogbPassword string `json:"nogbPassword" binding:"required"` // 非国标摄像头密码
	Manufacturer string `json:"manufacturer" binding:"required"` // 非国标ipc配置id（国标摄像头为厂商）
}

// ipc列表查询参数
type IpcPageQueryReq struct {
	Page     int    `json:"page" binding:"required,min=1"`
	Size     int    `json:"size" binding:"required,min=1"`
	DeviceID string `json:"deviceId" binding:"required"`
	GB       string `json:"gb"` //可选 no: 国标ipc, yes: 非国标ipc
}

type IpcControlReq struct {
	IpcId     string `json:"ipcId" binding:"required"`
	LeftRight int    `json:"leftRight" binding:"oneof=0 1 2"`  // 镜头左移右移 0:停止 1:左移 2:右移
	UpDown    int    `json:"upDown" binding:"oneof=0 1 2"`     // 镜头上移下移 0:停止 1:上移 2:下移
	InOut     int    `json:"inOut" binding:"oneof=0 1 2"`      // 镜头放大缩小 0:停止 1:缩小 2:放大
	MoveSpeed int    `json:"moveSpeed" binding:"gte=0,lte=10"` // 根据实际情况设置速度范围
	ZoomSpeed int    `json:"zoomSpeed" binding:"gte=0,lte=10"` // 根据实际情况设置速度范围
}
type IpcPushStreamResetReq struct {
	DeviceID string `json:"deviceId" binding:"required"`
	IpcId    string `json:"ipcId"`
}

type IpcRecordListQueryReq struct {
	IpcId string `json:"ipcId" binding:"required"` // 设备ID
	Type  string `json:"type" binding:"required"`  // 录像类型
	Start string `json:"start" binding:"required"` // 开始时间
	End   string `json:"end" binding:"required"`   // 结束时间
}

type IpcResetMergeStreamReq struct {
	ZlmDomain string `json:"zlmDomain" binding:"required"`
	DeviceId  string `json:"deviceId" binding:"required"`
	IpcList   string `json:"ipcList" binding:"required"` // ipc列表
}

type IpcPlaybackRecordData struct {
	StartTime   int64  `json:"startTime"`   // 录像开始时间
	EndTime     int64  `json:"endTime"`     // 录像结束时间
	FileUrl     string `json:"fileUrl"`     // 录像文件地址
	AiModelType string `json:"aiModelType"` // ai模型类型
}
