package zlm_api

import (
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"
	"strconv"
	"time"

	"fmt"

	"go.uber.org/zap"
)

type ZlmStartSendRtpReq struct {
	Vhost    string `json:"vhost"`
	StreamID string `json:"stream_id"`
	App      string `json:"app"`
	DstUrl   string `json:"dst_url"`
	DstPort  string `json:"dst_port"`
	Ssrc     string `json:"ssrc"`
	IsUdp    string `json:"is_udp"` // 1:udp active模式, 0:tcp active模式
}
type ZlmStopSendRtpReq struct {
	Vhost    string `json:"vhost"`
	StreamID string `json:"stream_id"`
	App      string `json:"app"`
	Ssrc     string `json:"ssrc"`
}

type ZlmGetRtpInfoReq struct {
	Vhost    string `json:"vhost"`
	StreamID string `json:"stream_id"`
	App      string `json:"app"`
}

type ZlmGetMediaListReq struct {
	Vhost    string `json:"vhost"`
	Schema   string `json:"schema"`
	StreamID string `json:"stream_id"`
	App      string `json:"app"`
}

type ZlmStartSendRtpResp struct {
	Code      int `json:"code"`
	LocalPort int `json:"local_port"`
}
type ZlmStopSendRtpResp struct {
	Code int `json:"code"`
}

type ZlmGetMediaListResp struct {
	Code int                       `json:"code"`
	Data []ZlmGetMediaListDataResp `json:"data"`
}

type ZlmGetRtpInfoResp struct {
	Code  int  `json:"code"`
	Exist bool `json:"exist"`
}
type ZlmGetMediaListDataResp struct {
	App        string                  `json:"app"`
	Stream     string                  `json:"stream"`
	Schema     string                  `json:"schema"`
	OriginType int                     `json:"originType"`
	Tracks     []ZlmGetMediaListTracks `json:"tracks"`
}
type ZlmGetMediaListTracks struct {
	Type    int     `json:"codec_type"`
	CodecID int     `json:"codec_id"`
	Height  int     `json:"height"`
	Width   int     `json:"width"`
	FPS     float64 `json:"fps"`
}

// Zlm 开始active模式发送rtp
// 作为zlm客户端，启动ps-rtp推流，支持rtp/udp方式；该接口支持rtsp/rtmp等协议转ps-rtp推流。第一次推流失败会直接返回错误，成功一次后，后续失败也将无限重试。
func ZlmStartSendRtp(url, secret string, req ZlmStartSendRtpReq) ZlmStartSendRtpResp {
	res := ZlmStartSendRtpResp{}
	reqStr := "/index/api/startSendRtp?secret=" + secret
	if req.StreamID != "" {
		reqStr += "&stream=" + req.StreamID
	}
	if req.App != "" {
		reqStr += "&app=" + req.App
	}
	if req.Vhost != "" {
		reqStr += "&vhost=" + req.Vhost
	}
	if req.DstUrl != "" {
		reqStr += "&dst_url=" + req.DstUrl
	}
	if req.DstPort != "" {
		reqStr += "&dst_port=" + req.DstPort
	}
	if req.IsUdp != "" {
		reqStr += "&is_udp=" + req.IsUdp
	}
	if req.Ssrc != "" {
		reqStr += "&ssrc=" + req.Ssrc
	}
	reqUrl := url + reqStr
	Logger.Info("开始active模式发送rtp", zap.Any("reqUrl", reqUrl))
	body, err := utils.GetRequest(reqUrl)
	if err != nil {
		Logger.Error("ZlmStartSendRtp fail, 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("ZlmStartSendRtp fail, 2", zap.Error(err))
		return res
	}
	if res.Code != 0 {
		Logger.Error("ZlmStartSendRtp fail, 3", zap.Error(err))
		return res
	}
	Logger.Info("ZlmStartSendRtp res", zap.Any("res", res))
	return res
}

// Zlm 停止GB28181 ps-rtp推流
func ZlmStopSendRtp(url, secret string, req ZlmStopSendRtpReq) ZlmStopSendRtpResp {
	res := ZlmStopSendRtpResp{}
	reqStr := "/index/api/stopSendRtp?secret=" + secret
	if req.StreamID != "" {
		reqStr += "&stream=" + req.StreamID
	}
	if req.App != "" {
		reqStr += "&app=" + req.App
	}
	if req.Vhost != "" {
		reqStr += "&vhost=" + req.Vhost
	}
	if req.Ssrc != "" {
		reqStr += "&ssrc=" + req.Ssrc
	}
	reqUrl := url + reqStr
	Logger.Info("停止rtp推流", zap.Any("reqUrl", reqUrl))
	body, err := utils.GetRequest(reqUrl)
	if err != nil {
		Logger.Error("ZlmStopSendRtp fail, 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("ZlmStopSendRtp fail, 2", zap.Error(err))
		return res
	}
	if res.Code != 0 {
		Logger.Error("ZlmStopSendRtp fail, 3", zap.Error(err))
		return res
	}
	Logger.Info("ZlmStopSendRtp res", zap.Any("res", res))
	return res
}

// Zlm 获取RTP流信息
func ZlmGetRtpInfo(url, secret string, req ZlmGetRtpInfoReq) ZlmGetRtpInfoResp {
	res := ZlmGetRtpInfoResp{}
	reqStr := "/index/api/getRtpInfo?secret=" + secret
	if req.StreamID != "" {
		reqStr += "&stream_id=" + req.StreamID
	}
	if req.App != "" {
		reqStr += "&app=" + req.App
	}
	if req.Vhost != "" {
		reqStr += "&vhost=" + req.Vhost
	}
	body, err := utils.GetRequest(url + reqStr)
	if err != nil {
		Logger.Error("get stream rtpInfo fail, 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("get stream rtpInfo fail, 2", zap.Error(err))
		return res
	}
	return res
}

// Zlm 获取流列表信息
func ZlmGetMediaList(url, secret string, req ZlmGetMediaListReq) ZlmGetMediaListResp {
	res := ZlmGetMediaListResp{}
	reqStr := "/index/api/getMediaList?secret=" + secret
	if req.StreamID != "" {
		reqStr += "&stream=" + req.StreamID
	}
	if req.App != "" {
		reqStr += "&app=" + req.App
	}
	if req.Schema != "" {
		reqStr += "&schema=" + req.Schema
	}
	if req.Vhost != "" {
		reqStr += "&vhost=" + req.Vhost
	}
	body, err := utils.GetRequest(url + reqStr)
	if err != nil {
		Logger.Error("get stream mediaList fail, 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("get stream mediaList fail, 2", zap.Error(err))
		return res
	}
	return res
}

var ZlmDeviceVFMap = map[int]string{
	0: "H264",
	1: "H265",
	2: "ACC",
	3: "G711A",
	4: "G711U",
}

func TransZlmDeviceVF(t int) string {
	if v, ok := ZlmDeviceVFMap[t]; ok {
		return v
	}
	return "undefind"
}

type RtpInfo struct {
	Code  int  `json:"code"`
	Exist bool `json:"exist"`
}

// 获取流在Zlm上的信息
func ZlmGetMediaInfo(url, secret, stream_id string) RtpInfo {
	res := RtpInfo{}
	body, err := utils.GetRequest(url + "/index/api/getRtpInfo?secret=" + secret + "&stream_id=" + stream_id)
	if err != nil {
		Logger.Error("get stream rtpInfo fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("get stream rtpInfo fail 2", zap.Error(err))
		return res
	}
	return res
}

// Zlm 关闭流
func ZlmCloseStreams(url, secret, stream_id string) {
	utils.GetRequest(url + "/index/api/close_streams?secret=" + secret + "&stream=" + stream_id + "&schema=rtsp&vhost=__defaultVhost__&app=rtp&force=1")
}

// Zlm 强制关闭所有流
func ZlmCloseAllStreams(url, secret string) {
	utils.GetRequest(url + "/index/api/close_streams?secret=" + secret + "&vhost=__defaultVhost__&force=1")
}

func ZlmStartRtpServer(url, secret, stream_id, app string, tcp_mode int) OpenRtpRsp {
	return ZlmOpenRtpServer(url, secret, stream_id, app, tcp_mode)
}

type OpenRtpRsp struct {
	Code int `json:"code"`
	Port int `json:"port"`
}

// /openRtpServer
func ZlmOpenRtpServer(url, secret, stream_id, app string, tcp_mode int) OpenRtpRsp {
	res := OpenRtpRsp{}
	// port: 接收端口，0则为随机端口
	// tcp_mode: 0 udp 模式，1 tcp 被动模式, 2 tcp 主动模式
	body, err := utils.GetRequest(url + "/index/api/openRtpServer?secret=" + secret + "&stream_id=" + stream_id + "&app=" + app + "&port=0" + "&tcp_mode=" + fmt.Sprintf("%d", tcp_mode))
	if err != nil {
		Logger.Error("open server rtp fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("open server rtp fail 2", zap.Error(err))
		return res
	}
	Logger.Info("open server rtp", zap.Any("res", res))
	return res
}

type CloseRtpRsp struct {
	Code int `json:"code"`
	Hit  int `json:"hit"`
}

// /openRtpServer
func ZlmCloseRtpServer(url, secret, stream_id string) CloseRtpRsp {
	res := CloseRtpRsp{}

	body, err := utils.GetRequest(url + "/index/api/closeRtpServer?secret=" + secret + "&stream_id=" + stream_id)
	if err != nil {
		Logger.Error("close server rtp fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("close server rtp fail 2", zap.Error(err))
		return res
	}
	return res
}

type StreamMergeInfoVO struct {
	Code int `json:"code"` // 返回码
}

// zlm视频流合屏
func ZlmMergeStream(dto model.StreamMergeInfoDTO, zlmInfo *model.ZlmInfo) StreamMergeInfoVO {

	res := StreamMergeInfoVO{}
	streamMergeConfigDTO := model.StreamMergeConfigDTO{
		GapV:   0,
		GapH:   0,
		Width:  320 * len(dto.IpcIdList),
		Height: 180,
		Row:    1,
		Col:    len(dto.IpcIdList),
		ID:     dto.StreamId,
	}

	streamMergeConfigDTO.Span = []int{}

	// 检查 ipcIdList 并赋值 spanEnd
	if len(dto.IpcIdList) == 0 {
		Logger.Error("ipcIdList 不能为空")
		return res
	}

	if zlmInfo == nil {
		Logger.Error("zlmInfo 不能为空")
		return res
	}

	// 遍历 ipcIdList
	urlList := [][]string{}
	urlSubList := []string{}
	for _, ipcId := range dto.IpcIdList {
		urlSubList = append(urlSubList, fmt.Sprintf("rtsp://%s:554/rtp/%s?originTypeStr=rtp_push&mode=1", zlmInfo.ZlmIp, ipcId))
	}
	urlList = append(urlList, urlSubList)
	streamMergeConfigDTO.URL = urlList
	// 使用http客户端调用 zlm 接口
	var zlmUrl string
	if dto.Type == 1 {
		// 合屏
		zlmUrl = fmt.Sprintf("%s/index/api/stack/start?secret=%s", zlmInfo.ZlmDomain, zlmInfo.ZlmSecret)
	} else if dto.Type == 2 {
		// 切屏
		zlmUrl = fmt.Sprintf("%s/index/api/stack/reset?secret=%s", zlmInfo.ZlmDomain, zlmInfo.ZlmSecret)
	}
	Logger.Info("zlm进行拼接流", zap.String("zlmUrl", zlmUrl), zap.Any("streamMergeConfigDTO", streamMergeConfigDTO))

	resp, err := utils.PostJSONRequest(zlmUrl, streamMergeConfigDTO)
	if err != nil || resp == nil {
		if dto.Type == 0 {
			Logger.Error("合屏失败", zap.Error(err))
		} else {
			Logger.Error("切屏失败", zap.Error(err))
		}
		return res
	}
	res.Code = 0
	return res
}

// 重置合屏，先关停合屏，延时2s再开启
func ZlmResetMergeStream(dto model.StreamMergeInfoDTO, zlmInfo *model.ZlmInfo) StreamMergeInfoVO {
	res := StreamMergeInfoVO{}
	// 停止合屏
	zlmUrl := fmt.Sprintf("%s/index/api/stack/stop?secret=%s&id=%s", zlmInfo.ZlmDomain, zlmInfo.ZlmSecret, dto.StreamId)
	Logger.Info("停止合屏", zap.String("zlmUrl", zlmUrl))
	result, err := utils.GetRequest(zlmUrl)
	if err != nil || result == nil {
		Logger.Error("停止合屏失败", zap.Error(err))
		return res
	}
	time.Sleep(2 * time.Second)
	return ZlmMergeStream(dto, zlmInfo)
}

type PauseRtpRsp struct {
	Code int `json:"code"`
}

func ZlmPauseRtpCheck(url, secret, stream_id string) PauseRtpRsp {
	res := PauseRtpRsp{}

	body, err := utils.GetRequest(url + "/index/api/pauseRtpCheck?secret=" + secret + "&app=rtp" + "&stream_id=" + stream_id)
	if err != nil {
		Logger.Error("pause server rtp check fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("pause server rtp check fail 2", zap.Error(err))
		return res
	}
	return res
}

func ZlmResumeRtpCheck(url, secret, stream_id string) PauseRtpRsp {
	res := PauseRtpRsp{}

	body, err := utils.GetRequest(url + "/index/api/resumeRtpCheck?secret=" + secret + "&app=rtp" + "&stream_id=" + stream_id)
	if err != nil {
		Logger.Error("resume server rtp check fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("resume server rtp fail 2", zap.Error(err))
		return res
	}
	return res
}

type OpenSendRtpRsp struct {
	Code      int `json:"code"`
	LocalPort int `json:"local_port"`
}

func ZlmStartSendRtpPassive(zlmDomain, secret, stream_id string) OpenSendRtpRsp {
	res := OpenSendRtpRsp{}
	url := zlmDomain + "/index/api/startSendRtpPassive?secret=" + secret +
		"&stream=" + stream_id + "&ssrc=1&app=broadcast&vhost=__defaultVhost__&only_audio=1&pt=8&use_ps=0&is_udp=0"
	body, err := utils.GetRequest(url)
	if err != nil {
		Logger.Error("start rtp passive fail 1", zap.Error(err))
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		Logger.Error("start rtp passive fail 2", zap.Error(err))
		return res
	}
	return res
}

type ZlmRecordRes struct {
	Code   int  `json:"code"`
	Result bool `json:"result"`
}

// Zlm 开始录制视频流
func ZlmStartRecord(zlmDomain, secret, stream_id, streamType string) ZlmRecordRes {
	res := ZlmRecordRes{}
	// 录像文件保存自定义根目录，为空则采用配置文件设置
	var customizedPath string
	if streamType == "" {
		customizedPath = "/userdata/gosip_playback_vedio/default"
	} else {
		customizedPath = "/userdata/gosip_playback_vedio/" + streamType
	}
	// mp4录像切片时间大小,单位秒，置0则采用配置项3600秒
	maxSecond := 1800
	// type: 0为hls，1为mp4
	respBody, err := utils.GetRequest(zlmDomain + "/index/api/startRecord?type=1&app=rtp&vhost=__defaultVhost__&secret=" + secret +
		"&stream=" + stream_id + "&customized_path=" + customizedPath + "&max_second=" + strconv.Itoa(maxSecond))
	if err != nil {
		Logger.Error("zlm start record GetRequest fail:", zap.Error(err))
		res.Code = 500
		res.Result = false
		return res
	}

	if err = utils.JSONDecode(respBody, &res); err != nil {
		Logger.Error("zlm start record json decode fail:", zap.Error(err))
		res.Code = 500
		res.Result = false
		return res
	}
	return res
}

// Zlm 停止录制
func ZlmStopRecord(zlmDomain, secret, stream_id string) ZlmRecordRes {
	res := ZlmRecordRes{}
	// type: 0为hls，1为mp4
	respBody, err := utils.GetRequest(zlmDomain + "/index/api/stopRecord?type=1&app=rtp&vhost=__defaultVhost__&secret=" + secret + "&stream=" + stream_id)
	if err != nil {
		Logger.Error("zlm start record GetRequest fail:", zap.Error(err))
		res.Code = 500
		res.Result = false
		return res
	}

	if err = utils.JSONDecode(respBody, &res); err != nil {
		Logger.Error("zlm start record json decode fail:", zap.Error(err))
		res.Code = 500
		res.Result = false
		return res
	}
	return res
}

type ZlmRecordStatusRes struct {
	Code   int  `json:"code"`
	Status bool `json:"status"` // false:未录制,true:正在录制
}

// zlm 获取流录制状态
func ZlmGetRecordStatus(zlmDomain, secret, stream_id string) ZlmRecordStatusRes {
	res := ZlmRecordStatusRes{}
	// type: 0为hls，1为mp4
	respBody, err := utils.GetRequest(zlmDomain + "/index/api/isRecording?type=1&app=rtp&vhost=__defaultVhost__&secret=" + secret + "&stream=" + stream_id)
	if err != nil {
		Logger.Error("zlm start record GetRequest fail:", zap.Error(err))
		res.Code = 500
		res.Status = false
		return res
	}

	if err = utils.JSONDecode(respBody, &res); err != nil {
		Logger.Error("zlm start record json decode fail:", zap.Error(err))
		res.Code = 500
		res.Status = false
		return res
	}
	return res
}
