package wvp

import (
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary ipc关联ai模型标签列表查询
// @Router /wvp/ipcAiModel/labels [post]
func GetIpcAiModelLabels(c *gin.Context) {
	var req model.IpcAiModelLabelsQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	ipcAiModelRelationList, err := dao.QueryIpcAiModelRelations(req.IpcID, "")
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}

	var ipcRelationAiModelInfoList []*model.IpcRelationAiModelLabelsRep
	for _, ipcAiModelRelation := range ipcAiModelRelationList {
		aiModels, _ := dao.GetAiModelByID(ipcAiModelRelation.AiModelID)
		if len(aiModels) > 0 {
			aiModel := aiModels[0]
			modelLabels := aiModel.ModelLabels
			var modelLabelInfoList = []*model.ModelLabelInfo{}
			err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
			if err != nil {
				continue
			}
			for _, modelLabel := range modelLabelInfoList {
				var label model.IpcRelationAiModelLabelsRep
				label.LabelName = modelLabel.LabelName
				label.IsDefault = aiModel.IsDefault
				ipcRelationAiModelInfoList = append(ipcRelationAiModelInfoList, &label)
			}
		}
	}
	// 根据设备id查询rk平台
	rkPlatform, err := redis_util.HGet_2(redis.AI_MODEL_DEVICE_RK_PLATFORM_KEY, req.DeviceID)
	if err != nil {
		rkPlatform = ""
	}
	// 查询默认模型
	defaultAiModelList, err := dao.GetDefaultAiModelList(rkPlatform)
	if err == nil {
		for _, defaultAiModel := range defaultAiModelList {
			modelLabels := defaultAiModel.ModelLabels
			var modelLabelInfoList = []*model.ModelLabelInfo{}
			err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
			if err != nil {
				continue
			}
			for _, modelLabel := range modelLabelInfoList {
				var label model.IpcRelationAiModelLabelsRep
				label.LabelName = modelLabel.LabelName
				label.IsDefault = 1
				ipcRelationAiModelInfoList = append(ipcRelationAiModelInfoList, &label)
			}
		}
	}

	model.JsonResponseSucc(c, ipcRelationAiModelInfoList)
}

// @Summary ipc关联ai模型列表查询
// @Router /wvp/ipcAiModel/list [post]
func GetIpcAiModelList(c *gin.Context) {
	var req model.IpcAiModelListQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	// 根据设备id查询rk平台
	rkPlatform, err := redis_util.HGet_2(redis.AI_MODEL_DEVICE_RK_PLATFORM_KEY, req.DeviceID)
	if err != nil {
		rkPlatform = ""
	}

	// 查询已启用ai模型列表
	aiModelList, err := dao.GetIpcAiModelList(req.AiModelID, req.AiModelName, rkPlatform)
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}

	var ipcRelationAiModelInfoList []*model.IpcRelationAiModelInfoRep
	// 遍历ai模型列表，查询每个模型关联的信息
	for _, aiModel := range aiModelList {
		var info model.IpcRelationAiModelInfoRep
		info.DeviceID = req.DeviceID
		info.IpcID = req.IpcID
		info.AiModelId = aiModel.ID
		info.AiModelName = aiModel.Name
		info.MianCategory = aiModel.MianCategory
		info.SubCategory = aiModel.SubCategory
		info.MianCategoryName = aiModel.MianCategoryName
		info.SubCategoryName = aiModel.SubCategoryName
		info.ModelVersion = aiModel.ModelVersion
		info.ModelPlatform = aiModel.ModelPlatform

		modelLabels := aiModel.ModelLabels
		var modelLabelInfoList = []*model.ModelLabelInfo{}
		err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
		if err != nil {
			continue
		}
		info.ModelLabelList = modelLabelInfoList

		aiModelRelationList, err := dao.QueryIpcAiModelRelations(req.IpcID, aiModel.ID)
		if err != nil || len(aiModelRelationList) == 0 {
			info.RelationId = -1
			info.RelationEnable = "no"
			info.VideoRecordStatus = 0
		} else {
			aiModelRelation := aiModelRelationList[0]
			info.RelationId = aiModelRelation.ID
			info.RelationEnable = "yes"
			info.VideoRecordStatus = aiModelRelation.Status
			info.AiModelConfidence = aiModelRelation.AiModelConfidence
		}
		info.IsDefault = aiModel.IsDefault

		ipcRelationAiModelInfoList = append(ipcRelationAiModelInfoList, &info)
	}

	model.JsonResponseSucc(c, ipcRelationAiModelInfoList)
}

// @Summary 启用/禁用视频录制
// @Router /wvp/ipcAiModel/statusUpdate/{relationId} [put]
func UpdateAiModelVideoRecordStatus(c *gin.Context) {
	relationId := c.Param("relationId")
	statusStr := c.Query("status")
	if statusStr != "0" && statusStr != "1" {
		model.JsonResponseSysERR(c, "参数status错误（必须为0或1）")
		return
	}

	aiModelRelation, err := dao.GetIpcAiModelRelationByID(relationId)
	if err != nil || aiModelRelation == nil {
		Logger.Error("AI模型关联不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型关联不存在")
		return
	}

	if statusStr == "0" {
		aiModelRelation.Status = 0
	} else {
		aiModelRelation.Status = 1
	}

	err = dao.UpdateIpcAiModelRelation(aiModelRelation)
	if err != nil {
		Logger.Error("AI模型关联状态更新失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型关联状态更新失败")
		return
	}

	model.JsonResponseSucc(c, aiModelRelation)
}

// @Summary 保存ipc关联ai模型
// @Router /wvp/ipcAiModel/saveRelation [post]
func SaveIpcAiModelRelation(c *gin.Context) {
	var req model.IpcAiModelRelationReqData
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 根据设备id查询rk平台
	rkPlatform, err := redis_util.HGet_2(redis.AI_MODEL_DEVICE_RK_PLATFORM_KEY, req.DeviceID)
	if err != nil {
		rkPlatform = ""
	}

	// 查询所有已启用的ai模型
	aiModelList, err := dao.GetAiModelListByPlatform(rkPlatform)
	if err != nil {
		Logger.Error("查询AI模型列表失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询AI模型列表失败")
		return
	}
	if len(aiModelList) != len(req.AiModelRelationInfoList) {
		model.JsonResponseSysERR(c, "传入的AI模型列表数量不一致")
		return
	}

	// 先删除全部关联，再进行新增
	delResult := dao.DeleteIpcAiModelRelation(req.DeviceID, req.IpcID)
	if delResult != nil {
		model.JsonResponseSysERR(c, "保存关联失败")
		return
	}

	aiModelRelationInfoList := req.AiModelRelationInfoList
	for _, aiModelRelationInfo := range aiModelRelationInfoList {
		if aiModelRelationInfo.ModelPlatform != "" && rkPlatform != aiModelRelationInfo.ModelPlatform {
			model.JsonResponseSysERR(c, "设备所属平台与AI模型平台不一致")
			return
		}
		if aiModelRelationInfo.AiModelConfidence < 0.7 || aiModelRelationInfo.AiModelConfidence > 1.0 {
			aiModelRelationInfo.AiModelConfidence = 0.8
		}
		// relationEnable，no: 不建立关联，yes: 建立关联
		if aiModelRelationInfo.RelationEnable == "yes" {
			// 创建新的关联对象
			relation := model.DeviceAiModelRelation{
				DeviceID:          req.DeviceID,
				IpcId:             req.IpcID,
				AiModelID:         aiModelRelationInfo.AiModelID,
				Status:            aiModelRelationInfo.VideoRecordStatus,
				AiModelConfidence: aiModelRelationInfo.AiModelConfidence,
			}
			count, err := dao.CreateDeviceAiModelRelation(&relation)
			if err != nil || count <= 0 {
				model.JsonResponseSysERR(c, "关联失败")
				return
			}
		}
	}
	model.JsonResponseSucc(c, "关联成功")
}
