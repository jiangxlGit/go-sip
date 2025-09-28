package redis

var (
	// iot平台相关key
	IOT_OPEN_API_KEY      = "OPEN_API_KEY:API_SECRET:%s" // iot平台-open api secret
	IOT_DEVICE_REGION_KEY = "device_region"              // iot平台-设备id关联地区
	IOT_DEVICE_STORE_KEY  = "device_store"               // iot平台-设备id关联门店

	// wvp zlm相关
	WVP_ZLM_NODE_INFO            = "GOSIP_zlm_node"            // zlmDomain关联zlm节点信息
	WVP_REGION_RELATION_ZLM_INFO = "GOSIP_region_relation_zlm" // 地区关联zlm信息

	// open api相关key
	OPEN_API_KEY_NONCE = "GOSIP_open_api_nonce" // open api随机值

	// sip相关key
	SIP_SERVER_HOST            = "GOSIP_sip_server_host"            // 客户端sipId关联sip服务内网或公网地址
	SIP_SERVER_PUBLIC_TCP_HOST = "GOSIP_sip_server_public_tcp_host" // sipId关联grpc的tcp地址
	SIP_IPC                    = "GOSIP_%s_ipc"                     // sipId关联ipcId
	SIP_SERVER_LAST_SELECT_KEY = "GOSIP_sip_server_last_select"     // 客户端上次选择的sipId

	// 设备与摄像头相关key
	DEVICE_STATUS_KEY                  = "GOSIP_device_status:%s"                        // 设备在线离线状态
	IPC_STATUS_KEY                     = "GOSIP_ipc_status:%s"                           // ipc在线离线状态
	DEVICE_SIP_KEY                     = "GOSIP_device_sip"                              // 设备id关联sipId
	DEVICE_IPC_KEY                     = "GOSIP_device_ipc:%s"                           // 设备id关联ipcId
	DEVICE_IPC_INFO_KEY                = "GOSIP_device_ipc_info"                         // ipcId关联ipc信息
	IPC_VIDEO_PLAYBACK_LIST_KEY        = "GOSIP_ipc_video_playback_list:%s"              // ipcId+日期为key的视频回放列表
	DEVICE_IPC_VIDEO_PLAYBACK_LIST_KEY = "GOSIP_devcie_ipc_video_playback_list:%s:%s:%s" // deviceId+ipcId+模型类型为key的视频回放列表
	DEVICE_ZLM_KEY                     = "GOSIP_device_zlm"                              // 设备id关联的zlmDomain
	IPC_HEARTBEAT_INFO_KEY             = "GOSIP_ipc_heartbeat_info:%s"                   // ipc心跳信息

	// 合屏流对应ipcList
	MERGE_VIDEO_STREAM_IPC_LIST_KEY = "GOSIP_merge_video_stream_ipc"

	// ai模型类别自增值
	AI_MODEL_CATEGORY_SEQ_KEY       = "GOSIP_ai_model_category_seq:%s"
	AI_MODEL_STREAM_CLASSNAME_KEY   = "GOSIP_ai_model_stream_classname"
	AI_MODEL_DEVICE_RK_PLATFORM_KEY = "GOSIP_ai_model_device_rk_platform"

	// 非国标摄像头id自增值
	NOT_GB_IPC_ID_SEQ_KEY = "GOSIP_not_gb_ipc_id_seq"
	// 非国标摄像头id关联设备id
	NOT_GB_IPC_DEVICE = "GOSIP_not_gb_ipc_device"

	// 全局锁
	IPC_STATUS_SYNC_LOCK_KEY = "GOSIP_ipc_status_sync_lock"
)
