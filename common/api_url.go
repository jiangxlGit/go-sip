package api

const (
	ZLMWebHookBaseURL = "/zlm/webhook"
	// 客户端相关接口
	ZLMWebHookClientURL   = "/client" + ZLMWebHookBaseURL + "/:method"
	AiModelStartRecordURL = "/aiModel/startRecord"
	AiModelStopRecordURL  = "/aiModel/stopRecord"

	// 播放接口
	PlayURL     = "/ipc/play"
	PlaybackURL = "/ipc/playback"
	PauseURL    = "/ipc/pause"
	ResumeURL   = "/ipc/resume"
	SpeedURL    = "/ipc/speed"
	SeekURL     = "/ipc/seek"
	StopURL     = "/ipc/stop"
	// 录像列表接口
	RecordsListURL = "/ipc/records"

	// 服务端ZLM Webhook接口
	ZLMWebHookServerURL = ZLMWebHookBaseURL + "/:method"
	// 设置设备语音音量
	DeviceControlURL = "/device/control"
	// 设置设备语音音量
	DevcieSetAudioVloumeURL = "/device/setAudioVolume"
	// 非国标设备推流重置
	IpcPushStreamResetURL = "/ipc/nogbPushStreamReset"
	// 国标设备推流重置
	IpcStreamResetURL = "/ipc/streamReset"
	// 设备OTA固件拉取升级
	DeviceOtaFirmwarePullURL = "/device/ota/firmwarePull"

	// 获取所有已启用的AI模型接口
	AiModelListURL = "/aiModel/list"

	// 对外获取sip服务信息接口
	OpenSipServerInfoURL = "/open/sip/getSipServerInfo"
	// 对外开放的录像列表接口
	OpenRecordsListURL = "/open/ipc/recordList"
	// 对外开放的ZLM信息接口
	OpenZLMInfoURL = "/open/ipc/getZlm"
	// 对外开放的回放速度接口
	OpenPlaybackSpeedURL = "/open/ipc/playbackSpeed"
	// 对外开放的暂停接口
	OpenPauseURL = "/open/ipc/playbackPause"
	// 对外开放的恢复接口
	OpenResumeURL = "/open/ipc/playbackResume"
	// 对外开放的拖动播放接口
	OpenSeekURL = "/open/ipc/playbackSeek"
	// 对外开放的停止接口
	OpenStopURL = "/open/ipc/playbackStop"
	// 对外开放的ZLM Webhook接口(不进行签名鉴权)
	OpenZLMWebHookURL = "/open_no_sign" + ZLMWebHookBaseURL + "/:method"
	// 对外设置设备语音音量
	OpenSetDevcieAudioVloumeURL = "/open/device/setAudioVolume"
	// 获取所有已启用的AI模型接口
	OpenAiModelListURL = "/open/aiModel/list"
	// 查询所有已启用的ai模型并且已关联设备的
	OpenAiModelRelationListURL = "/open/aiModelRelation/list"
	// 保存录制下的视频回放时间记录接口
	OpenIpcPlaybackRecordURL = "/open/ipc/playbackRecord"
	// 重置合屏流接口
	OpenIpcResetMergeStreamURL = "/open/ipc/resetMergeStream"
	// 获取所有非国标IPC接口
	OpenGetNotGbIpcListURL = "/open/ipc/notGbList"
	// 获取所有IPC接口
	OpenGetIpcListURL = "/open/ipc/list"
	// 更新非国标iIPC信息
	OpenGetNotGbIpcUpdateURL = "/open/ipc/notGbUpdate"
	// 设备推流
	OpenIpcPushStreamURL = "/open/ipc/pushStream"

	// 登录接口
	WvpAuthURL = "/login/auth"

	// 文件上传alioss
	WvpFileAliossUploadURL = "/wvp/file/alioss/upload"

	// wvp相关接口
	WvpIpcControlURL    = "/wvp/ipc/control"
	WvpIpcClarityURL    = "/wvp/ipc/clarity"
	WvpIpcListURL       = "/wvp/ipc/list"
	WvpIpcRecordListURL = "/wvp/ipc/recordList"
	WvpRecordsListURL   = "/wvp/ipc/records"
	WvpZLMInfoURL       = "/wvp/ipc/getZlm"
	WvpPlaybackSpeedURL = "/wvp/ipc/playbackSpeed"
	WvpPauseURL         = "/wvp/ipc/playbackPause"
	WvpResumeURL        = "/wvp/ipc/playbackResume"
	WvpSeekURL          = "/wvp/ipc/playbackSeek"
	WvpStopURL          = "/wvp/ipc/playbackStop"

	WvpGetZlmNodeListURL               = "/wvp/zlm/nodeList"
	WvpGetZlmNodeInfoURL               = "/wvp/zlm/nodeInfo/:deviceId"
	WvpAddZlmNodeURL                   = "/wvp/zlm/nodeAdd"
	WvpDeleteZlmNodeURL                = "/wvp/zlm/nodeDelete/:id"
	WvpUpdateZlmNodeURL                = "/wvp/zlm/nodeUpdate/:id"
	WvpUpdateZlmNodeStatusURL          = "/wvp/zlm/nodeStatusUpdate/:id"
	WvpGetZlmNodeRelationRegionListURL = "/wvp/zlm/nodeRelationRegionList/:nodeId"

	WvpGetZlmRegionListURL = "/wvp/zlm/regionList"
	WvpUpdateZlmRegionURL  = "/wvp/zlm/regionUpdate/:id"

	WvpGetIotDeviceListURL       = "/wvp/iotdevice/list"
	WvpIotDeviceListByAiModelURL = "/wvp/iotdevice/listByAiModel"

	WvpAiModelListURL         = "/wvp/aiModel/list"
	WvpAiModelCountURL        = "/wvp/aiModel/count"
	WvpAiModelAddURL          = "/wvp/aiModel/add"
	WvpAiModelUpdateURL       = "/wvp/aiModel/update/:id"
	WvpAiModelStatusUpdateURL = "/wvp/aiModel/statusUpdate/:id"
	WvpAiModelDeleteURL       = "/wvp/aiModel/delete/:id"
	WvpAiModelRelationURL     = "/wvp/aiModel/relation/:id"

	WvpDeviceRelationAiModelURL     = "/wvp/deviceAiModel/relation/:deviceId"
	WvpDeviceRelationManyAiModelURL = "/wvp/deviceAiModel/relationMany/:deviceId"
	WvpDeviceAiModelListURL         = "/wvp/deviceAiModel/list"
	WvpDeviceAiModelCountURL        = "/wvp/deviceAiModel/count"
	WvpDeviceAiModelDeleteURL       = "/wvp/deviceAiModel/delete/:relationId"
	WvpDeviceAiModelStatusUpdateURL = "/wvp/deviceAiModel/statusUpdate/:relationId"

	WvpAiModelCategorySaveURL   = "/wvp/aiModelCategory/add"
	WvpAiModelCategoryUpdateURL = "/wvp/aiModelCategory/update/:id"
	WvpAiModelCategoryDeleteURL = "/wvp/aiModelCategory/delete/:id"
	WvpAiModelCategoryListURL   = "/wvp/aiModelCategory/list"

	WvpIpcAiModelLabelsURL                  = "/wvp/ipcAiModel/labels"
	WvpIpcAiModelListURL                    = "/wvp/ipcAiModel/list"
	WvpIpcAiModelVideoRecordStatusUpdateURL = "/wvp/ipcAiModel/statusUpdate/:relationId"
	WvpIpcSaveAiModelRelationURL            = "/wvp/ipcAiModel/saveRelation"

	WvpIpcNotGbAddURL               = "/wvp/ipc/addNotGb"
	WvpIpcNotGbUpdateURL            = "/wvp/ipc/updateNotGb"
	WvpIpcAddOrUpdateNotGbConfigURL = "/wvp/ipc/addOrUpdateNotGbConfig"
	WvpIpcGetNotGbConfigURL         = "/wvp/ipc/getNotGbConfig"
	WvpIpcGetNotGbConfigListURL     = "/wvp/ipc/getNotGbConfigList"
	WvpIpcNotGbPushStreamResetURL   = "/wvp/ipc/nogbPushStreamReset"
	WvpIpcStreamResetURL            = "/wvp/ipc/streamReset"
	WvpIpcNotGbDeleteURL            = "/wvp/ipc/nogbDelete"
	WvpIpcDeleteURL                 = "/wvp/ipc/delete"
)
