package wvp

import (
	"go-sip/dao"
	. "go-sip/db/alioss"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 根据主分类查询ai模型列表
// @Router /wvp/aiModel/list [post]
func QueryAiModelList(c *gin.Context) {
	var query model.AiModelInfoListQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 如果都为空，查询所有
	models, err := dao.GetAiModelList(query.MianCategory, query.SubCategory, query.ID, query.Name, "", "", "")
	if err != nil {
		Logger.Error("查询AI模型失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询AI模型失败")
		return
	}
	// 遍历models

	for _, m := range models {
		modelLabelInfoJson := m.ModelLabels
		var modelLabelInfoList = []*model.ModelLabelInfo{}
		err := utils.JSONDecode([]byte(modelLabelInfoJson), &modelLabelInfoList)
		if err != nil {
			continue
		}
		m.ModelLabelList = modelLabelInfoList
	}
	model.JsonResponseSucc(c, models)
}

// @Summary 根据主分类查询AI模型总数
// @Router /wvp/aiModel/count [get]
func GetAiModelCount(c *gin.Context) {
	list, err := dao.GetAiModelCountByMianCategory()
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	model.JsonResponseSucc(c, list)
}

// @Summary 新增AI模型
// @Router /wvp/aiModel/add [post]
func AddAiModel(c *gin.Context) {
	var aiModelInfoSave model.AiModelInfoSaveReq
	if err := c.ShouldBindJSON(&aiModelInfoSave); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	exist, err := dao.GetAiModelByID(aiModelInfoSave.ID)
	if err != nil {
		Logger.Error("AI模型查询失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型查询失败")
		return
	}
	if exist != nil {
		model.JsonResponseSysERR(c, "模型ID已存在")
		return
	}
	// 模型名称不能重复
	if dao.GetAiModelByName(aiModelInfoSave.Name) != nil {
		model.JsonResponseSysERR(c, "模型名称已存在")
		return
	}

	// 查询AI模型类别是否存在
	aiModelCategoryMian, err1 := dao.GetAiModelCategorieListByCode(aiModelInfoSave.MianCategory)
	aiModelCategorySub, err2 := dao.GetAiModelCategorieListByCode(aiModelInfoSave.SubCategory)
	if err1 != nil || err2 != nil {
		Logger.Error("AI模型类别查询失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型类别查询失败")
		return
	}
	if aiModelCategoryMian == nil || aiModelCategorySub == nil {
		model.JsonResponseSysERR(c, "模型类别不存在")
		return
	}
	// 模型标签不能重复，a,b,c,c这样的就是有重复的
	if _, err := utils.StrArrCheckDuplicates(aiModelInfoSave.ModelLabels); err != nil {
		model.JsonResponseSysERR(c, "模型标签不能重复")
		return
	}

	// 同一个主类下的子类不能重复
	if dao.GetAiModelCountByMianAndSubCategory(aiModelCategoryMian.CategoryCode, aiModelCategorySub.CategoryCode) > 0 {
		model.JsonResponseSysERR(c, "模型类别已存在相关模型")
		return
	}

	rows, err := dao.CreateAiModel(model.FromAIModelInfoSave(&aiModelInfoSave))
	if err != nil || rows == 0 {
		Logger.Error("AI模型写入失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型写入失败")
		return
	}

	model.JsonResponseSucc(c, aiModelInfoSave)
}

// @Summary 修改AI模型信息
// @Router /wvp/aiModel/update/{id} [put]
func UpdateAiModel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}

	var aiModelInfoUpdate model.AiModelInfoUpdateReq
	if err := c.ShouldBindJSON(&aiModelInfoUpdate); err != nil {
		DeleteObject(aiModelInfoUpdate.ModelFileKey)
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	aiModelList, err := dao.GetAiModelByID(id)
	if err != nil || aiModelList == nil {
		DeleteObject(aiModelInfoUpdate.ModelFileKey)
		Logger.Error("AI模型不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型不存在")
		return
	}

	// 查询AI模型类别是否存在
	aiModelCategoryMian, err1 := dao.GetAiModelCategorieListByCode(aiModelInfoUpdate.MianCategory)
	aiModelCategorySub, err2 := dao.GetAiModelCategorieListByCode(aiModelInfoUpdate.SubCategory)
	if err1 != nil || err2 != nil {
		DeleteObject(aiModelInfoUpdate.ModelFileKey)
		Logger.Error("AI模型类别查询失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型类别查询失败")
		return
	}
	if aiModelCategoryMian == nil || aiModelCategorySub == nil {
		DeleteObject(aiModelInfoUpdate.ModelFileKey)
		model.JsonResponseSysERR(c, "模型类别不存在")
		return
	}

	// 模型标签不能重复，a,b,c,c这样的就是有重复的
	if _, err := utils.StrArrCheckDuplicates(aiModelInfoUpdate.ModelLabels); err != nil {
		model.JsonResponseSysERR(c, err)
		return
	}

	newAiModel := model.FromAIModelInfoUpdate(aiModelList[0], &aiModelInfoUpdate)
	newAiModel.ID = aiModelList[0].ID

	err = dao.UpdateAiModel(newAiModel)
	if err != nil {
		Logger.Error("AI模型更新失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型更新失败")
		return
	}

	// 更新成功后，如果file_key有变化，则删除旧文件
	if aiModelList[0].ModelFileKey != aiModelInfoUpdate.ModelFileKey {
		DeleteObject(aiModelList[0].ModelFileKey)
	}

	model.JsonResponseSucc(c, newAiModel)
}

// @Summary 启用/禁用AI模型
// @Router /wvp/aiModel/statusUpdate/{id} [put]
func UpdateAiModelStatus(c *gin.Context) {
	id := c.Param("id")
	statusStr := c.Query("status")
	if statusStr != "0" && statusStr != "1" {
		model.JsonResponseSysERR(c, "参数status错误（必须为0或1）")
		return
	}

	aiModel, err := dao.GetAiModelByID(id)
	if err != nil || aiModel == nil {
		Logger.Error("AI模型不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型不存在")
		return
	}

	if statusStr == "0" {
		aiModel[0].Status = 0
	} else {
		aiModel[0].Status = 1
	}

	err = dao.UpdateAiModel(aiModel[0])
	if err != nil {
		Logger.Error("AI模型状态更新失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型状态更新失败")
		return
	}

	model.JsonResponseSucc(c, aiModel)
}

// @Summary 删除AI模型
// @Router /wvp/aiModel/delete/{id} [delete]
func DeleteAiModel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}
	aiModelList, err := dao.GetAiModelByID(id)
	if err != nil || aiModelList == nil {
		Logger.Error("AI模型不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型不存在")
		return
	}

	// 判断是否有关联的设备
	list, err := dao.QueryIpcAiModelRelations("", id)
	if err != nil {
		model.JsonResponseSysERR(c, "查询关联的设备失败")
		return
	}
	if len(list) > 0 {
		model.JsonResponseSysERR(c, "该AI模型有关联的设备，请先解除关联")
		return
	}

	err = dao.DeleteAiModel(id)
	if err != nil {
		model.JsonResponseSysERR(c, "删除失败")
		return
	}

	// 删除oss文件
	DeleteObject(aiModelList[0].ModelFileKey)

	model.JsonResponseSucc(c, "删除成功")
}

// @Summary AI模型批量关联中控设备
// @Router /wvp/aiModel/relation/{id} [post]
func AddAiModelRelation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "参数错误")
	}
	var req model.IotDeviceRelationAiModelListRep
	if err := c.ShouldBindJSON(&req); err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	aiModel, err := dao.GetAiModelByID(id)
	if err != nil || aiModel == nil {
		Logger.Error("AI模型不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型不存在")
		return
	}

	// 遍历req.DeviceIdList，将每个设备ID添加到数据库中
	relations := []*model.DeviceAiModelRelation{}
	for _, deviceInfo := range req.DeviceInfoList {
		// 先删除旧数据
		dao.DeleteDeviceAiModelRelationByAiModelIdAndDeviceId(id, deviceInfo.DeviceId)
		relationStatus := deviceInfo.Relation
		// 关联的对象
		if relationStatus == "yes" {
			relation := model.DeviceAiModelRelation{
				AiModelID: id,
				DeviceID:  deviceInfo.DeviceId,
			}
			relations = append(relations, &relation)
		}
	}
	// 调用数据库函数将关联对象写入数据库
	err = dao.BatchInsertDeviceAiModelRelation(relations)
	if err != nil {
		model.JsonResponseSysERR(c, "添加关联失败")
		return
	}
	model.JsonResponseSucc(c, "添加关联成功")
}
