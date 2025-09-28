package dao

import (
	"go-sip/db/mysql"
	"go-sip/model"
	"go-sip/utils"
)

// 创建 AI 模型
func CreateAiModel(aiModelInfo *model.AiModelInfo) (int64, error) {
	query := `
		INSERT INTO gowvp_ai_model_info (
			id, name, mian_category, sub_category, model_file_name, model_file_key,
			model_file_url, model_file_md5, model_version, model_platform, model_labels, 
			function_info, create_time,update_time, remarks
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := mysql.MysqlDB.Exec(query,
		aiModelInfo.ID, aiModelInfo.Name, aiModelInfo.MianCategory, aiModelInfo.SubCategory,
		aiModelInfo.ModelFileName, aiModelInfo.ModelFileKey, aiModelInfo.ModelFileURL, aiModelInfo.ModelFileMd5,
		aiModelInfo.ModelVersion, aiModelInfo.ModelPlatform, aiModelInfo.ModelLabels,
		aiModelInfo.FunctionInfo, utils.NowInCn(), utils.NowInCn(), aiModelInfo.Remarks)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// 分页查询已经启用的AI模型列表
func GetAiModelPage(page, pageSize int, id, name, mianCategory, subCategory string) ([]*model.AiModelInfo, error) {
	query := `SELECT 
		gami.id AS id, 
		gami.name AS name, 
		IFNULL(gamc.category_code, '') AS mianCategory,
		IFNULL(gamc2.category_code, '') AS subCategory,
		IFNULL(gamc.category_name, '') AS mianCategoryName, 
		IFNULL(gamc2.category_name, '') AS subCategoryName,
		IFNULL(gami.function_info, '') AS functionInfo,
		IFNULL(gami.model_file_name, '') AS modelFileName,
		IFNULL(gami.model_file_url, '') AS modelFileUrl,
		IFNULL(gami.model_file_md5, '') AS modelFileMd5,
		IFNULL(gami.model_version, '') AS modelVersion,
		IFNULL(gami.model_platform, '') AS modelPlatform,
		IFNULL(gami.model_labels, '') AS modelLabels,
		gami.status AS status,
		gami.create_time AS createTime,
		gami.update_time AS updateTime,
		IFNULL(gami.remarks, '') AS remarks
	FROM gowvp_ai_model_info gami  
	LEFT JOIN gowvp_ai_model_category gamc ON gami.mian_category = gamc.category_code
	LEFT JOIN gowvp_ai_model_category gamc2 ON gami.sub_category = gamc2.category_code
	WHERE gami.status = 1
	`
	args := make([]interface{}, 0)
	if mianCategory != "" {
		query += " AND gami.mian_category = ? "
		args = append(args, mianCategory)
	}
	if subCategory != "" {
		query += " AND gami.sub_category = ? "
		args = append(args, subCategory)
	}
	if id != "" {
		query += " AND gami.id = ? "
		args = append(args, id)
	}
	if name != "" {
		query += " AND gami.name LIKE ? "
		args = append(args, "%"+name+"%")
	}
	// 排序
	query += " ORDER BY gami.mian_category, gami.sub_category"
	query += ` LIMIT ?, ?`
	args = append(args, (page-1)*pageSize, pageSize)

	rows, err := mysql.MysqlDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.AiModelInfo
	for rows.Next() {
		var m model.AiModelInfo
		if err := rows.Scan(&m.ID, &m.Name, &m.MianCategory, &m.SubCategory, &m.MianCategoryName, &m.SubCategoryName,
			&m.FunctionInfo, &m.ModelFileName, &m.ModelFileURL, &m.ModelFileMd5, &m.ModelVersion, &m.ModelPlatform,
			&m.ModelLabels, &m.Status, &m.CreateTime, &m.UpdateTime, &m.Remarks); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}

func GetAiModelCount() int64 {
	return GetAiModelCountByCondition("", "", "", "")
}

// 查询已经启用的AI模型总数
func GetAiModelCountByCondition(id, name, mianCategory, subCategory string) int64 {
	query := `
	SELECT count(*) AS count
	FROM gowvp_ai_model_info  
	WHERE status = 1`

	args := make([]interface{}, 0)
	if mianCategory != "" {
		query += " AND mian_category = ? "
		args = append(args, mianCategory)
	}
	if subCategory != "" {
		query += " AND sub_category = ? "
		args = append(args, subCategory)
	}
	if id != "" {
		query += " AND id = ? "
		args = append(args, id)
	}
	if name != "" {
		query += " AND name LIKE ? "
		args = append(args, "%"+name+"%")
	}

	var count int64
	err := mysql.MysqlDB.QueryRow(query, args...).Scan(&count)
	if err != nil || count < 0 {
		return 0
	}
	return count
}

func GetIpcAiModelList(id, name, modelPlatform string) ([]*model.AiModelInfo, error) {
	return GetAiModelList("", "", id, name, "1", modelPlatform, "")
}
func GetDefaultAiModelList(modelPlatform string) ([]*model.AiModelInfo, error) {
	return GetAiModelList("", "", "", "", "1", modelPlatform, "1")
}
func GetAiModelListByPlatform(modelPlatform string) ([]*model.AiModelInfo, error) {
	return GetAiModelList("", "", "", "", "1", modelPlatform, "0")
}

// 根据条件查询AI模型列表
func GetAiModelList(mianCategory, subCategory, id, name, status, modelPlatform, isDefault string) ([]*model.AiModelInfo, error) {
	query := `SELECT 
		gami.id AS id, 
		gami.name AS name, 
		IFNULL(gamc.category_code, '') AS mianCategory,
		IFNULL(gamc2.category_code, '') AS subCategory,
		IFNULL(gamc.category_name, '') AS mianCategoryName, 
		IFNULL(gamc2.category_name, '') AS subCategoryName,
		IFNULL(gami.function_info, '') AS functionInfo,
		IFNULL(gami.model_file_name, '') AS modelFileName,
		IFNULL(gami.model_file_key, '') AS modelFileKey,
		IFNULL(gami.model_file_url, '') AS modelFileUrl,
		IFNULL(gami.model_file_md5, '') AS modelFileMd5,
		IFNULL(gami.model_version, '') AS modelVersion,
		IFNULL(gami.model_platform, '') AS modelPlatform,
		IFNULL(gami.model_labels, '') AS modelLabels,
		gami.status AS status,
		gami.is_default AS isDefault,
		gami.create_time AS createTime,
		gami.update_time AS updateTime,
		IFNULL(gami.remarks, '') AS remarks
	FROM gowvp_ai_model_info gami  
	LEFT JOIN gowvp_ai_model_category gamc ON gami.mian_category = gamc.category_code
	LEFT JOIN gowvp_ai_model_category gamc2 ON gami.sub_category = gamc2.category_code
	WHERE 1=1 `
	args := make([]interface{}, 0)
	if mianCategory != "" {
		query += " AND gami.mian_category = ? "
		args = append(args, mianCategory)
	}
	if subCategory != "" {
		query += " AND gami.sub_category = ? "
		args = append(args, subCategory)
	}
	if id != "" {
		query += " AND gami.id LIKE ? "
		args = append(args, "%"+id+"%")
	}
	if name != "" {
		query += " AND gami.name LIKE ? "
		args = append(args, "%"+name+"%")
	}
	if status != "" {
		query += " AND gami.status = ? "
		args = append(args, status)
	}
	if modelPlatform != "" {
		query += " AND gami.model_platform = ? "
		args = append(args, modelPlatform)
	}
	if isDefault != "" {
		query += " AND gami.is_default = ? "
		args = append(args, isDefault)
	}
	// 排序
	query += " ORDER BY gami.is_default DESC, gami.mian_category, gami.sub_category"
	rows, err := mysql.MysqlDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.AiModelInfo
	for rows.Next() {
		var m model.AiModelInfo
		if err := rows.Scan(&m.ID, &m.Name, &m.MianCategory, &m.SubCategory, &m.MianCategoryName, &m.SubCategoryName,
			&m.FunctionInfo, &m.ModelFileName, &m.ModelFileKey, &m.ModelFileURL, &m.ModelFileMd5, &m.ModelVersion, &m.ModelPlatform,
			&m.ModelLabels, &m.Status, &m.IsDefault, &m.CreateTime, &m.UpdateTime, &m.Remarks); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}

// 查询所有已启用的AI模型列表
func GetAllAiModelList() ([]*model.AiModelInfo, error) {
	query := `SELECT 
			id, 
			name, 
			IFNULL(mian_category, '') AS mianCategory,
			IFNULL(sub_category, '') AS subCategory,
			IFNULL(function_info, '') AS functionInfo,
			IFNULL(model_file_name, '') AS modelFileName,
			IFNULL(model_file_key, '') AS modelFileKey,
			IFNULL(model_file_url, '') AS modelFileUrl,
			IFNULL(model_file_md5, '') AS modelFileMd5,
			IFNULL(model_version, '') AS modelVersion,
			IFNULL(model_platform, '') AS modelPlatform,
			IFNULL(model_labels, '') AS modelLabels,
			status, 
			is_default AS isDefault,
			create_time AS createTime, 
			update_time AS updateTime, 
			IFNULL(remarks, '') AS remarks
		FROM gowvp_ai_model_info 
		WHERE status = 1`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.AiModelInfo
	for rows.Next() {
		var n model.AiModelInfo
		if err := rows.Scan(&n.ID, &n.Name, &n.MianCategory, &n.SubCategory, &n.FunctionInfo,
			&n.ModelFileName, &n.ModelFileKey, &n.ModelFileURL, &n.ModelFileMd5, &n.ModelVersion, &n.ModelPlatform,
			&n.ModelLabels, &n.Status, &n.IsDefault, &n.CreateTime, &n.UpdateTime, &n.Remarks); err != nil {
			return nil, err
		}
		list = append(list, &n)
	}
	return list, nil
}

// 根据子分类查询AI模型列表
func GetAiModelListBySubCategory(subCategory string) ([]*model.AiModelInfo, error) {
	return GetAiModelList("", subCategory, "", "", "", "", "")
}

// 根据主分类查询AI模型列表
func GetAiModelListByMainCategory(mianCategory string) ([]*model.AiModelInfo, error) {
	return GetAiModelList(mianCategory, "", "", "", "", "", "")
}

// 根据 ID 查询 AI 模型
func GetAiModelByID(id string) ([]*model.AiModelInfo, error) {
	return GetAiModelList("", "", id, "", "", "", "")
}

// 根据 name 查询 AI 模型
func GetAiModelByName(name string) []*model.AiModelInfo {
	query := `
		SELECT id, name, mian_category, sub_category, 
		       model_file_name, model_file_url, function_info, 
		       status, create_time, update_time, remarks
		FROM gowvp_ai_model_info ami 
		WHERE ami.name = ?`
	rows, err := mysql.MysqlDB.Query(query, name)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []*model.AiModelInfo
	for rows.Next() {
		var m model.AiModelInfo
		if err := rows.Scan(&m.ID, &m.Name, &m.MianCategory, &m.SubCategory,
			&m.ModelFileName, &m.ModelFileURL, &m.FunctionInfo,
			&m.Status, &m.CreateTime, &m.UpdateTime, &m.Remarks); err != nil {
			return nil
		}
		list = append(list, &m)
	}
	return list
}

// 根据设备id查询未被关联的AI模型列表
func GetUnboundAiModelsListByDeviceId(deviceId string) ([]*model.AiModelInfo, error) {
	query := `
		SELECT id, name, mian_category, sub_category, 
		       model_file_name, model_file_url, function_info, 
		       status, create_time, update_time, remarks
		FROM gowvp_ai_model_info ami 
		WHERE ami.id not in ( SELECT ai_model_id FROM gowvp_device_ai_model_relation WHERE device_id = ? )`
	rows, err := mysql.MysqlDB.Query(query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.AiModelInfo
	for rows.Next() {
		var m model.AiModelInfo
		if err := rows.Scan(&m.ID, &m.Name, &m.MianCategory, &m.SubCategory,
			&m.ModelFileName, &m.ModelFileURL, &m.FunctionInfo,
			&m.Status, &m.CreateTime, &m.UpdateTime, &m.Remarks); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}

// 根据主分类分组查询ai模型总数
func GetAiModelCountByMianCategory() ([]*model.AiModelCountByMianCategory, error) {
	query := `
		SELECT 
		mian_category AS mianCategory, 
		count(*) AS count
		FROM gowvp_ai_model_info
		GROUP BY mian_category`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.AiModelCountByMianCategory
	for rows.Next() {
		var m model.AiModelCountByMianCategory
		if err := rows.Scan(&m.MianCategory, &m.Count); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}

// 更新 AI 模型
func UpdateAiModel(modelInfo *model.AiModelInfo) error {
	query := `
		UPDATE gowvp_ai_model_info 
		SET name = ?, mian_category = ?, sub_category = ?, 
		    model_file_name = ?, model_file_key = ?, model_file_url = ?, 
			model_file_md5 =?, model_version=?, model_platform=?, 
			model_labels=?, function_info = ?, status = ?, is_default=?, update_time = ?, remarks = ?
		WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query,
		modelInfo.Name, modelInfo.MianCategory, modelInfo.SubCategory,
		modelInfo.ModelFileName, modelInfo.ModelFileKey, modelInfo.ModelFileURL,
		modelInfo.ModelFileMd5, modelInfo.ModelVersion, modelInfo.ModelPlatform,
		modelInfo.ModelLabels, modelInfo.FunctionInfo, modelInfo.Status,
		modelInfo.IsDefault, utils.NowInCn(), modelInfo.Remarks, modelInfo.ID)
	return err
}

// 删除 AI 模型
func DeleteAiModel(id string) error {
	query := `DELETE FROM gowvp_ai_model_info WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query, id)
	return err
}
