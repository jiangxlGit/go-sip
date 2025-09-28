package model

import "go-sip/utils"

// AI模型信息表
type AiModelInfo struct {
	ID                 string            `json:"id"`                 // ai模型ID
	Name               string            `json:"name"`               // ai模型名称
	MianCategory       string            `json:"mianCategory"`       // 主类别code
	SubCategory        string            `json:"subCategory"`        // 子类别code
	MianCategoryName   string            `json:"mianCategoryName"`   // 主类别name(不是表字段)
	SubCategoryName    string            `json:"subCategoryName"`    // 子类别name(不是表字段)
	ModelFileName      string            `json:"modelFileName"`      // 模型文件名
	ModelFileKey       string            `json:"modelFileKey"`       // 模型文件key
	ModelFileURL       string            `json:"modelFileUrl"`       // 模型文件下载url
	ModelFileMd5       string            `json:"modelFileMd5"`       // 模型文件md5
	ModelVersion       string            `json:"modelVersion"`       // 模型版本
	ModelPlatform      string            `json:"modelPlatform"`      // 模型平台
	ModelLabels        string            `json:"modelLabels"`        // 模型标签json字符串
	ModelLabelList     []*ModelLabelInfo `json:"modelLabelList"`     // 模型标签列表
	AiModelConfidence  float32           `json:"aiModelConfidence"`  //模型置信度
	FunctionInfo       string            `json:"functionInfo"`       // 功能
	Status             int               `json:"status"`             // 状态（0：禁用，1：启用）
	IsDefault          int               `json:"isDefault"`          // 是否默认（0：否，1：是）
	CreateTime         utils.JSONTime    `json:"createTime"`         // 创建时间
	UpdateTime         utils.JSONTime    `json:"updateTime"`         // 更新时间
	Remarks            string            `json:"remarks"`            // 备注
	ModelLocalFilePath string            `json:"modelLocalFilePath"` // 模型本地文件路径
}
type IotDeviceRelationAiModelInfo struct {
	ModelSaveFileName string            `json:"modelFileName"`  // 板端保存的模型文件名
	ModelSaveFileMd5  string            `json:"modelFileMd5"`   // 板端保存的模型文件md5
	ModelVersion      string            `json:"modelVersion"`   // 模型版本
	ModelPlatform     string            `json:"modelPlatform"`  // 模型平台
	ModelLabelList    []*ModelLabelInfo `json:"modelLabelList"` // 模型标签列表
}

type ModelLabelInfo struct {
	LabelName string  `json:"labelName"` // 标签名称
	Score     float64 `json:"score"`     // 准确度
}

// 设备AI模型关联表
type DeviceAiModelRelation struct {
	ID                int64          `json:"id"`
	DeviceID          string         `json:"deviceId"`          //设备ID
	IpcId             string         `json:"ipcId"`             //摄像头id
	AiModelID         string         `json:"aiModelId"`         //模型ID
	AiModelConfidence float32        `json:"aiModelConfidence"` //模型置信度
	Status            int            `json:"status"`            //ai视频录制启用状态（0：禁用，1：启用）
	CreateTime        utils.JSONTime `json:"createTime"`        //创建时间
	UpdateTime        utils.JSONTime `json:"updateTime"`        //更新时间
	Remarks           string         `json:"remarks"`           //备注
}

// AI模型类别表
type AiModelCategory struct {
	ID           int64  `json:"id"`           // 记录id
	CategoryName string `json:"categoryName"` // 类别名称
	CategoryCode string `json:"categoryCode"` // 类别code
	ParentCode   string `json:"parentCode"`   // 父级code，000表示顶级类别
	Remarks      string `json:"remarks"`      // 备注
}

// AI模型类别响应
type AiModelCategoryRep struct {
	ID                 int64  `json:"id"`                 // 记录id
	CategoryName       string `json:"categoryName"`       // 类别名称
	CategoryCode       string `json:"categoryCode"`       // 类别code
	ParentCategoryName string `json:"parentCategoryName"` // 父级类别名称
	ParentCode         string `json:"parentCode"`         // 父级code，000表示顶级类别
	Count              int    `json:"count"`              // 该类别下的模型数量
	Remarks            string `json:"remarks"`            // 备注
}

type AiModelCategorySaveReq struct {
	CategoryName string `json:"categoryName" binding:"required"` // 类别名称
	// CategoryCode string `json:"categoryCode" binding:"required"` // 类别code
	ParentCode string `json:"parentCode" binding:"required"` // 父级code，000表示顶级类别
	Remarks    string `json:"remarks"`                       // 备注
}

type AiModelCategoryUpdateReq struct {
	CategoryName string `json:"categoryName" binding:"required"` // 类别名称
	Remarks      string `json:"remarks"`                         // 备注
}

// 设备关联ai模型信息
type DeviceRelationAiModelInfoRep struct {
	DeviceID         string            `json:"deviceId"`         // 设备ID
	RelationId       int64             `json:"relationId"`       // 设备关联ID
	AiModelId        string            `json:"aiModelId"`        // ai模型ID
	Name             string            `json:"name"`             // ai模型名称
	MianCategory     string            `json:"mianCategory"`     // 主类别code
	SubCategory      string            `json:"subCategory"`      // 子类别code
	MianCategoryName string            `json:"mianCategoryName"` // 主类别名称
	SubCategoryName  string            `json:"subCategoryName"`  // 子类别名称
	ModelFileName    string            `json:"modelFileName"`    // 模型文件名
	ModelFileKey     string            `json:"modelFileKey"`     // 模型文件key
	ModelFileURL     string            `json:"modelFileUrl"`     // 模型文件下载url
	ModelFileMd5     string            `json:"modelFileMd5"`     // 模型文件md5
	ModelVersion     string            `json:"modelVersion"`     // 模型版本
	ModelPlatform    string            `json:"modelPlatform"`    // 模型平台
	ModelLabelList   []*ModelLabelInfo `json:"modelLabelList"`   // 模型标签列表
	FunctionInfo     string            `json:"functionInfo"`     // 功能
	RelationStatus   string            `json:"relationStatus"`   // 关联状态（no：未关联 ，yes：已关联）
	ActiveStatus     int               `json:"activeStatus"`     // 启用/禁用状态（0：未启用，1：已启用）
	CreateTime       utils.JSONTime    `json:"createTime"`       // 创建时间
	UpdateTime       utils.JSONTime    `json:"updateTime"`       // 更新时间
	Remarks          string            `json:"remarks"`          // 备注
}

// 摄像头关联ai模型信息
type IpcRelationAiModelInfoRep struct {
	DeviceID          string            `json:"deviceId"` // 设备ID
	IpcID             string            `json:"ipcId"`
	RelationId        int64             `json:"relationId"`        // 模型关联ID
	RelationEnable    string            `json:"relationEnable"`    // 关联启用状态
	AiModelId         string            `json:"aiModelId"`         // ai模型ID
	AiModelName       string            `json:"aiModelName"`       // ai模型名称
	MianCategory      string            `json:"mianCategory"`      // 主类别code
	SubCategory       string            `json:"subCategory"`       // 子类别code
	MianCategoryName  string            `json:"mianCategoryName"`  // 主类别名称
	SubCategoryName   string            `json:"subCategoryName"`   // 子类别名称
	ModelVersion      string            `json:"modelVersion"`      // 模型版本
	ModelPlatform     string            `json:"modelPlatform"`     // 模型平台
	ModelLabelList    []*ModelLabelInfo `json:"modelLabelList"`    // 模型标签列表
	VideoRecordStatus int               `json:"videoRecordStatus"` // ai视频录制启用/禁用状态（0：未启用，1：已启用）
	IsDefault         int               `json:"isDefault"`         // 是否默认模型（0：否，1：是）
	AiModelConfidence float32           `json:"aiModelConfidence"` //模型置信度
}

// 摄像头关联ai模型标签信息
type IpcRelationAiModelLabelsRep struct {
	LabelName string `json:"labelName"` // 模型标签名称
	IsDefault int    `json:"isDefault"` // 是否默认模型（0：否，1：是）
}

// 新增ai模型结构体
type AiModelInfoSaveReq struct {
	ID            string `json:"id" binding:"required"`             // ai模型ID
	Name          string `json:"name"  binding:"required"`          // ai模型名称
	MianCategory  string `json:"mianCategory"  binding:"required"`  // 主类别
	SubCategory   string `json:"subCategory"  binding:"required"`   // 子类别
	ModelFileName string `json:"modelFileName"  binding:"required"` // 模型文件名
	ModelFileKey  string `json:"modelFileKey"  binding:"required"`  // 模型文件key
	ModelFileURL  string `json:"modelFileUrl"  binding:"required"`  // 模型文件下载url
	ModelFileMd5  string `json:"modelFileMd5"`                      // 模型文件md5
	ModelVersion  string `json:"modelVersion"`                      // 模型版本
	ModelPlatform string `json:"modelPlatform"`                     // 模型平台
	ModelLabels   string `json:"modelLabels"  binding:"required"`   // 模型标签json字符串
	FunctionInfo  string `json:"functionInfo"`                      // 功能
	IsDefault     int    `json:"isDefault" binding:"required"`      // 是否默认（0：否，1：是）
	Remarks       string `json:"remarks"`                           // 备注
}

// 修改ai模型结构体
type AiModelInfoUpdateReq struct {
	Name          string `json:"name"`          // ai模型名称
	MianCategory  string `json:"mianCategory"`  // 主类别
	SubCategory   string `json:"subCategory"`   // 子类别
	ModelFileName string `json:"modelFileName"` // 模型文件名
	ModelFileKey  string `json:"modelFileKey"`  // 模型文件key
	ModelFileURL  string `json:"modelFileUrl"`  // 模型文件下载url
	ModelFileMd5  string `json:"modelFileMd5"`  // 模型文件md5
	ModelVersion  string `json:"modelVersion"`  // 模型版本
	ModelPlatform string `json:"modelPlatform"` // 模型平台
	ModelLabels   string `json:"modelLabels"`   // 模型标签json字符串
	FunctionInfo  string `json:"functionInfo"`  // 功能
	Status        int    `json:"status"`        // 状态（0：禁用，1：启用）'
	IsDefault     int    `json:"isDefault"`     // 是否默认（0：否，1：是）
	Remarks       string `json:"remarks"`       // 备注
}

type IotDeviceRelationAiModelListRep struct {
	DeviceInfoList []*IotDeviceRelationAiModel `json:"deviceInfoList" binding:"required"`
}

type IotDeviceRelationAiModel struct {
	DeviceId string `json:"deviceId"` // IOT设备id
	Relation string `json:"relation"` // 关联关系， no：未关联，yes：已关联
}

type AiModelIdListReq struct {
	AiModelIdList []string `json:"aiModelIdList" binding:"required"`
}

type AiModelCountByMianCategory struct {
	MianCategory string `json:"mianCategory"` // 主类别
	Count        int    `json:"count"`        // 数量
}
type AiModelFileInfo struct {
	FileId          string `json:"fileId"`          // 文件名
	FileName        string `json:"fileName"`        // 文件名
	FileKey         string `json:"fileKey"`         // 文件key
	FileDownloadUrl string `json:"fileDownloadUrl"` // 文件下载地址
	FileMd5         string `json:"fileMd5"`         // 文件md5
}

type DeviceAiModelRelationSave struct {
	AiModelID string `json:"aiModelId" binding:"required"`
	DeviceID  string `json:"deviceId" binding:"required"`
	Remarks   string `json:"remarks"`
}

type DeviceAiModelRelationUpdate struct {
	AiModelID string `json:"aiModelId"`
	DeviceID  string `json:"deviceId"`
	Status    int    `json:"status"`
	Remarks   string `json:"remarks"`
}

type AiModelInfoListQuery struct {
	MianCategory string `json:"mianCategory" binding:"required"`
	SubCategory  string `json:"subCategory"`
	ID           string `json:"id"`
	Name         string `json:"name"`
}

type IotDeviceListByAiModelPageQuery struct {
	AiModelId string `json:"aiModelId" binding:"required"`
	Page      int    `json:"page" binding:"required,min=1"`
	Size      int    `json:"size" binding:"required,min=1"`
}

type DeviceAiModelRelationQuery struct {
	DeviceID     string `json:"deviceId" binding:"required"`
	Page         int    `json:"page" binding:"required,min=1"`
	Size         int    `json:"size" binding:"required,min=1"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	MianCategory string `json:"mianCategory"`
	SubCategory  string `json:"subCategory"`
}

type IpcAiModelLabelsQuery struct {
	DeviceID string `json:"deviceId" binding:"required"`
	IpcID    string `json:"ipcId" binding:"required"`
}
type IpcAiModelListQuery struct {
	DeviceID    string `json:"deviceId" binding:"required"`
	IpcID       string `json:"ipcId" binding:"required"`
	AiModelID   string `json:"aiModelId"`
	AiModelName string `json:"aiModelName"`
}

type IpcAiModelRelationReqData struct {
	DeviceID                string                `json:"deviceId" binding:"required"`                          // 设备ID，不能为空
	IpcID                   string                `json:"ipcId" binding:"required"`                             // 摄像头ID，不能为空
	AiModelRelationInfoList []AiModelRelationInfo `json:"aiModelIdList" binding:"required,min=1,dive,required"` // 模型ID列表，不能为空且每项非空
}

type AiModelRelationInfo struct {
	AiModelID         string  `json:"aiModelId" binding:"required"`          // 模型ID，必填
	ModelPlatform     string  `json:"modelPlatform" binding:"required"`      // 模型平台，必填
	RelationEnable    string  `json:"relationEnable" binding:"required"`     // 是否建立关联，no: 不建立关联，yes: 建立关联
	VideoRecordStatus int     `json:"videoRecordStatus" binding:"oneof=0 1"` // 录制状态，0或1
	AiModelConfidence float32 `json:"aiModelConfidence"`                     //模型置信度
}

func FromAIModelInfoSave(save *AiModelInfoSaveReq) *AiModelInfo {
	return &AiModelInfo{
		ID:            save.ID,
		Name:          save.Name,
		MianCategory:  save.MianCategory,
		SubCategory:   save.SubCategory,
		ModelFileName: save.ModelFileName,
		ModelFileKey:  save.ModelFileKey,
		ModelFileURL:  save.ModelFileURL,
		ModelFileMd5:  save.ModelFileMd5,
		ModelVersion:  save.ModelVersion,
		ModelPlatform: save.ModelPlatform,
		ModelLabels:   save.ModelLabels,
		FunctionInfo:  save.FunctionInfo,
		Remarks:       save.Remarks,
	}
}

func FromDevcieAiModelRelationSave(save *DeviceAiModelRelationSave) *DeviceAiModelRelation {
	return &DeviceAiModelRelation{
		DeviceID:  save.DeviceID,
		AiModelID: save.AiModelID,
		Remarks:   save.Remarks,
	}
}

func FromAIModelInfoUpdate(aiModelInfo *AiModelInfo, update *AiModelInfoUpdateReq) *AiModelInfo {
	if update == nil {
		return aiModelInfo
	}
	// 判断update参数是否为空，为空的不赋值
	if update.Name != "" {
		aiModelInfo.Name = update.Name
	}
	if update.MianCategory != "" {
		aiModelInfo.MianCategory = update.MianCategory
	}
	if update.SubCategory != "" {
		aiModelInfo.SubCategory = update.SubCategory
	}
	if update.ModelFileName != "" {
		aiModelInfo.ModelFileName = update.ModelFileName
	}
	if update.ModelFileKey != "" {
		aiModelInfo.ModelFileKey = update.ModelFileKey
	}
	if update.ModelFileURL != "" {
		aiModelInfo.ModelFileURL = update.ModelFileURL
	}
	if update.ModelFileMd5 != "" {
		aiModelInfo.ModelFileMd5 = update.ModelFileMd5
	}
	if update.ModelVersion != "" {
		aiModelInfo.ModelVersion = update.ModelVersion
	}
	if update.ModelPlatform != "" {
		aiModelInfo.ModelPlatform = update.ModelPlatform
	}
	if update.ModelLabels != "" {
		aiModelInfo.ModelLabels = update.ModelLabels
	}
	if update.FunctionInfo != "" {
		aiModelInfo.FunctionInfo = update.FunctionInfo
	}
	if update.Status != aiModelInfo.Status {
		aiModelInfo.Status = update.Status
	}
	if update.IsDefault != aiModelInfo.IsDefault {
		aiModelInfo.IsDefault = update.IsDefault
	}
	if update.Remarks != "" {
		aiModelInfo.Remarks = update.Remarks
	}
	return aiModelInfo
}

func FromDeviceAIModelRelationUpdate(r *DeviceAiModelRelation, update *DeviceAiModelRelationUpdate) *DeviceAiModelRelation {
	if update == nil {
		return r
	}
	if update.AiModelID != "" {
		r.AiModelID = update.AiModelID
	}
	if update.DeviceID != "" {
		r.DeviceID = update.DeviceID
	}
	if update.Status != r.Status {
		r.Status = update.Status
	}
	if update.Remarks != "" {
		r.Remarks = update.Remarks
	}
	return r
}

func FromAiModelCategorySave(save *AiModelCategorySaveReq) *AiModelCategory {
	return &AiModelCategory{
		CategoryName: save.CategoryName,
		ParentCode:   save.ParentCode,
		Remarks:      save.Remarks,
	}
}

func FromAiModelCategoryUpdate(aiModelCategory *AiModelCategory, update *AiModelCategoryUpdateReq) *AiModelCategory {
	if update.CategoryName != "" {
		aiModelCategory.CategoryName = update.CategoryName
	}
	if update.Remarks != "" {
		aiModelCategory.Remarks = update.Remarks
	}
	return aiModelCategory
}
