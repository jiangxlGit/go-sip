package model

type ZLMStreamChangedData struct {
	MediaServerId string `json:"mediaServerId"`
	Regist        bool   `json:"regist"`
	Params        string `json:"params"`
	APP           string `json:"app"`
	Stream        string `json:"stream"`
	Schema        string `json:"schema"`
}

type ZlmInfo struct {
	ZlmIp     string `json:"zlmIp" validate:"required"`     // ZLM IP
	ZlmDomain string `json:"zlmDomain" validate:"required"` // ZLM 域名
	ZlmSecret string `json:"zlmSecret" validate:"required"` // ZLM 密钥
	ZlmPort   string `json:"zlmPort" validate:"required"`   // ZLM 端口
}

type ZlmAndRegionInfo struct {
	ID            int    `json:"id"`
	ZlmIp         string `json:"zlmIp"`
	ZlmPort       int    `json:"zlmPort"`
	ZlmDomain     string `json:"zlmDomain"`
	ZlmSecret     string `json:"zlmSecret"`
	ZlmNodeRegion string `json:"zlmNodeRegion"`
	RegionCode    string `json:"regionCode"`
	ZlmNodeStatus string `json:"zlmNodeStatus"`
	Remarks       string `json:"remarks"`
}

type StreamMergeConfigDTO struct {
	GapV   int        `json:"gapv"`
	GapH   int        `json:"gaph"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Row    int        `json:"row"`
	Col    int        `json:"col"`
	ID     string     `json:"id"`
	URL    [][]string `json:"url"`
	Span   []int      `json:"span"`
}

type StreamMergeInfoDTO struct {
	DeviceId  string   `json:"deviceId" validate:"required"`        // 设备ID
	IpcIdList []string `json:"ipcIdList" validate:"required,min=1"` // 国标设备ID列表，不能为空
	StreamId  string   `json:"streamId" validate:"required"`        // 流 ID
	Type      int      `json:"type" validate:"required"`            // 合屏还是切屏 1 合屏 2 切屏
}

type ZLMRecordMp4Data struct {
	MediaServerId      string  `json:"mediaServerId"`
	App                string  `json:"app"`
	FileName           string  `json:"file_name"`
	FilePath           string  `json:"file_path"`
	FileSize           int64   `json:"file_size"`
	Folder             string  `json:"folder"`
	StartTime          int64   `json:"start_time"`
	Stream             string  `json:"stream"`
	TimeLen            float64 `json:"time_len"`
	URL                string  `json:"url"`
	Vhost              string  `json:"vhost"`
	FileOssDownloadUrl string  `json:"file_oss_download_url"`
}

type ZLMStreamNotFoundData struct {
	APP           string `json:"app"`
	Params        string `json:"params"`
	Stream        string `json:"stream"`
	Schema        string `json:"schema"`
	ID            string `json:"id"`
	IP            string `json:"ip"`
	Port          int    `json:"port"`
	MediaServerID string `json:"mediaServerId"`
}

type ZLMStreamNoneReaderData struct {
	APP           string `json:"app"`
	Stream        string `json:"stream"`
	Schema        string `json:"schema"`
	MediaServerID string `json:"mediaServerId"`
}

type ZlmStreamOnPlayData struct {
	App           string `json:"app"`
	HookIndex     int    `json:"hook_index"`
	ID            string `json:"id"`
	IP            string `json:"ip"`
	MediaServerID string `json:"mediaServerId"`
	Params        string `json:"params"`
	Port          int    `json:"port"`
	Protocol      string `json:"protocol"`
	Schema        string `json:"schema"`
	Stream        string `json:"stream"`
	Vhost         string `json:"vhost"`
}

type ZlmRtpServerTimeoutData struct {
	LocalPort     int    `json:"local_port"`
	ReUsePort     bool   `json:"re_use_port"`
	Ssrc          int    `json:"ssrc"`
	StreamID      string `json:"stream_id"`
	TCPMode       int    `json:"tcp_mode"`
	MediaServerID string `json:"mediaServerId"`
}

type ZlmStreamPublishData struct {
	MediaServerID string `json:"mediaServerId"` // 媒体服务器ID
	App           string `json:"app"`           // 应用名称
	ID            string `json:"id"`            // 流唯一标识符
	IP            string `json:"ip"`            // 客户端IP地址
	Params        string `json:"params"`        // 附加参数
	Port          int    `json:"port"`          // 客户端端口
	Schema        string `json:"schema"`        // 协议方案
	Protocol      string `json:"protocol"`      // 实际使用协议
	Stream        string `json:"stream"`        // 流名称
	VHost         string `json:"vhost"`         // 虚拟主机
}
