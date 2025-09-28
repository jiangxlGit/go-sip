package wvp

import (
	"go-sip/dao"
	. "go-sip/logger"
	"go-sip/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 查询设备关联AI模型的总数
// @Router /wvp/deviceAiModel/count [get]
func GetDeviceAiModelCount(c *gin.Context) {
	deviceId := c.Query("deviceId")
	if deviceId == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	model.JsonResponseSucc(c, dao.QueryDeviceAiModelRelationCount(deviceId))
}

// @Summary 根据设备id分页查询AI模型列表
// @Router /wvp/deviceAiModel/list [post]
func GetDeviceAiModelList(c *gin.Context) {
	var req model.DeviceAiModelRelationQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 分页查询已启用ai模型列表
	aiModelList, err := dao.GetAiModelPage(req.Page, req.Size, req.ID, req.Name, req.MianCategory, req.SubCategory)
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}

	var deviceRelationAiModelInfoList []*model.DeviceRelationAiModelInfoRep
	// 遍历ai模型列表，查询每个模型关联的信息
	for _, aiModel := range aiModelList {
		aiModelRelationList, err := dao.QueryDevcieAiModelRelations(req.DeviceID, aiModel.ID)
		if err != nil {
			Logger.Warn("查询设备关联的模型失败", zap.Any("devceId", req.DeviceID), zap.Any("aiModelId", aiModel.ID))
			continue
		}
		var info model.DeviceRelationAiModelInfoRep
		info.DeviceID = req.DeviceID
		info.AiModelId = aiModel.ID
		info.MianCategory = aiModel.MianCategory
		info.SubCategory = aiModel.SubCategory
		info.MianCategoryName = aiModel.MianCategoryName
		info.SubCategoryName = aiModel.SubCategoryName
		info.Name = aiModel.Name
		info.ModelFileName = aiModel.ModelFileName
		info.ModelFileKey = aiModel.ModelFileKey
		info.ModelFileURL = aiModel.ModelFileURL
		info.ModelFileMd5 = aiModel.ModelFileMd5
		info.ModelVersion = aiModel.ModelVersion
		info.FunctionInfo = aiModel.FunctionInfo
		info.CreateTime = aiModel.CreateTime
		info.UpdateTime = aiModel.UpdateTime
		info.Remarks = aiModel.Remarks
		if len(aiModelRelationList) == 0 {
			info.RelationStatus = "no"
			info.RelationId = -1
			info.ActiveStatus = 0
		} else {
			aiModelRelation := aiModelRelationList[0]
			info.RelationId = aiModelRelation.ID
			info.RelationStatus = "yes"
			info.ActiveStatus = aiModelRelation.Status
		}

		deviceRelationAiModelInfoList = append(deviceRelationAiModelInfoList, &info)
	}

	model.JsonResponsePageSucc(c, dao.GetAiModelCountByCondition(req.ID, req.Name,
		req.MianCategory, req.SubCategory), req.Page, req.Size, deviceRelationAiModelInfoList)
}

// @Summary 设备关联单个AI模型
// @Router /wvp/deviceAiModel/relation/{deviceId} [get]
func DeviceRelationAiModel(c *gin.Context) {
	deviceId := c.Param("deviceId")
	if deviceId == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	aiModelId := c.Query("aiModelId")
	// 创建新的关联对象
	relation := model.DeviceAiModelRelation{
		DeviceID:  deviceId,
		AiModelID: aiModelId,
	}
	// 根据设备ID和AI模型ID查询关联对象
	aiModelRelationList, err := dao.QueryDevcieAiModelRelations(deviceId, aiModelId)
	if err != nil {
		model.JsonResponseSysERR(c, "查询关联对象失败")
		return
	}
	if len(aiModelRelationList) > 0 {
		model.JsonResponseSysERR(c, "该设备已关联该AI模型")
		return
	}

	result, err := dao.CreateDeviceAiModelRelation(&relation)
	if err != nil || result <= 0 {
		model.JsonResponseSysERR(c, "写入失败")
		return
	}
	model.JsonResponseSucc(c, "关联成功")
}

// @Summary 设备关联多个AI模型
// @Router /wvp/deviceAiModel/relationMany/{deviceId} [post]
func DeviceRelationManyAiModel(c *gin.Context) {
	deviceId := c.Param("deviceId")
	if deviceId == "" {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	var req model.AiModelIdListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 遍历req.AiModelIdList,将关联对象写入数据库
	relationList := []*model.DeviceAiModelRelation{}
	for _, aiModelId := range req.AiModelIdList {
		// 创建新的关联对象
		relation := model.DeviceAiModelRelation{
			DeviceID:  deviceId,
			AiModelID: aiModelId,
		}
		relationList = append(relationList, &relation)
	}

	err := dao.BatchInsertDeviceAiModelRelation(relationList)
	if err != nil {
		model.JsonResponseSysERR(c, "批量写入失败")
		return
	}
	model.JsonResponseSucc(c, "关联成功")
}

// @Summary 取消设备AI模型关联
// @Router /wvp/deviceAiModel/delete/{relationId} [delete]
func DeleteDeviceAiModelRelation(c *gin.Context) {
	relationId := c.Param("relationId")
	if relationId == "" {
		model.JsonResponseSysERR(c, "relationId不能为空")
		return
	}

	err := dao.DeleteIpcAiModelRelationByID(relationId)
	if err != nil {
		model.JsonResponseSysERR(c, "取消关联失败")
		return
	}

	model.JsonResponseSucc(c, "取消关联成功")
}

// @Summary 启用/禁用关联的AI模型
// @Router /wvp/deviceAiModel/statusUpdate/{relationId} [put]
func UpdateAiModelRelationStatus(c *gin.Context) {
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
