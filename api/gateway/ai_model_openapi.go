package gateway

import (
	"go-sip/dao"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 查询所有已启用的ai模型
// @Router /open/aiModel/list [get]
func GetAiModelList(c *gin.Context) {
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		Logger.Error("deviceId不能为空")
		model.JsonResponseSysERR(c, "deviceId不能为空")
		return
	}

	// 筛选出已关联并且启用了关联的模型
	deviceAiModelRelationList, err := dao.QueryDevcieAiModelRelations(deviceId, "")
	if err != nil {
		Logger.Error("查询设备关联的AI模型失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询设备关联的AI模型失败")
		return
	}
	if len(deviceAiModelRelationList) == 0 {
		model.JsonResponseSysERR(c, "设备未关联AI模型")
		return
	}
	result := []*model.AiModelInfo{}
	for _, relation := range deviceAiModelRelationList {
		status := relation.Status
		if status == 1 {
			aiModel, err := dao.GetAiModelByID(relation.AiModelID)
			if err != nil || aiModel == nil || len(aiModel) == 0 {
				Logger.Error("设备未关联AI模型", zap.Error(err))
				model.JsonResponseSysERR(c, "设备未关联AI模型")
				return
			}
			modelLabels := aiModel[0].ModelLabels
			var modelLabelInfoList = []*model.ModelLabelInfo{}
			err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
			if err != nil {
				continue
			}
			aiModel[0].ModelLabelList = modelLabelInfoList
			result = append(result, aiModel[0])
		}
	}

	model.JsonResponseSucc(c, result)
}

// @Summary 查询所有已启用的ai模型并且已关联设备的+默认模型
// @Router /open/aiModelRelation/list [get]
func GetAiModelRelationList(c *gin.Context) {
	deviceId := c.Query("deviceId")
	deviceType := c.Query("deviceType")
	if deviceId == "" {
		Logger.Error("deviceId不能为空")
		model.JsonResponseSysERR(c, "deviceId不能为空")
		return
	}

	// 筛选出已关联并且启用了关联的模型
	deviceAiModelRelationList, err := dao.QueryDevcieAiModelRelations(deviceId, "")
	if err != nil {
		Logger.Error("查询设备关联的AI模型失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询设备关联的AI模型失败")
		return
	}

	ipcModelMap := make(map[string][]*model.AiModelInfo)
	modelDup := make(map[string]bool)
	for _, relation := range deviceAiModelRelationList {
		status := relation.Status
		if status == 1 && !modelDup[relation.IpcId+relation.AiModelID] {
			modelDup[relation.IpcId+relation.AiModelID] = true
			aiModel, err := dao.GetAiModelByID(relation.AiModelID)
			if err != nil || len(aiModel) == 0 || aiModel[0].ModelLabels == "" {
				Logger.Error("查询AI模型失败", zap.Any("aiModelId", relation.AiModelID), zap.Error(err))
				continue
			}
			modelLabels := aiModel[0].ModelLabels
			var modelLabelInfoList = []*model.ModelLabelInfo{}
			err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
			if err != nil {
				continue
			}
			aiModel[0].ModelLabelList = modelLabelInfoList
			aiModel[0].AiModelConfidence = relation.AiModelConfidence
			ipcModelMap[relation.IpcId] = append(ipcModelMap[relation.IpcId], aiModel[0])
		}
	}
	// 查询默认模型
	defaultAiModelList, err := dao.GetDefaultAiModelList(deviceType)
	if err == nil {
		for _, defaultAiModel := range defaultAiModelList {
			modelLabels := defaultAiModel.ModelLabels
			var modelLabelInfoList = []*model.ModelLabelInfo{}
			err = utils.JSONDecode([]byte(modelLabels), &modelLabelInfoList)
			if err != nil {
				continue
			}
			defaultAiModel.ModelLabelList = modelLabelInfoList
			defaultAiModel.AiModelConfidence = 0.85 // 默认模型置信度
		}
		ipcModelMap["default_ai_models"] = append(ipcModelMap["default_ai_models"], defaultAiModelList...)
	}

	model.JsonResponseSucc(c, ipcModelMap)
}
