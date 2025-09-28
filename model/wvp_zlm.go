package model

import "strconv"

// ZlmRegionInfo 表示 zlm 区域信息表
type ZlmRegionInfo struct {
	ID                 int64  `json:"id"`                 // 记录id
	RegionCode         string `json:"regionCode"`         // 地区国标编号（如北京11）
	RegionName         string `json:"regionName"`         // 地区名称
	RelationRegionCode string `json:"relationRegionCode"` // 关联地区国标编号
	RelationRegionName string `json:"relationRegionName"` // zlm节点关联地区
	Remarks            string `json:"remarks"`            // 备注
}

// ZlmNodeInfo 表示 zlm 节点信息表
type ZlmNodeInfo struct {
	ID            int64  `json:"id"`            // 记录ID
	ZlmIP         string `json:"zlmIp"`         // zlm ip
	ZlmPort       int    `json:"zlmPort"`       // zlm端口
	ZlmDomain     string `json:"zlmDomain"`     // zlm域名
	ZlmSecret     string `json:"zlmSecret"`     // zlm密钥
	ZlmNodeRegion string `json:"zlmNodeRegion"` // zlm所在地区名
	RegionCode    string `json:"regionCode"`    // 地区编号
	ZlmNodeStatus string `json:"zlmNodeStatus"` // 启用状态（enable/disable）
	Remarks       string `json:"remarks"`       // 备注
}

// ZlmNodeInfoSaveDTO zlm节点信息新增DTO
type ZlmNodeInfoSaveDTO struct {
	ZlmIP         string `json:"zlmIp" binding:"required" example:"192.168.1.100"`       // zlm ip
	ZlmPort       int    `json:"zlmPort" binding:"required" example:"9092"`              // zlm端口
	ZlmDomain     string `json:"zlmDomain" binding:"required" example:"zlm.example.com"` // zlm域名
	ZlmSecret     string `json:"zlmSecret" binding:"required" example:"mySecretKey"`     // zlm的secret
	ZlmNodeRegion string `json:"zlmNodeRegion" binding:"required" example:"北京"`          // zlm所在地区名
	RegionCode    string `json:"regionCode" binding:"required" example:"11"`             // 地区编号
	Remarks       string `json:"remarks" example:"主用节点"`                                 // 备注
}

// ZlmNodeInfoRedisDTO 用于 Redis 缓存存储的 ZLM 节点信息结构体
type ZlmNodeInfoRedisDTO struct {
	ZlmIP     string `json:"zlmIp"`     // zlm ip
	ZlmDomain string `json:"zlmDomain"` // zlm域名
	ZlmSecret string `json:"zlmSecret"` // zlm密钥
	ZlmPort   string    `json:"zlmPort"`   // zlm端口
}

// ZlmRegionDTO 表示 zlm 节点信息新增请求的 DTO
type ZlmRegionDTO struct {
	RegionCode string `json:"regionCode" binding:"required"` // 地区国标编号（如北京11）
	RegionName string `json:"regionName" binding:"required"` // 地区名称
}

type ZlmRegionInfoDTO struct {
	RelationRegionCode string `json:"relationRegionCode" binding:"required"` // 关联地区国标编号
	RelationRegionName string `json:"relationRegionName" binding:"required"` // zlm节点关联地区
	Remarks            string `json:"remarks"`                               // 备注
}

// FromZlmNodeInfoSaveDTO 将 DTO 转为实体
func FromZlmNodeInfoSaveDTO(dto *ZlmNodeInfoSaveDTO) *ZlmNodeInfo {
	return &ZlmNodeInfo{
		ZlmIP:         dto.ZlmIP,
		ZlmPort:       dto.ZlmPort,
		ZlmDomain:     dto.ZlmDomain,
		ZlmSecret:     dto.ZlmSecret,
		ZlmNodeRegion: dto.ZlmNodeRegion,
		RegionCode:    dto.RegionCode,
		Remarks:       dto.Remarks,
		ZlmNodeStatus: "enable", // 默认值
	}
}

// FromZlmNodeInfo 转换实体到RedisDTO
func FromZlmNodeInfoSaveDTOToRedisDTO(dto *ZlmNodeInfoSaveDTO) *ZlmNodeInfoRedisDTO {
	return &ZlmNodeInfoRedisDTO{
		ZlmIP:     dto.ZlmIP,
		ZlmDomain: dto.ZlmDomain,
		ZlmSecret: dto.ZlmSecret,
		ZlmPort:   strconv.Itoa(dto.ZlmPort),
	}
}
// FromZlmNodeInfo 转换实体到RedisDTO
func FromZlmNodeInfoToRedisDTO(dto *ZlmNodeInfo) *ZlmNodeInfoRedisDTO {
	return &ZlmNodeInfoRedisDTO{
		ZlmIP:     dto.ZlmIP,
		ZlmDomain: dto.ZlmDomain,
		ZlmSecret: dto.ZlmSecret,
		ZlmPort:   strconv.Itoa(dto.ZlmPort),
	}
}
