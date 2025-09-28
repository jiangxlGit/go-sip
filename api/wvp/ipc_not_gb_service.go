package wvp

import (
	"fmt"
	. "go-sip/common"
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"

	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		新增非国标ipc
// @Router		/wvp/ipc/addNotGb [post]
func AddNotGbIpcInfo(c *gin.Context) {
	var req model.IpcInfoNotGbAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	ipcIdSeq, err := redis_util.Get_2(redis.NOT_GB_IPC_ID_SEQ_KEY)
	if err != nil || ipcIdSeq == "" {
		ipcIdSeq = "1"
	} else {
		ipcIdSeqInt, err := strconv.Atoi(ipcIdSeq)
		if err != nil || ipcIdSeqInt < 1 {
			ipcIdSeq = "1"
		} else {
			ipcIdSeq = strconv.Itoa(ipcIdSeqInt + 1)
		}
	}
	redis_util.Set_2(redis.NOT_GB_IPC_ID_SEQ_KEY, ipcIdSeq, -1)
	ipcIdInt, _ := strconv.Atoi(ipcIdSeq)
	ipcId := fmt.Sprintf("IPC%010d", ipcIdInt)
	ipcInfoDB, err := dao.GetIpcInfoByIpcId(ipcId)
	if err != nil || ipcInfoDB != nil {
		model.JsonResponseSysERR(c, "创建非国标ipc失败")
		return
	}

	manufacturer := req.Manufacturer
	notGbConifg, err := dao.GetNotGBConfigByManufacturer(manufacturer)
	if err != nil {
		model.JsonResponseSysERR(c, "创建非国标ipc失败")
		return
	}
	if notGbConifg == nil {
		model.JsonResponseSysERR(c, "创建非国标ipc失败, 非国标配置不存在")
		return
	}

	var ipcInfo = model.IpcInfo{}
	ipcInfo.DeviceID = req.DeviceId
	ipcInfo.IpcIP = req.InnerIP
	ipcInfo.IpcId = ipcId
	ipcInfo.IpcName = req.IpcName
	ipcInfo.InnerIP = req.InnerIP
	ipcInfo.Manufacturer = req.Manufacturer
	ipcInfo.NogbUsername = req.NogbUsername
	ipcInfo.NogbPassword = req.NogbPassword
	ipcInfo.SipId = "-"
	ipcInfo.Transport = "RTSP"
	ipcInfo.Status = "OFFLINE"
	ipcInfo.ActiveTime = time.Now().Unix()

	err = dao.CreateIpcInfo(&ipcInfo)
	if err != nil {
		model.JsonResponseSysERR(c, "创建非国标ipc失败")
		return
	}
	// 存入redis
	redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, ipcId, ipcInfo)
	redis_util.HSet_2(redis.NOT_GB_IPC_DEVICE, ipcId, req.DeviceId)
	model.JsonResponseSucc(c, ipcInfo)
}

// @Summary 更新非国标ipc信息
// @Router /wvp/ipc/updateNotGb [put]
func UpdateNotGBIpcInfo(c *gin.Context) {
	var req model.NotGbIpcInfoUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	// 查询ipc是否存在
	ipcInfo, err := dao.GetIpcInfoByIpcId(req.IpcId)
	if err != nil {
		model.JsonResponseSysERR(c, "查询ipc失败")
		return
	}
	if ipcInfo == nil {
		model.JsonResponseSysERR(c, "ipc不存在")
		return
	}
	// 国标设备不能更新
	if ipcInfo.InnerIP == "" {
		model.JsonResponseSysERR(c, "国标设备不能手动更新")
		return
	}
	ipcInfo.IpcName = req.IpcName
	ipcInfo.IpcIP = req.InnerIP
	ipcInfo.InnerIP = req.InnerIP
	ipcInfo.NogbUsername = req.NogbUsername
	ipcInfo.NogbPassword = req.NogbPassword
	result := dao.UpdateIpcInfoSelective(ipcInfo)
	if result != nil {
		model.JsonResponseSysERR(c, "更新失败")
		return
	}
	redis_util.HSetStruct_2(redis.DEVICE_IPC_INFO_KEY, req.IpcId, ipcInfo)
	model.JsonResponseSucc(c, ipcInfo)
}

// @Summary 新增或更新非国标ipc配置信息
// @Route /wvp/ipc/addOrUpdateNotGbConfig [post]
func AddOrUpdateNotGbConfig(c *gin.Context) {
	var req model.NotGBConfigAddOrUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	manufacturer := req.Manufacturer
	notGbConifg, err := dao.GetNotGBConfigByManufacturer(manufacturer)
	if err != nil {
		model.JsonResponseSysERR(c, "创建非国标ipc配置失败")
		return
	}
	if notGbConifg == nil {
		notGbConfigNew := model.FormNotGbConfig(&req)
		id, err := dao.CreateNotGBConfig(notGbConfigNew)
		if err != nil {
			Logger.Error("非国标ipc配置创建失败", zap.Error(err))
			model.JsonResponseSysERR(c, "非国标ipc配置创建失败")
			return
		}
		notGbConfigNew.ID = id
		model.JsonResponseSucc(c, notGbConfigNew)
	} else {
		if req.IpcId != "" {
			ipcInfo, _ := dao.GetIpcInfoByIpcId(req.IpcId)
			if ipcInfo.InnerIP == "" {
				model.JsonResponseSysERR(c, "国标ipc无法进行配置")
				return
			}
			if ipcInfo != nil && ipcInfo.Manufacturer != req.Manufacturer {
				ipcInfo.Manufacturer = req.Manufacturer
				dao.UpdateIpcInfoSelective(ipcInfo)
			}
		}
		notGbConifg.RtspMainSuffix = req.RtspMainSuffix
		notGbConifg.RtspSubSuffix = req.RtspSubSuffix
		err := dao.UpdateNotGBConfig(notGbConifg)
		if err != nil {
			Logger.Error("非国标ipc配置更新失败", zap.Error(err))
			model.JsonResponseSysERR(c, "非国标ipc配置更新失败")
			return
		}
		model.JsonResponseSucc(c, notGbConifg)
	}

}

// @Summary 查询非国标ipc配置
// @Route /wvp/ipc/getNotGbConfig [get]
func GetNotGbConfig(c *gin.Context) {
	manufacturer := c.Query("manufacturer")
	if manufacturer == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	notGbConfig, err := dao.GetNotGBConfigByManufacturer(manufacturer)
	if err != nil {
		model.JsonResponseSysERR(c, "非国标ipc配置查询失败")
		return
	}
	model.JsonResponseSucc(c, notGbConfig)

}

// @Summary 查询非国标ipc配置列表
// @Route /wvp/ipc/getNotGbConfigList [get]
func GetNotGbConfigList(c *gin.Context) {
	notGbConfigList, err := dao.GetAllNotGBConfigs()
	if err != nil {
		model.JsonResponseSysERR(c, "非国标ipc配置列表查询失败")
		return
	}
	model.JsonResponseSucc(c, notGbConfigList)
}

// @Summary		非国标ipc推送流重置
// @Router		/wvp/ipc/nogbPushStreamReset [post]
func IpcPushStreamReset(c *gin.Context) {
	var req model.IpcPushStreamResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误: "+err.Error())
		return
	}
	Logger.Info("非国标ipc推送流重置", zap.Any("req", req))

	params := url.Values{}
	params.Add("device_id", req.DeviceID)
	params.Add("ipc_id", req.IpcId)

	response := WvpDeviceGetRequestHandler(req.DeviceID, IpcPushStreamResetURL, params)
	if response == nil || response.Data == nil {
		model.JsonResponseSysERR(c, "调用失败")
		return
	}
	model.JsonResponseSucc(c, response.Data)

}

// @Summary		非国标ipc删除
// @Router		/wvp/ipc/nogbDelete [delete]
func IpcNotGbDelete(c *gin.Context) {
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
