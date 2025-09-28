package dao

import (
	"fmt"
	"go-sip/db/mysql"
	"go-sip/model"
	"go-sip/utils"
	"strings"

	"database/sql"
	"errors"
)

func CreateDeviceAiModelRelation(r *model.DeviceAiModelRelation) (int64, error) {
	insertSql := `
		INSERT INTO gowvp_device_ai_model_relation (
			device_id, ipc_id, ai_model_id, ai_model_confidence, status, create_time,update_time, remarks
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := mysql.MysqlDB.Exec(insertSql, r.DeviceID, r.IpcId,
		r.AiModelID, r.AiModelConfidence, r.Status, utils.NowInCn(), utils.NowInCn(), r.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func BatchInsertDeviceAiModelRelation(relations []*model.DeviceAiModelRelation) error {
	if len(relations) == 0 {
		return nil
	}

	// 动态拼接 VALUES 占位符
	var placeholders []string
	var args []interface{}
	for _, rel := range relations {
		placeholders = append(placeholders, "(?, ?, ?)")
		args = append(args, rel.DeviceID, rel.AiModelID, rel.Remarks)
	}

	// 拼接完整 SQL
	sqlStr := fmt.Sprintf(`
		INSERT INTO gowvp_device_ai_model_relation (device_id, ai_model_id, remarks)
		VALUES %s
		`, strings.Join(placeholders, ", "))

	// 执行
	_, err := mysql.MysqlDB.Exec(sqlStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func GetIpcAiModelRelationByID(id string) (*model.DeviceAiModelRelation, error) {
	selectSql := `SELECT id, device_id, ipc_id, ai_model_id, ai_model_confidence, status, create_time, update_time, IFNULL(remarks, '')
		FROM gowvp_device_ai_model_relation	WHERE id = ?`
	row := mysql.MysqlDB.QueryRow(selectSql, id)
	var r model.DeviceAiModelRelation
	if err := row.Scan(&r.ID, &r.DeviceID, &r.IpcId, &r.AiModelID, &r.AiModelConfidence, &r.Status, &r.CreateTime, &r.UpdateTime, &r.Remarks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func QueryIpcAiModelRelations(ipcId, aiModelID string) ([]model.DeviceAiModelRelation, error) {
	querySql := `
		SELECT id, device_id, ipc_id, ai_model_id, ai_model_confidence, status, create_time, update_time, IFNULL(remarks, '')
		FROM gowvp_device_ai_model_relation WHERE 1=1
	`
	var args []interface{}

	if ipcId != "" {
		querySql += " AND ipc_id = ?"
		args = append(args, ipcId)
	}
	if aiModelID != "" {
		querySql += " AND ai_model_id = ?"
		args = append(args, aiModelID)
	}

	rows, err := mysql.MysqlDB.Query(querySql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DeviceAiModelRelation
	for rows.Next() {
		var r model.DeviceAiModelRelation
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.IpcId, &r.AiModelID, &r.AiModelConfidence, &r.Status, &r.CreateTime, &r.UpdateTime, &r.Remarks); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func QueryDevcieAiModelRelations(deviceId, aiModelID string) ([]model.DeviceAiModelRelation, error) {
	querySql := `
		SELECT id, device_id, ipc_id, ai_model_id, ai_model_confidence, status, create_time, update_time, IFNULL(remarks, '')
		FROM gowvp_device_ai_model_relation WHERE 1=1
	`
	var args []interface{}

	if deviceId != "" {
		querySql += " AND device_id = ?"
		args = append(args, deviceId)
	}
	if aiModelID != "" {
		querySql += " AND ai_model_id = ?"
		args = append(args, aiModelID)
	}

	rows, err := mysql.MysqlDB.Query(querySql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.DeviceAiModelRelation
	for rows.Next() {
		var r model.DeviceAiModelRelation
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.IpcId, &r.AiModelID, &r.AiModelConfidence, &r.Status, &r.CreateTime, &r.UpdateTime, &r.Remarks); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func QueryDeviceAiModelRelationCount(deviceId string) int64 {
	querySql := `
		SELECT count(*) AS count
		FROM gowvp_device_ai_model_relation 
		WHERE device_id = ?
	`
	var count int64
	err := mysql.MysqlDB.QueryRow(querySql, deviceId).Scan(&count)
	if err != nil || count < 0 {
		return 0
	}
	return count
}

func UpdateIpcAiModelRelation(r *model.DeviceAiModelRelation) error {
	updateSql := `
		UPDATE gowvp_device_ai_model_relation
		SET device_id = ?, ipc_id = ?, ai_model_id = ?, ai_model_confidence = ? , status = ?, create_time = ?,update_time = ?,remarks = ?
		WHERE id = ?
	`
	_, err := mysql.MysqlDB.Exec(updateSql, r.DeviceID, r.IpcId, r.AiModelID, r.AiModelConfidence, r.Status, utils.NowInCn(), utils.NowInCn(), r.Remarks, r.ID)
	return err
}

func DeleteIpcAiModelRelationByID(id string) error {
	deleteSql := `DELETE FROM gowvp_device_ai_model_relation WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(deleteSql, id)
	return err
}

func DeleteIpcAiModelRelation(deviceId, ipcId string) error {
	deleteSql := `DELETE FROM gowvp_device_ai_model_relation WHERE device_id = ? AND ipc_id = ?`
	_, err := mysql.MysqlDB.Exec(deleteSql, deviceId, ipcId)
	return err
}

func DeleteDeviceAiModelRelationByAiModelIdAndDeviceId(aiModelId, deviceId string) error {
	deleteSql := `DELETE FROM gowvp_device_ai_model_relation WHERE ai_model_id = ? AND device_id = ?`
	_, err := mysql.MysqlDB.Exec(deleteSql, aiModelId, deviceId)
	return err
}
