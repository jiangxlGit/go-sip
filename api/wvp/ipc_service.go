package wvp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"encoding/json"
	. "go-sip/common"
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"

	"net/url"
)

// ipc列表初始化到mysql数据库
func IpcInfoInit() {
	go func() {
		// 重试3次进行ipcInfo同步
		for i := 0; i < 3; i++ {
			if _, err := IpcInfoSync(); err != nil {
				Logger.Error("IpcInfoSync error", zap.Error(err))
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

// IpcInfo同步到mysql数据库
func IpcInfoSync() ([]*model.IpcInfo, error) {
	// 查询redis获取ipc列表
	device_ipc_info_map, err := redis_util.HGetAll_2(redis.DEVICE_IPC_INFO_KEY)
	if err != nil || device_ipc_info_map == nil || len(device_ipc_info_map) == 0 {
		Logger.Debug("未查询到ipc列表")
		return nil, err
	}
	// 遍历device_ipc_info_map
	var ipc_list []*model.IpcInfo
	for _, v := range device_ipc_info_map {
		ipc_info := model.IpcInfo{}
		// 反序列化
		err := json.Unmarshal([]byte(v), &ipc_info)
		if err != nil {
			Logger.Error("json反序列化失败", zap.Error(err))
			continue
		}
		// 查询ipc状态
		ipcStatus, err := redis_util.Get_2(fmt.Sprintf(redis.IPC_STATUS_KEY, ipc_info.IpcId))
		if err != nil || ipcStatus == "" {
			ipc_info.Status = "OFFLINE"
		} else {
			// 状态为在线,则更新心跳时间
			if ipcStatus == "ON" {
				ipc_info.LastHeartbeatTime = time.Now().Unix()
			}
			ipc_info.Status = ipcStatus
		}
		if ipc_info.InnerIP != "" && ipc_info.IpcIP == "" {
			ipc_info.IpcIP = ipc_info.InnerIP
		}
		// 如果ipc_info.LastHeartbeatTime距离当前时间超过24小时, 则置为错误状态
		if ipc_info.Status == "OFFLINE" && time.Now().Unix()-ipc_info.LastHeartbeatTime > 60 { // 改成1分钟为了测试
			ipc_info.Status = "ERROR"
		}
		ipc_list = append(ipc_list, &ipc_info)
	}
	err = dao.BatchUpsertIpcInfo(ipc_list)
	if err != nil {
		Logger.Error("批量更新ipc信息失败", zap.Error(err))
		return nil, err
	}
	for _, ipc_info := range ipc_list {
		redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, ipc_info.IpcId, ipc_info)
	}
	return ipc_list, nil
}

// @Summary		ipc控制接口
// @Router		/wvp/device/control [post]
// leftRight: 镜头左移右移 0:停止 1:左移 2:右移
// upDown: 镜头上移下移 0:停止 1:上移 2:下移
// inOut: 镜头放大缩小 0:停止 1:缩小 2:放大
// moveSpeed: 镜头移动速度，根据实际情况设置速度范围
// zoomSpeed: 镜头缩放速度，根据实际情况设置速度范围
func IpcControl(c *gin.Context) {
	var req model.IpcControlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误: "+err.Error())
		return
	}
	Logger.Info("ipc控制接口", zap.Any("req", req))

	params := url.Values{}
	params.Add("ipc_id", req.IpcId)
	params.Add("leftRight", strconv.Itoa(req.LeftRight))
	params.Add("upDown", strconv.Itoa(req.UpDown))
	params.Add("inOut", strconv.Itoa(req.InOut))
	params.Add("moveSpeed", strconv.Itoa(req.MoveSpeed))
	params.Add("zoomSpeed", strconv.Itoa(req.ZoomSpeed))

	response := WvpIpcGetRequestHandler(req.IpcId, DeviceControlURL, params)
	if response == nil || response.Data == nil {
		model.JsonResponseSysERR(c, "调用失败")
		return
	}
	model.JsonResponseSucc(c, response.Data)

}

// @Summary		分页查询ipc
// @Router		/wvp/ipc/list [post]
func GetIpcPage(c *gin.Context) {
	var req model.IpcPageQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误: "+err.Error())
		return
	}
	Logger.Info("分页查询ipc", zap.Any("req", req))
	// ipc列表查询并同步到数据库
	_, err := IpcInfoSync()
	if err != nil {
		Logger.Error("查询ipc列表失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询ipc列表失败")
		return
	}
	ipcList, err := dao.GetPageIpcInfo(req.Page, req.Size, req.DeviceID, req.GB)
	if err != nil {
		model.JsonResponseSysERR(c, "获取IPC列表失败")
		return
	}
	model.JsonResponsePageSucc(c, dao.GetIpcInfoTotal(req.DeviceID, req.GB), req.Page, req.Size, ipcList)
}

// @Summary		ipc回放视频时间列表（新）
// @Description	用来获取通道设备存储的可回放时间段列表，注意控制时间跨度，跨度越大，数据量越多，返回越慢，甚至会超时（最多10s）。
// @Router		/wvp/ipc/recordList [post]
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

// @Summary		国标ipc流重置
// @Router		/wvp/ipc/streamReset [post]
func IpcStreamReset(c *gin.Context) {
	var req model.IpcPushStreamResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误: "+err.Error())
		return
	}
	Logger.Info("非国标ipc推送流重置", zap.Any("req", req))

	params := url.Values{}
	params.Add("device_id", req.DeviceID)
	params.Add("ipc_id", req.IpcId)

	response := WvpDeviceGetRequestHandler(req.DeviceID, IpcStreamResetURL, params)
	if response == nil || response.Data == nil {
		model.JsonResponseSysERR(c, "调用失败")
		return
	}
	model.JsonResponseSucc(c, response.Data)

}

// @Summary		非国标ipc删除
// @Router		/wvp/ipc/delete [delete]
func IpcDelete(c *gin.Context) {
	ipcId := c.Query("ipcId")
	if ipcId == "" {
		model.JsonResponseSysERR(c, "ipcId不能为空")
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
	if ipcInfo.Status != "ERROR" {
		model.JsonResponseSysERR(c, "ipc不是错误状态,无法删除")
		return
	}
	// 先删除redis中的
	err = redis_util.HDel_2(redis.DEVICE_IPC_INFO_KEY, ipcId)
	if err != nil {
		model.JsonResponseSysERR(c, "删除失败")
		return
	}
	err = dao.DeleteIpcInfo(ipcId)
	if err != nil {
		model.JsonResponseSysERR(c, "删除失败")
		return
	}
	model.JsonResponseSucc(c, "删除成功")
}
