package wvp

import (
	"fmt"
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 查询iot设备列表(包含摄像头总数)
// @Router /wvp/iotdevice/list [get]
func GetIotDeviceList(c *gin.Context) {
	iotDeviceId := c.Query("id")
	// 查询mysql
	iotDeviceList, err := dao.GetIotDeviceList(iotDeviceId)
	if err != nil {
		Logger.Error("GetIotDeviceList error", zap.Error(err))
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	// 遍历iot设备列表，同步实时状态
	for _, device := range iotDeviceList {
		// 查看redis设备状态
		val, err := redis_util.Get_2(fmt.Sprintf(redis.DEVICE_STATUS_KEY, device.IotDeviceID))
		if err != nil || val == "" {
			Logger.Error("GetIotDeviceList error", zap.Error(err))
			device.State = "offline"
		} else {
			// 如果设备离线，则摄像头全部更新成离线
			if val == "offline" {
				ipcList, err := dao.GetAllNogbIpcList(device.IotDeviceID)
				if err != nil {
					Logger.Error("获取ipc信息失败", zap.Any("deviceId", device.IotDeviceID), zap.Error(err))
				} else {
					for _, ipcInfo := range ipcList {
						ipcInfo.Status = "OFFLINE"
						dao.UpdateIpcInfoSelective(&ipcInfo)
						redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, ipcInfo.IpcId, ipcInfo)
					}
				}
			}
			device.State = val
		}
	}
	model.JsonResponseSucc(c, iotDeviceList)
}

// @Summary 根据AI模型id分页查询中控设备列表
// @Router /wvp/iotdevice/listByAiModel [post]
func GetIotDeviceListByAiModel(c *gin.Context) {
	var pageReq model.IotDeviceListByAiModelPageQuery
	if err := c.ShouldBindJSON(&pageReq); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	var aiModelId = pageReq.AiModelId
	aiModel, err := dao.GetAiModelByID(aiModelId)
	if err != nil || aiModel == nil {
		Logger.Error("AI模型不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型不存在")
		return
	}
	list, err := dao.GetIotDevicePageByAiModel(pageReq.AiModelId, pageReq.Page, pageReq.Size)
	if err != nil {
		Logger.Error("GetUnboundIotDeviceList error", zap.Error(err))
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	model.JsonResponsePageSucc(c, dao.GetIotDeviceCountByAiModel(aiModelId), pageReq.Page, pageReq.Size, list)
}
