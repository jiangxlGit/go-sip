package wvp

import (
	"go-sip/dao"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/model"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// zlm节点信息初始化
func ZlmNodeInfoInit() {
	// 查询mysql
	zlm_node_list, err := dao.GetAllZlmNodes()
	if err != nil {
		Logger.Error("zlm节点信息列表查询失败", zap.Error(err))
		return
	}
	// 遍历列表，写入redis
	for _, zlm_node := range zlm_node_list {
		zlmRedisDTO := model.ZlmInfo{}
		zlmRedisDTO.ZlmIp = zlm_node.ZlmIP
		zlmRedisDTO.ZlmPort = strconv.Itoa(zlm_node.ZlmPort)
		zlmRedisDTO.ZlmSecret = zlm_node.ZlmSecret
		zlmRedisDTO.ZlmDomain = zlm_node.ZlmDomain
		redis_util.HSetStruct_2(redis.WVP_ZLM_NODE_INFO, zlm_node.ZlmDomain, zlmRedisDTO)
	}
	Logger.Info("zlm节点信息初始化完成")
}

// @Summary 查询zlm节点地区信息列表
// @Router /wvp/zlm/nodeList [get]
func GetZlmNodeInfoList(c *gin.Context) {
	// 查询mysql
	zlm_node_list, err := dao.GetAllZlmNodes()
	if err != nil {
		Logger.Error("zlm节点信息列表查询失败", zap.Error(err))
		model.JsonResponseSysERR(c, "zlm节点信息列表查询失败")
		return
	}
	model.JsonResponseSucc(c, zlm_node_list)
}

// @Summary		根据deviceId查询zlm节点信息接口
// @Router		/wvp/zlm/nodeInfo/{deviceId} [get]
func GetZlmNodeInfo(c *gin.Context) {
	deviceId := c.Param("deviceId")
	// 不能为空
	if deviceId == "" {
		model.JsonResponseSysERR(c, "devcieId不能为空")
		return
	}

	zlmInfo, err := WvpGetZlmInfo(deviceId)
	if err != nil || zlmInfo == nil {
		Logger.Error("根据device_id查询zlmInfo失败", zap.Error(err))
		model.JsonResponseSysERR(c, "根据device_id查询zlmInfo失败")
		return
	}
	model.JsonResponseSucc(c, zlmInfo)
}

// @Summary 新增zlm节点
// @Router /wvp/zlm/nodeAdd [post]
func AddZlmNode(c *gin.Context) {
	var dto model.ZlmNodeInfoSaveDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	zlmNodeInfo, err := dao.GetZlmNodeByZlmIPAndZlmPort(dto.ZlmIP, dto.ZlmPort)
	if zlmNodeInfo != nil {
		Logger.Error("zlm节点已存在", zap.Error(err))
		model.JsonResponseSysERR(c, "zlm节点已存在")
		return
	}

	// zlmIp为hash key，zlm信息为value，写入到Redis
	zlmRedisDTO := model.FromZlmNodeInfoSaveDTOToRedisDTO(&dto)
	err = redis_util.HSetStruct_2(redis.WVP_ZLM_NODE_INFO, dto.ZlmDomain, zlmRedisDTO)
	if err != nil {
		Logger.Error("ZLM_NODE Redis写入失败", zap.Error(err))
		model.JsonResponseSysERR(c, "Redis写入失败")
		return
	}
	zlmNodeInfo = model.FromZlmNodeInfoSaveDTO(&dto)
	// 写入msyql
	id, err := dao.CreateZlmNode(zlmNodeInfo)
	if err != nil || id <= 0 {
		Logger.Error("ZLM_NODE mysql写入失败", zap.Error(err))
		model.JsonResponseSysERR(c, "mysql写入失败")
		return
	}
	zlmNodeInfo.ID = id
	model.JsonResponseSucc(c, zlmNodeInfo)
}

// @Summary 删除zlm节点信息
// @Router /wvp/zlm/nodeDelete/{id} [delete]
func ZlmNodeDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}
	zlmNodeInfo, err := dao.GetZlmNodeByID(id)
	if err != nil || zlmNodeInfo == nil {
		Logger.Error("id不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "id不存在")
		return
	}
	// 判断节点是否有关联的地区
	zlmRegionInfo, err := dao.GetRegionByRelationRegionCode(zlmNodeInfo.RegionCode)
	if err != nil {
		model.JsonResponseSysERR(c, "删除失败")
		return
	}
	if zlmRegionInfo != nil {
		model.JsonResponseSysERR(c, "该节点有地区关联，请先解除关联")
		return
	}
	err = dao.DeleteZlmNode(id)
	if err != nil {
		Logger.Error("删除失败", zap.Error(err))
		model.JsonResponseSysERR(c, "删除失败")
		return
	}
	// 地区关联zlm重新初始化
	err = ZlmNodeRegionInfoInit()
	if err != nil {
		Logger.Error("地区关联zlm重新初始化失败", zap.Error(err))
		model.JsonResponseSysERR(c, "地区关联zlm重新初始化失败")
		return
	}
	model.JsonResponseSucc(c, "删除成功")
}

// @Summary 修改zlm节点信息
// @Router /wvp/zlm/nodeUpdate/{id} [put]
func UpdateZlmNodeInfo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.JsonResponseSysERR(c, "id不能为空")
		return
	}

	var dto model.ZlmNodeInfoSaveDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		Logger.Error("参数错误", zap.Error(err))
		model.JsonResponseSysERR(c, "参数错误")
		return
	}

	// 查询mysql
	zlmNodeInfo, err := dao.GetZlmNodeByID(id)
	if err != nil || zlmNodeInfo == nil {
		Logger.Error("id不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "id不存在")
		return
	}
	zlmNodeInfoNew := model.FromZlmNodeInfoSaveDTO(&dto)
	zlmNodeInfoNew.ID = zlmNodeInfo.ID
	// 更新msyql
	err = dao.UpdateZlmNode(zlmNodeInfoNew)
	if err != nil {
		Logger.Error("ZLM_NODE mysql更新失败", zap.Error(err))
		model.JsonResponseSysERR(c, "mysql更新失败")
		return
	}
	// 更新redis
	// 先删除旧数据
	err = redis_util.HDel_2(redis.WVP_ZLM_NODE_INFO, zlmNodeInfo.ZlmDomain)
	if err != nil {
		Logger.Error("ZLM_NODE redis删除旧数据失败", zap.Error(err))
		model.JsonResponseSysERR(c, "redis删除旧数据失败")
		return
	}
	// 再set新数据
	zlmRedisDTO := model.FromZlmNodeInfoSaveDTOToRedisDTO(&dto)
	err = redis_util.HSetStruct_2(redis.WVP_ZLM_NODE_INFO, dto.ZlmDomain, zlmRedisDTO)
	if err != nil {
		Logger.Error("ZLM_NODE redis更新失败", zap.Error(err))
		model.JsonResponseSysERR(c, "redis更新失败")
		return
	}

	// 地区关联zlm重新初始化
	err = ZlmNodeRegionInfoInit()
	if err != nil {
		Logger.Error("地区关联zlm重新初始化失败", zap.Error(err))
		model.JsonResponseSysERR(c, "地区关联zlm重新初始化失败")
		return
	}

	model.JsonResponseSucc(c, zlmNodeInfoNew)
}

// @Summary 启用/禁用zlm节点
// @Router /wvp/zlm/nodeStatusUpdate/{id} [post]
func UpdateZlmNodeStatus(c *gin.Context) {
	id := c.Param("id")
	_, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		model.JsonResponseSysERR(c, "参数错误")
		return
	}
	// zlm节点状态(enable：启用，disable：禁用)，默认启用
	status := c.Query("status")
	// 参数status校验
	if status != "enable" && status != "disable" {
		model.JsonResponseSysERR(c, "参数status错误")
		return
	}

	// 查询mysql
	zlmNodeInfo, err := dao.GetZlmNodeByID(id)
	if err != nil || zlmNodeInfo == nil {
		Logger.Error("id不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "id不存在")
		return
	}
	region_code := zlmNodeInfo.RegionCode
	// 根据地区码查询所有启用zlm节点信息
	zlmNodeInfoList, err := dao.GetZlmNodesByCode(region_code)
	if status == "disable" && err == nil && len(zlmNodeInfoList) == 1 {
		Logger.Error("当前地区最后一个已启用的节点，无法禁用", zap.Error(err))
		model.JsonResponseSysERR(c, "当前地区最后一个已启用的节点，无法禁用")
		return
	}

	zlmNodeInfo.ZlmNodeStatus = status
	err = dao.UpdateZlmNode(zlmNodeInfo)
	if err != nil {
		Logger.Error("更新mysql节点状态失败", zap.Error(err))
		model.JsonResponseSysERR(c, "更新mysql节点状态失败")
		return
	}

	zlmRedisDTO := model.FromZlmNodeInfoToRedisDTO(zlmNodeInfo)
	err = redis_util.HSetStruct_2(redis.WVP_ZLM_NODE_INFO, zlmNodeInfo.ZlmDomain, zlmRedisDTO)
	if err != nil {
		Logger.Error("更新redis节点状态失败", zap.Error(err))
		model.JsonResponseSysERR(c, "更新redis节点状态失败")
		return
	}

	// 地区关联zlm重新初始化
	err = ZlmNodeRegionInfoInit()
	if err != nil {
		Logger.Error("地区关联zlm重新初始化失败", zap.Error(err))
		model.JsonResponseSysERR(c, "地区关联zlm重新初始化失败")
		return
	}

	model.JsonResponseSucc(c, zlmNodeInfo)
}

// @Summary		查询zlm节点关联地区列表
// @Router		/wvp/zlm/nodeRelationRegionList/{nodeId} [get]
func ZlmNodeRelationRegionList(c *gin.Context) {
	nodeId := c.Param("nodeId")
	if nodeId == "" {
		model.JsonResponseSysERR(c, "nodeId不能为空")
		return
	}

	// 查询mysql
	zlmNodeInfo, err := dao.GetZlmNodeByID(nodeId)
	if err != nil || zlmNodeInfo == nil {
		Logger.Error("nodeId不存在", zap.Error(err))
		model.JsonResponseSysERR(c, "nodeId不存在")
		return
	}
	regionCode := zlmNodeInfo.RegionCode
	if regionCode == "" {
		model.JsonResponseSysERR(c, "zlm节点信息中的regionCode不能为空")
		return
	}
	zlmRegionList, err := dao.GetRegionByRelationRegionCode(regionCode)
	if err != nil {
		Logger.Error("获取zlm_region_info失败", zap.Error(err))
		model.JsonResponseSysERR(c, "获取zlm_region_info失败")
		return
	}
	model.JsonResponseSucc(c, zlmRegionList)
}
