package gateway

import (
	"encoding/json"
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_gateway_util"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"
	"go-sip/zlm_api"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		查询所有ipc
// @Router		/open/ipc/list [get]
func GetAllIpcList(c *gin.Context) {
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	ipcList, err := dao.GetAllIpcList(deviceId)
	if err != nil {
		model.JsonResponseSysERR(c, "获取IPC列表失败")
		return
	}
	model.JsonResponseSucc(c, ipcList)
}

// @Summary		根据设备id查询所有非国标ipc
// @Router		/open/ipc/notGbList [get]
func GetNotGbIpcList(c *gin.Context) {
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	Logger.Debug("非国标ipc列表查询", zap.Any("deviceId", deviceId))
	ipcList, err := dao.GetIpcList(deviceId, "yes")
	if err != nil {
		model.JsonResponseSysERR(c, "获取IPC列表失败")
		return
	}
	var iotNotGbIpcInfoList = []*model.IotNotGbIpcInfo{}
	for _, ipc := range ipcList {
		var iotNotGbIpcInfo = model.IotNotGbIpcInfo{}
		iotNotGbIpcInfo.IpcId = ipc.IpcId
		iotNotGbIpcInfo.IpcName = ipc.IpcName
		iotNotGbIpcInfo.InnerIP = ipc.InnerIP
		iotNotGbIpcInfo.Manufacturer = ipc.Manufacturer
		iotNotGbIpcInfo.Username = ipc.NogbUsername
		iotNotGbIpcInfo.Password = ipc.NogbPassword
		iotNotGbIpcInfo.Status = ipc.Status
		// 根据Manufacturer查询配置
		cfg, err := dao.GetNotGBConfigByManufacturer(ipc.Manufacturer)
		if err != nil || cfg == nil {
			Logger.Error("根据Manufacturer查询配置失败", zap.Error(err))
		} else {
			iotNotGbIpcInfo.RtspMainSuffix = cfg.RtspMainSuffix
			iotNotGbIpcInfo.RtspSubSuffix = cfg.RtspSubSuffix
		}
		iotNotGbIpcInfoList = append(iotNotGbIpcInfoList, &iotNotGbIpcInfo)
	}
	model.JsonResponseSucc(c, iotNotGbIpcInfoList)
}

// @Summary		更新非国标ipc
// @Router		/open/ipc/notGbUpdate [get]
func IpcNotGbInfoUpdate(c *gin.Context) {
	ipcId := c.Query("ipcId")
	status := c.Query("status")
	if ipcId == "" || status == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	if status != "ON" && status != "OFFLINE" && status != "ERROR" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	// 查询ipc是否存在
	ipcInfo, err := dao.GetIpcInfoByIpcId(ipcId)
	if err != nil {
		model.JsonResponseSysERR(c, "查询ipc失败")
		return
	}
	if ipcInfo == nil {
		model.JsonResponseSysERR(c, "ipc不存在")
		return
	}
	if ipcInfo.Status == "ERROR" && status == "OFFLINE" {
		model.JsonResponseSysERR(c, "ipc已处于错误状态,不能更新到离线状态")
	}
	redis_util.Set_2(fmt.Sprintf(redis.IPC_STATUS_KEY, ipcId), status, -1)
	model.JsonResponseSucc(c, "更新成功")
}

// @Summary		ipc回放视频时间列表
// @Description	用来获取通道设备存储的可回放时间段列表，注意控制时间跨度，跨度越大，数据量越多，返回越慢，甚至会超时（最多10s）。
// @Router		/open/ipc/recordList [post]
func IpcRecordsList(c *gin.Context) {

	var query model.IpcRecordListQueryReq
	if err := c.ShouldBindJSON(&query); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	start := query.Start
	end := query.End

	startStamp, err := strconv.ParseInt(start, 10, 64)
	if err != nil || startStamp <= 0 {
		model.JsonResponseSysERR(c, "开始时间格式错误")
		return
	}
	endStamp, err := strconv.ParseInt(end, 10, 64)
	if err != nil || endStamp <= 0 || endStamp <= startStamp {
		model.JsonResponseSysERR(c, "结束时间格式错误")
		return
	}
	// 如果结束时间大于当前时间，则取当前时间
	if endStamp > time.Now().Unix() {
		endStamp = time.Now().Unix()
	}

	// start和end都是秒级时间戳，转date
	startDate := time.Unix(startStamp, 0).Format("2006-01-02")
	endDate := time.Unix(endStamp, 0).Format("2006-01-02")
	// start和end必须是同一天
	if startDate != endDate {
		model.JsonResponseSysERR(c, "开始时间和结束时间必须在同一天")
		return
	}

	// 7天之前的录像不能查询
	if startStamp < time.Now().Add(-7*24*time.Hour).Unix() {
		model.JsonResponseSysERR(c, "开始时间不能小于7天")
		return
	}

	ipcId := query.IpcId

	var deviceId string
	// 根据ipc_id查询device_id
	if strings.HasPrefix(ipcId, "IPC") {
		deviceId, err = redis_util.HGet_2(redis.NOT_GB_IPC_DEVICE, ipcId)
		if err != nil || deviceId == "" {
			Logger.Warn("ipcId没有关联任何设备", zap.Any("ipcId", ipcId))
			model.JsonResponseSysERR(c, "未找到任何ipc")
			return
		}
	} else {
		device_ipc_info_str, err := redis_util.HGet_2(redis.DEVICE_IPC_INFO_KEY, ipcId)
		if err != nil || device_ipc_info_str == "" {
			Logger.Error("ipcId没有关联任何设备", zap.Error(err), zap.Any("ipcId", ipcId))
			model.JsonResponseSysERR(c, "ipcId没有关联任何设备")
			return
		}
		ipc_info := model.IpcInfo{}
		// 反序列化
		err = json.Unmarshal([]byte(device_ipc_info_str), &ipc_info)
		if err != nil {
			Logger.Error("json反序列化失败", zap.Error(err))
			model.JsonResponseSysERR(c, "json反序列化失败")
		}
		deviceId = ipc_info.DeviceID
	}

	var ipcRecordList []model.IpcPlaybackRecordData
	if query.Type == "all" {
		ipcRecordListSub, _ := redis_util.ScanPlaybackRecords(fmt.Sprintf(redis.DEVICE_IPC_VIDEO_PLAYBACK_LIST_KEY, deviceId, ipcId, "*"), start, end)
		ipcRecordList = append(ipcRecordList, ipcRecordListSub...)
	} else {
		typeArr := strings.Split(query.Type, ",")
		for _, typeStr := range typeArr {
			ipcRecordListSub, _ := redis_util.ZRangeByScore_2(fmt.Sprintf(redis.DEVICE_IPC_VIDEO_PLAYBACK_LIST_KEY, deviceId, ipcId, typeStr), start, end)
			ipcRecordList = append(ipcRecordList, ipcRecordListSub...)
		}
	}
	model.JsonResponseSucc(c, ipcRecordList)
}

// @Summary		记录ipc视频回放
// @Router		/open/ipc/playbackRecord [post]
func IpcPlaybackRecord(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("IpcPlaybackRecord read body error", zap.Error(err))
		model.JsonResponseSysERR(c, "read body error")
		return
	}
	zlmRecordMp4Data := model.ZLMRecordMp4Data{}
	if err := utils.JSONDecode(bodyBytes, &zlmRecordMp4Data); err != nil {
		Logger.Error("IpcPlaybackRecord bind json error", zap.Error(err))
		model.JsonResponseSysERR(c, "invalid request")
		return
	}
	streamId := zlmRecordMp4Data.Stream
	startTime := zlmRecordMp4Data.StartTime - 2 // 修正时间
	timeLen := zlmRecordMp4Data.TimeLen
	fileName := zlmRecordMp4Data.FileName
	filePath := zlmRecordMp4Data.FilePath
	fileOssDownloadUrl := zlmRecordMp4Data.FileOssDownloadUrl
	if streamId == "" || startTime == 0 || timeLen == 0 || fileName == "" || fileOssDownloadUrl == "" || filePath == "" {
		Logger.Error("关键参数为空")
		model.JsonResponseSysERR(c, "关键参数为空")
		return
	}
	var fileType = ""
	// 对fileName进行/分割
	filePathArr := strings.Split(filePath, "/")
	// 遍历fileNameArr中gosip_playback_vedio后面的值
	for i, name := range filePathArr {
		if name == "gosip_playback_vedio" {
			fileType = filePathArr[i+1]
			break
		}
	}
	if fileType == "" {
		Logger.Error("fileName格式错误")
		model.JsonResponseSysERR(c, "fileName格式错误")
		return
	}

	// timeLen向下取整
	timeLen = math.Floor(timeLen)
	if timeLen <= 0 {
		Logger.Error("时间长度必须大于0")
		model.JsonResponseSysERR(c, "时间长度必须大于0")
		return
	}
	// 计算结束时间戳
	endTime := startTime + int64(timeLen)

	stream_arr := strings.Split(streamId, "_")
	ipcId := stream_arr[0]
	deviceId := ""
	// 根据ipcId查询device_id
	if strings.HasPrefix(ipcId, "IPC") {
		deviceId, err = redis_util.HGet_2(redis.NOT_GB_IPC_DEVICE, ipcId)
		if err != nil || deviceId == "" {
			Logger.Warn("ipcId没有关联任何设备", zap.Any("ipcId", ipcId))
			model.JsonResponseSysERR(c, "未找到任何ipc")
			return
		}
	} else {
		device_ipc_info_str, err := redis_util.HGet_2(redis.DEVICE_IPC_INFO_KEY, ipcId)
		if err != nil || device_ipc_info_str == "" {
			Logger.Error("未找到任何ipc", zap.Error(err))
			model.JsonResponseSysERR(c, "未找到任何ipc")
			return
		} else {
			ipcInfo := model.IpcInfo{}
			// 反序列化
			err = json.Unmarshal([]byte(device_ipc_info_str), &ipcInfo)
			if err != nil {
				Logger.Error("json反序列化失败", zap.Error(err))
				model.JsonResponseSysERR(c, "json反序列化失败")
				return
			} else {
				deviceId = ipcInfo.DeviceID
			}
		}
	}
	// 根据streamId查询ai模型类别
	className, err := redis_util.HGet_2(redis.AI_MODEL_STREAM_CLASSNAME_KEY, streamId)
	if err != nil || className == "" {
		className = "person"
	}

	var ipcPlaybackRecordDataList []model.IpcPlaybackRecordData
	var ipcPlaybackRecordData = model.IpcPlaybackRecordData{}
	ipcPlaybackRecordData.StartTime = startTime
	ipcPlaybackRecordData.EndTime = endTime
	ipcPlaybackRecordData.FileUrl = fileOssDownloadUrl
	ipcPlaybackRecordData.AiModelType = className
	ipcPlaybackRecordDataList = append(ipcPlaybackRecordDataList, ipcPlaybackRecordData)

	// 存入redis
	err = redis_util.ZAdd_2(fmt.Sprintf(redis.DEVICE_IPC_VIDEO_PLAYBACK_LIST_KEY, deviceId, ipcId, fileType), ipcPlaybackRecordDataList, time.Hour*24*7)
	if err != nil {
		Logger.Error("存入redis失败", zap.Error(err))
		model.JsonResponseSysERR(c, "存入redis失败")
		return
	}
	model.JsonResponseSucc(c, "成功")
}

// @Summary		重置合屏流
// @Router		/open/ipc/resetMergeStream [post]
func IpcResetMergeStream(c *gin.Context) {

	var req model.IpcResetMergeStreamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	redisZlmInfo, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, req.ZlmDomain)
	if err != nil || redisZlmInfo == "" {
		model.JsonResponseSysERR(c, "查询ZLM节点信息失败")
		return
	}

	// 反序列化 JSON 字符串
	var zlmInfo model.ZlmInfo
	err = json.Unmarshal([]byte(redisZlmInfo), &zlmInfo)
	if err != nil {
		model.JsonResponseSysERR(c, "参数格式错误，json反序列化失败")
		return
	}

	device_id := req.DeviceId
	ipcListStr := req.IpcList
	// 对sub_ipc进行分割
	sub_ipc_list := strings.Split(ipcListStr, "_")

	stream_id_list := make([]string, len(sub_ipc_list))
	// 点播所有拼接的摄像头
	for i, ipc_id := range sub_ipc_list {
		// 主屏使用高清，其他的使用标清
		stream_id := fmt.Sprintf("%s_0", ipc_id)
		stream_id_list[i] = stream_id
	}
	// 进行切屏
	dto := model.StreamMergeInfoDTO{
		DeviceId:  device_id,
		IpcIdList: stream_id_list,
		StreamId:  string(device_id),
		Type:      2,
	}

	// 是否需要重置合屏流
	isReset := false

	// 判断sub_stream流是否在线，在线则进行切屏，不在线则进行合屏
	var zlmGetMediaListReq = zlm_api.ZlmGetMediaListReq{}
	zlmGetMediaListReq.App = "rtp"
	zlmGetMediaListReq.Vhost = "__defaultVhost__"
	zlmGetMediaListReq.Schema = "rtsp"
	zlmGetMediaListReq.StreamID = device_id
	resp := zlm_api.ZlmGetMediaList(zlmInfo.ZlmDomain, zlmInfo.ZlmSecret, zlmGetMediaListReq)
	if resp.Code == 0 && len(resp.Data) > 0 {
		// 流存在, 且ipc列表个数与id不变，则切屏，否则重置合屏流
		redisIpcListStr, _ := redis_util.HGet_2(redis.MERGE_VIDEO_STREAM_IPC_LIST_KEY, device_id)
		if redisIpcListStr != "" {
			redis_sub_ipc_list := strings.Split(redisIpcListStr, "_")
			if !utils.EqualStringSliceSet(sub_ipc_list, redis_sub_ipc_list) {
				dto.Type = 1 // 合屏
				isReset = true
			}
		}
	} else {
		// 流不存在，直接合屏
		dto.Type = 1 // 合屏
	}
	if !isReset {
		resp2 := zlm_api.ZlmMergeStream(dto, &zlmInfo)
		if resp2.Code != 0 {
			Logger.Error("重置合屏流失败", zap.String("stream_id", device_id), zap.Any("resp", resp2))
			model.JsonResponseSysERR(c, "重置合屏流失败")
			return
		}
	} else {
		resp2 := zlm_api.ZlmResetMergeStream(dto, &zlmInfo)
		if resp2.Code != 0 {
			Logger.Error("重置合屏流失败", zap.String("stream_id", device_id), zap.Any("resp", resp2))
			model.JsonResponseSysERR(c, "重置合屏流失败")
			return
		}
	}
	redis_util.HSet_2(redis.MERGE_VIDEO_STREAM_IPC_LIST_KEY, device_id, ipcListStr)
	model.JsonResponseSucc(c, "成功")

}
