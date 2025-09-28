package dao

import (
	"database/sql"
	"go-sip/db/mysql"
	"go-sip/model"
)

// 查询iot设备列表(返回ai模型总数)
func GetIotDeviceList2(iotDeviceId string) ([]model.IotDeviceAndAiModelInfo, error) {
	var (
		query string
		rows  *sql.Rows
		err   error
	)

	if iotDeviceId == "" {
		query = `
			SELECT id, IFNULL(state, 'offline') AS state
			FROM dev_device_instance
			ORDER BY state DESC
		`
		rows, err = mysql.MysqlDB.Query(query)
	} else {
		query = `
			SELECT id, IFNULL(state, 'offline') AS state
			FROM dev_device_instance
			WHERE id = ?
		`
		rows, err = mysql.MysqlDB.Query(query, iotDeviceId)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IotDeviceAndAiModelInfo
	for rows.Next() {
		var n model.IotDeviceAndAiModelInfo
		if err := rows.Scan(&n.IotDeviceID, &n.State); err != nil {
			return nil, err
		}
		// 获取已启用的ai模型总数
		n.AiModelCount = GetAiModelCount()
		// 根据设备id获取已关联的ai模型总数
		n.RelationAiModelCount = QueryDeviceAiModelRelationCount(n.IotDeviceID)
		list = append(list, n)
	}

	// 这里也要检查 rows.Err()，防止隐藏的遍历错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// 查询iot设备列表(返回摄像头总数)
func GetIotDeviceList(iotDeviceId string) ([]*model.IotDeviceAndIpcInfo, error) {
	var (
		query string
		rows  *sql.Rows
		err   error
	)

	if iotDeviceId == "" {
		query = `
			SELECT id, IFNULL(state, 'offline') AS state
			FROM dev_device_instance
			ORDER BY state DESC
		`
		rows, err = mysql.MysqlDB.Query(query)
	} else {
		query = `
			SELECT id, IFNULL(state, 'offline') AS state
			FROM dev_device_instance
			WHERE id = ?
		`
		rows, err = mysql.MysqlDB.Query(query, iotDeviceId)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.IotDeviceAndIpcInfo
	for rows.Next() {
		var n model.IotDeviceAndIpcInfo
		if err := rows.Scan(&n.IotDeviceID, &n.State); err != nil {
			return nil, err
		}
		// 查询ipc列表
		ipcInfoList, err := GetAllIpcList(n.IotDeviceID)
		if err != nil {
			n.IpcCount = 0
			continue
		}
		n.IpcCount = len(ipcInfoList)
		list = append(list, &n)
	}

	// 这里也要检查 rows.Err()，防止隐藏的遍历错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// 查询未关联和已关联某个ai模型的中控设备列表
func GetIotDevicePageByAiModel(aiModelId string, page, pageSize int) ([]*model.IotDeviceSimpleInfo, error) {
	query := `
		SELECT 
			d.id AS id, 
			d.name AS name, 
			IFNULL(d.state, '') AS state,
			IF(r.device_id is null, 'no', 'yes') AS relation
		FROM 
			dev_device_instance d
		LEFT JOIN 
			gowvp_device_ai_model_relation r
			ON d.id = r.device_id AND r.ai_model_id = ?
		LIMIT ?, ?
	`
	offset := (page - 1) * pageSize
	rows, err := mysql.MysqlDB.Query(query, aiModelId, offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.IotDeviceSimpleInfo
	for rows.Next() {
		var device model.IotDeviceSimpleInfo
		if err := rows.Scan(&device.ID, &device.Name, &device.State, &device.Relation); err != nil {
			return nil, err
		}
		list = append(list, &device)
	}
	return list, nil
}

// 查询未关联和已关联某个ai模型的中控设备总数
func GetIotDeviceCountByAiModel(aiModelId string) int64 {
	query := `
		SELECT 
			count(*) AS count
		FROM 
			dev_device_instance d
		LEFT JOIN 
			gowvp_device_ai_model_relation r
			ON d.id = r.device_id AND r.ai_model_id = ?
	`
	var count int64
	err := mysql.MysqlDB.QueryRow(query, aiModelId).Scan(&count)
	if err != nil || count < 0 {
		return 0
	}
	return count
}
