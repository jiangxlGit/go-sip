package model

type IotDeviceAndIpcInfo struct {
	IotDeviceID string `json:"iotDeviceId"` // IOT设备id
	State      string `json:"state"`        // IOT设备状态，online：在线，offline：离线
	IpcCount  int    `json:"ipcCount"`       // IOT设备下的IPC数量
}

type IotDeviceAndAiModelInfo struct {
	IotDeviceID string `json:"iotDeviceId"` // IOT设备id
	State      string `json:"state"`        // IOT设备状态，online：在线，offline：离线
	AiModelCount  int64    `json:"aiModelCount"`       // IOT设备下的AI模型总数量
	RelationAiModelCount int64 `json:"relationAiModelCount"` // IOT设备下的已关联AI模型数量
}

type IotDeviceSimpleInfo struct {
	ID    string `json:"id"` // IOT设备id
	Name  string `json:"name"` // IOT设备名称
	State string `json:"state"` // IOT设备状态，online：在线，offline：离线
	Relation string `json:"relation"` // 关联关系， no：未关联，yes：已关联
}
