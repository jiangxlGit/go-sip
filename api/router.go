package api

import (
	gapi "go-sip/api/gateway"
	"go-sip/api/middleware"
	sapi "go-sip/api/s"
	wvpapi "go-sip/api/wvp"
	. "go-sip/common"

	"github.com/gin-gonic/gin"
)

// 公共初始化
func Init(r *gin.Engine) {
	// 中间件
	r.Use(middleware.ApiAuth)
	r.Use(middleware.CORS())
}

// 服务端api初始化
func ServerApiInit(r *gin.Engine) {
	Init(r)

	// 播放类接口
	{
		r.GET(PlayURL, sapi.Play)
		r.GET(PlaybackURL, sapi.Playback)
		r.GET(PauseURL, sapi.Pause)
		r.GET(ResumeURL, sapi.Resume)
		r.GET(SpeedURL, sapi.Speed)
		r.GET(SeekURL, sapi.Seek)
		r.GET(StopURL, sapi.Stop)
	}
	// 录像类
	{
		r.GET(RecordsListURL, sapi.RecordsList)
	}
	// server zlm webhook
	{
		r.POST(ZLMWebHookServerURL, sapi.ZLMWebHook)
	}

	// 设备控制类
	{
		r.GET(DeviceControlURL, sapi.DeviceControl)
		r.GET(DevcieSetAudioVloumeURL, sapi.SetDevcieAudioVloume)
		r.GET(IpcPushStreamResetURL, sapi.IpcNoGbPushStreamReset)
		r.GET(IpcStreamResetURL, sapi.IpcStreamReset)
		r.GET(DeviceOtaFirmwarePullURL, sapi.OTAFirmwarePull)
	}

}

// 网关api初始化
func GatewayApiInit(r *gin.Engine) {
	Init(r)

	// 对外开放接口
	{
		// sip服务相关接口
		r.POST(OpenSipServerInfoURL, gapi.GetSipServerInfo)

		// zlm相关接口
		r.POST(OpenZLMInfoURL, gapi.ZLMInfo)

		// hook相关接口
		r.POST(OpenZLMWebHookURL, gapi.ZLMWebHook)

		// 中控设备相关接口
		r.POST(OpenSetDevcieAudioVloumeURL, gapi.DevcieSetAudioVloume)

		// AI模型相关接口
		r.GET(OpenAiModelListURL, gapi.GetAiModelList)
		r.GET(OpenAiModelRelationListURL, gapi.GetAiModelRelationList)

		// ipc相关接口
		r.GET(OpenGetIpcListURL, gapi.GetAllIpcList)
		r.GET(OpenGetNotGbIpcListURL, gapi.GetNotGbIpcList)
		r.GET(OpenGetNotGbIpcUpdateURL, gapi.IpcNotGbInfoUpdate)
		r.POST(OpenRecordsListURL, gapi.IpcRecordsList)
		r.POST(OpenIpcPlaybackRecordURL, gapi.IpcPlaybackRecord)
		r.POST(OpenIpcResetMergeStreamURL, gapi.IpcResetMergeStream)
	}

}

// wvp api初始化
func WvpApiInit(r *gin.Engine) {
	Init(r)

	// go-wvp平台相关接口
	{
		r.POST(WvpAuthURL, middleware.GetAuth)

		r.POST(WvpFileAliossUploadURL, middleware.FileUploadAliOSSHandler)

		r.POST(WvpIpcControlURL, wvpapi.IpcControl)
		r.POST(WvpIpcListURL, wvpapi.GetIpcPage)
		// 视频回放相关接口
		r.POST(WvpIpcRecordListURL, wvpapi.IpcRecordsList)

		r.GET(WvpGetZlmNodeListURL, wvpapi.GetZlmNodeInfoList)
		r.GET(WvpGetZlmNodeInfoURL, wvpapi.GetZlmNodeInfo)
		r.POST(WvpAddZlmNodeURL, wvpapi.AddZlmNode)
		r.DELETE(WvpDeleteZlmNodeURL, wvpapi.ZlmNodeDelete)
		r.PUT(WvpUpdateZlmNodeURL, wvpapi.UpdateZlmNodeInfo)
		r.GET(WvpUpdateZlmNodeStatusURL, wvpapi.UpdateZlmNodeStatus)
		r.GET(WvpGetZlmNodeRelationRegionListURL, wvpapi.ZlmNodeRelationRegionList)

		r.GET(WvpGetZlmRegionListURL, wvpapi.ZlmNodeRegionInfoList)
		r.POST(WvpUpdateZlmRegionURL, wvpapi.ZlmNodeRegionUpdate)

		r.GET(WvpGetIotDeviceListURL, wvpapi.GetIotDeviceList)
		r.POST(WvpIotDeviceListByAiModelURL, wvpapi.GetIotDeviceListByAiModel)

		r.POST(WvpAiModelListURL, wvpapi.QueryAiModelList)
		r.GET(WvpAiModelCountURL, wvpapi.GetAiModelCount)
		r.POST(WvpAiModelAddURL, wvpapi.AddAiModel)
		r.PUT(WvpAiModelUpdateURL, wvpapi.UpdateAiModel)
		r.PUT(WvpAiModelStatusUpdateURL, wvpapi.UpdateAiModelStatus)
		r.DELETE(WvpAiModelDeleteURL, wvpapi.DeleteAiModel)
		r.POST(WvpAiModelRelationURL, wvpapi.AddAiModelRelation)

		r.GET(WvpDeviceRelationAiModelURL, wvpapi.DeviceRelationAiModel)
		r.POST(WvpDeviceRelationManyAiModelURL, wvpapi.DeviceRelationManyAiModel)
		r.POST(WvpDeviceAiModelListURL, wvpapi.GetDeviceAiModelList)
		r.GET(WvpDeviceAiModelCountURL, wvpapi.GetDeviceAiModelCount)
		r.DELETE(WvpDeviceAiModelDeleteURL, wvpapi.DeleteDeviceAiModelRelation)
		r.PUT(WvpDeviceAiModelStatusUpdateURL, wvpapi.UpdateAiModelRelationStatus)

		r.POST(WvpAiModelCategorySaveURL, wvpapi.AddAiModelCategory)
		r.PUT(WvpAiModelCategoryUpdateURL, wvpapi.UpdateAiModelCategory)
		r.DELETE(WvpAiModelCategoryDeleteURL, wvpapi.DeleteAiModelCategory)
		r.GET(WvpAiModelCategoryListURL, wvpapi.QueryAiModelCategoryList)

		r.POST(WvpIpcAiModelLabelsURL, wvpapi.GetIpcAiModelLabels)
		r.POST(WvpIpcAiModelListURL, wvpapi.GetIpcAiModelList)
		r.PUT(WvpIpcAiModelVideoRecordStatusUpdateURL, wvpapi.UpdateAiModelVideoRecordStatus)
		r.POST(WvpIpcSaveAiModelRelationURL, wvpapi.SaveIpcAiModelRelation)

		r.POST(WvpIpcNotGbAddURL, wvpapi.AddNotGbIpcInfo)
		r.PUT(WvpIpcNotGbUpdateURL, wvpapi.UpdateNotGBIpcInfo)
		r.POST(WvpIpcAddOrUpdateNotGbConfigURL, wvpapi.AddOrUpdateNotGbConfig)
		r.GET(WvpIpcGetNotGbConfigURL, wvpapi.GetNotGbConfig)
		r.GET(WvpIpcGetNotGbConfigListURL, wvpapi.GetNotGbConfigList)
		r.POST(WvpIpcNotGbPushStreamResetURL, wvpapi.IpcPushStreamReset)
		r.POST(WvpIpcStreamResetURL, wvpapi.IpcStreamReset)
		r.DELETE(WvpIpcNotGbDeleteURL, wvpapi.IpcNotGbDelete)
		r.DELETE(WvpIpcDeleteURL, wvpapi.IpcDelete)

	}
}
