package wvp

import (
	"fmt"
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"

	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 添加ai模型类别
// @Router /wvp/aiModelCategory/add [post]
func AddAiModelCategory(c *gin.Context) {
	var req model.AiModelCategorySaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	// 如果parentCode不为000，则判断父类code是否存在
	parentCode := req.ParentCode
	if parentCode != "000" {
		if _, err := dao.GetAiModelCategorieListByCode(parentCode); err != nil {
			Logger.Error("父类code不存在", zap.Error(err))
			model.JsonResponseSysERR(c, "父类code不存在")
			return
		}
	}
	// 根据id查询是否存在
	aiModelCategory, err := dao.GetAiModelCategorieListByName(req.CategoryName)
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	if aiModelCategory != nil {
		model.JsonResponseSysERR(c, "该分类名称已存在，无法添加")
		return
	}
	aiModelCategoryNew := model.FromAiModelCategorySave(&req)
	// 查询redis，获取分类id自增值
	code, err := redis_util.Get_2(fmt.Sprintf(redis.AI_MODEL_CATEGORY_SEQ_KEY, parentCode))
	if err != nil || code == "" {
		code = "100"
		redis_util.Set_2(fmt.Sprintf(redis.AI_MODEL_CATEGORY_SEQ_KEY, parentCode), code, -1)
	} else {
		lastCodeInt, err := strconv.Atoi(code)
		if err != nil {
			code = "100"
		} else {
			code = strconv.Itoa(lastCodeInt + 1)
			redis_util.Set_2(fmt.Sprintf(redis.AI_MODEL_CATEGORY_SEQ_KEY, parentCode), code, -1)
		}
	}
	if parentCode != "000" {
		aiModelCategoryNew.CategoryCode = parentCode + "-" + code
	} else {
		aiModelCategoryNew.CategoryCode = code
	}
	id, err := dao.InsertAiModelCategory(aiModelCategoryNew)
	if err != nil {
		Logger.Error("AI模型类别写入失败", zap.Error(err))
		model.JsonResponseSysERR(c, "AI模型类别写入失败")
		return
	}
	aiModelCategoryNew.ID = id
	model.JsonResponseSucc(c, aiModelCategoryNew)
}

// @Summary 更新ai模型类别
// @Router /wvp/aiModelCategory/update/{id} [put]
func UpdateAiModelCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}
	// id转为int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		model.JsonResponseSysERR(c, "id格式错误")
		return
	}
	// 根据id查询是否存在
	aiModelCategory, err := dao.GetAiModelCategoryByID(idInt)
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	if aiModelCategory == nil {
		model.JsonResponseSysERR(c, "该分类不存在")
		return
	}

	var req model.AiModelCategoryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	aiModelCategory2, err2 := dao.GetAiModelCategorieListByName(req.CategoryName)
	if err2 != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	if aiModelCategory2 != nil {
		model.JsonResponseSysERR(c, "该分类名称已存在，无法修改")
		return
	}

	aiModelCategoryUpdate := model.FromAiModelCategoryUpdate(aiModelCategory, &req)
	err = dao.UpdateAiModelCategory(aiModelCategoryUpdate)
	if err != nil {
		Logger.Error("更新ai模型类别失败", zap.Error(err))
		model.JsonResponseSysERR(c, "更新ai模型类别失败")
		return
	}
	model.JsonResponseSucc(c, aiModelCategoryUpdate)
}

// @Summary 删除ai模型类别
// @Router /wvp/aiModelCategory/delete/{id} [delete]
func DeleteAiModelCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		model.JsonResponseSysERR(c, "参数格式错误")
		return
	}
	// 根据id查询是否存在
	aiModelCategory, err := dao.GetAiModelCategoryByID(idInt)
	if err != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	if aiModelCategory == nil {
		model.JsonResponseSysERR(c, "该分类不存在")
		return
	}

	// 是否有子分类
	list, err := dao.GetAiModelCategorieListByParentCode(aiModelCategory.CategoryCode, "")
	if err != nil {
		model.JsonResponseSysERR(c, "查询子分类失败")
		return
	}
	if len(list) > 0 {
		model.JsonResponseSysERR(c, "该分类下有子分类，请先删除子分类")
		return
	}

	// 判断是否有关联的ai模型
	list2, err2 := dao.GetAiModelListByMainCategory(aiModelCategory.CategoryCode)
	list3, err3 := dao.GetAiModelListBySubCategory(aiModelCategory.CategoryCode)
	if err2 != nil || err3 != nil {
		model.JsonResponseSysERR(c, "查询失败")
		return
	}
	if len(list2) > 0 || len(list3) > 0 {
		model.JsonResponseSysERR(c, "该分类下有关联的AI模型，请先删除")
		return
	}

	err = dao.DeleteAiModelCategory(idInt)
	if err != nil {
		model.JsonResponseSysERR(c, "删除失败")
		return
	}
	model.JsonResponseSucc(c, "删除成功")
}

// @Summary 查询ai模型类别列表
// @Router /wvp/aiModelCategory/list [get]
func QueryAiModelCategoryList(c *gin.Context) {
	parentCode := c.Query("parentCode")
	name := c.Query("name")
	if parentCode == "" {
		parentCode = "000"
	}
	aiModelCategoryList, err := dao.GetAiModelCategorieListByParentCode(parentCode, name)
	if err != nil {
		Logger.Error("查询ai模型类别列表失败", zap.Error(err))
		model.JsonResponseSysERR(c, "查询ai模型类别列表失败")
		return
	}
	model.JsonResponseSucc(c, aiModelCategoryList)
}
