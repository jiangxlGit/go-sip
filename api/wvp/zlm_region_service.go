package wvp

import (
	"fmt"
	"go-sip/dao"
	. "go-sip/logger"
	"go-sip/model"

	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// zlm节点地区信息列表初始化
func ZlmNodeRegionInfoInit() error {
	// 查询mysql
	zlm_node_region_list, err := dao.GetRegions()
	if err != nil {
		Logger.Error("zlm节点地区信息列表查询失败", zap.Error(err))
		return fmt.Errorf("zlm节点地区信息列表查询失败")
	}
	// 遍历列表，写入redis
	for _, zlm_node_region := range zlm_node_region_list {
		// 根据关联地区编码查询zlm节点信息
		zlm_node_info_list, err := dao.GetZlmNodesByCode(zlm_node_region.RelationRegionCode)
		if err != nil || len(zlm_node_info_list) == 0 {
			Logger.Error("地区关联的zlm节点信息查询失败", zap.Error(err))
			continue
		}
		redis_util.HSetStruct_2(redis.WVP_REGION_RELATION_ZLM_INFO, zlm_node_region.RegionCode, zlm_node_info_list)
	}
	Logger.Info("zlm节点地区信息列表初始化完成")
	return nil
}

// @Summary 查询zlm节点地区信息列表
// @Router /wvp/zlm/regionList [get]
func ZlmNodeRegionInfoList(c *gin.Context) {
	// 查询mysql
	zlm_node_region_list, err := dao.GetRegions()
	if err != nil {
		Logger.Error("zlm节点地区信息列表查询失败", zap.Error(err))
		model.JsonResponseSysERR(c, "zlm节点地区信息列表查询失败")
		return
	}
	model.JsonResponseSucc(c, zlm_node_region_list)
}

// @Summary 更新zlm节点地区信息
// @Router /wvp/zlm/regionUpdate/{id} [post]
func ZlmNodeRegionUpdate(c *gin.Context) {
	id := c.Param("id")

	zlmRegionInfoDTO := model.ZlmRegionInfoDTO{}
	if err := c.ShouldBindJSON(&zlmRegionInfoDTO); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 查询mysql
	zlmRegionInfo, err := dao.GetRegionByID(id)
	if err != nil || zlmRegionInfo == nil {
		Logger.Error("id不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "id不存在")
		return
	}

	if zlmRegionInfo.RelationRegionCode == zlmRegionInfoDTO.RelationRegionCode {
		Logger.Warn("关联节点地区相同，不需要修改")
		model.JsonResponseSucc(c, "关联节点地区相同，不需要修改")
		return
	}

	// 更新字段
	zlmRegionInfo.RelationRegionCode = zlmRegionInfoDTO.RelationRegionCode
	zlmRegionInfo.RelationRegionName = zlmRegionInfoDTO.RelationRegionName

	// 更新mysql
	err = dao.UpdateRegion(zlmRegionInfo)
	if err != nil {
		model.JsonResponseSysERR(c, "更新失败")
		return
	}
	// 根据关联地区编码查询zlm节点信息
	zlm_node_info_list, err := dao.GetZlmNodesByCode(zlmRegionInfo.RelationRegionCode)
	if err != nil || len(zlm_node_info_list) == 0 {
		Logger.Error("地区关联的zlm节点信息查询失败", zap.Error(err))
	} else {
		redis_util.HSetStruct_2(redis.WVP_REGION_RELATION_ZLM_INFO, zlmRegionInfo.RegionCode, zlm_node_info_list)
	}
	model.JsonResponseSucc(c, zlmRegionInfo)
}
