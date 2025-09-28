package dao

import (
	"database/sql"
	"errors"
	"go-sip/db/mysql"
	"go-sip/model"
	"strings"
	"time"
)

// 插入摄像头信息
func CreateIpcInfo(ipc *model.IpcInfo) error {
	query := `
		INSERT INTO gowvp_ipc_info (
			ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer, 
			transport, stream_type, status, active_time, sip_id, 
			last_heartbeat_time, inner_ip, nogb_username, nogb_password
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := mysql.MysqlDB.Exec(query,
		ipc.IpcId, ipc.IpcIP, ipc.IpcName, ipc.DeviceID, ipc.ChannelId, ipc.Manufacturer,
		ipc.Transport, ipc.StreamType, ipc.Status, ipc.ActiveTime, ipc.SipId,
		ipc.LastHeartbeatTime, ipc.InnerIP, ipc.NogbUsername, ipc.NogbPassword,
	)
	return err
}

// 批量更新或插入ipc信息
func BatchUpsertIpcInfo(ipcList []*model.IpcInfo) error {
	if len(ipcList) == 0 {
		return nil
	}

	var (
		valueStrings []string
		valueArgs    []interface{}
		updateFields = []string{}
	)

	// 构造通用 INSERT 部分
	for _, ipc := range ipcList {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			ipc.IpcId, ipc.IpcIP, ipc.IpcName, ipc.DeviceID, ipc.ChannelId, ipc.Manufacturer,
			ipc.Transport, ipc.StreamType, ipc.Status, ipc.ActiveTime, ipc.SipId,
			ipc.LastHeartbeatTime, ipc.InnerIP, ipc.NogbUsername, ipc.NogbPassword,
		)
	}

	// 用第一条记录判断更新字段（仅用于构造 updateFields，所有记录字段结构应一致）
	sample := ipcList[0]
	updateFields = append(updateFields, "ipc_ip = VALUES(ipc_ip)")
	if sample.IpcName != "" {
		updateFields = append(updateFields, "ipc_name = VALUES(ipc_name)")
	}
	if sample.DeviceID != "" {
		updateFields = append(updateFields, "device_id = VALUES(device_id)")
	}
	if sample.ChannelId != "" {
		updateFields = append(updateFields, "channel_id = VALUES(channel_id)")
	}
	if sample.Manufacturer != "" {
		updateFields = append(updateFields, "manufacturer = VALUES(manufacturer)")
	}
	if sample.Transport != "" {
		updateFields = append(updateFields, "transport = VALUES(transport)")
	}
	if sample.StreamType != "" {
		updateFields = append(updateFields, "stream_type = VALUES(stream_type)")
	}
	if sample.Status != "" {
		updateFields = append(updateFields, "status = VALUES(status)")
	}
	if sample.ActiveTime > 0 {
		updateFields = append(updateFields, "active_time = VALUES(active_time)")
	}
	if sample.SipId != "" {
		updateFields = append(updateFields, "sip_id = VALUES(sip_id)")
	}
	if sample.LastHeartbeatTime > 0 {
		updateFields = append(updateFields, "last_heartbeat_time = VALUES(last_heartbeat_time)")
	}
	if sample.InnerIP != "" {
		updateFields = append(updateFields, "inner_ip = VALUES(inner_ip)")
	}
	if sample.NogbUsername != "" {
		updateFields = append(updateFields, "nogb_username = VALUES(nogb_username)")
	}
	if sample.NogbPassword != "" {
		updateFields = append(updateFields, "nogb_password = VALUES(nogb_password)")
	}

	if len(updateFields) == 0 {
		// 没有需要更新的字段，说明是纯插入
		query := `
			INSERT INTO gowvp_ipc_info (
				ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
				transport, stream_type, status, active_time, sip_id,
				last_heartbeat_time, inner_ip, nogb_username, nogb_password
			) VALUES ` + strings.Join(valueStrings, ",")
		_, err := mysql.MysqlDB.Exec(query, valueArgs...)
		return err
	}

	// 构造完整 SQL
	query := `
		INSERT INTO gowvp_ipc_info (
			ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
			transport, stream_type, status, active_time, sip_id,
			last_heartbeat_time, inner_ip, nogb_username, nogb_password
		) VALUES ` + strings.Join(valueStrings, ",") + `
		ON DUPLICATE KEY UPDATE ` + strings.Join(updateFields, ", ")

	_, err := mysql.MysqlDB.Exec(query, valueArgs...)
	return err
}

func GetFullIpcList() ([]model.IpcInfo, error) {
	query := `
		SELECT ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info
	`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IpcInfo
	for rows.Next() {
		var ipc model.IpcInfo
		if err := rows.Scan(
			&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
			&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
			&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
		); err != nil {
			return nil, err
		}
		list = append(list, ipc)
	}
	return list, nil
}

func GetAllIpcList(deviceId string) ([]model.IpcInfo, error) {
	return GetIpcList(deviceId, "")
}
func GetAllNogbIpcList(deviceId string) ([]model.IpcInfo, error) {
	return GetIpcList(deviceId, "yes")
}

// 查询ipc列表
func GetIpcList(deviceId, gb string) ([]model.IpcInfo, error) {
	query := `
		SELECT ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info
		WHERE device_id = ?
	`
	switch gb {
	case "no":
		query += " AND inner_ip = ''"
	case "yes":
		query += " AND inner_ip != ''"
	}
	rows, err := mysql.MysqlDB.Query(query, deviceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IpcInfo
	for rows.Next() {
		var ipc model.IpcInfo
		if err := rows.Scan(
			&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
			&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
			&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
		); err != nil {
			return nil, err
		}
		list = append(list, ipc)
	}
	return list, nil
}

// 根据manufacturer查询非国标ipc列表
func GetIpcListByManufacturer(manufacturer string) ([]model.IpcInfo, error) {
	query := `
		SELECT ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info
		WHERE manufacturer = ? AND inner_ip != ''
	`
	rows, err := mysql.MysqlDB.Query(query, manufacturer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IpcInfo
	for rows.Next() {
		var ipc model.IpcInfo
		if err := rows.Scan(
			&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
			&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
			&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
		); err != nil {
			return nil, err
		}
		list = append(list, ipc)
	}
	return list, nil
}

func GetIpcInfoTotal(deviceId, gb string) int64 {
	query := `SELECT COUNT(*) FROM gowvp_ipc_info
	WHERE device_id = ?`
	switch gb {
	case "no":
		query += " AND inner_ip = ''"
	case "yes":
		query += " AND inner_ip != ''"
	}
	row := mysql.MysqlDB.QueryRow(query, deviceId)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0
	}
	return count
}

func GetPageIpcInfo(pageNum, pageSize int, deviceId, gb string) ([]model.IpcInfo, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20 // 默认值或限制上限
	}
	offset := (pageNum - 1) * pageSize

	query := `
		SELECT ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info
		WHERE device_id = ?
	`
	switch gb {
	case "no":
		query += " AND inner_ip = ''"
	case "yes":
		query += " AND inner_ip != ''"
	}
	query += " ORDER BY device_id, ipc_id LIMIT ? OFFSET ? "
	rows, err := mysql.MysqlDB.Query(query, deviceId, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.IpcInfo
	for rows.Next() {
		var ipc model.IpcInfo
		if err := rows.Scan(
			&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
			&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
			&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
		); err != nil {
			return nil, err
		}
		list = append(list, ipc)
	}
	return list, nil
}

// 根据 设备id和内网ip 查询摄像头
func GetIpcInfoByDeviceIdAndInnerIp(deviceId, innerIp string) (*model.IpcInfo, error) {
	query := `
		SELECT ipc_id, ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info WHERE device_id = ? AND inner_ip = ?
	`
	row := mysql.MysqlDB.QueryRow(query, deviceId, innerIp)
	var ipc model.IpcInfo
	if err := row.Scan(
		&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
		&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
		&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ipc, nil
}

// 更新在线ipc的状态
func UpdateIpcStatus(deviceId, ipcId, status string) error {
	query := `
	update gowvp_ipc_info set last_heartbeat_time = ?, status = ?
	where device_id = ? and ipc_id = ?
	`
	_, err := mysql.MysqlDB.Exec(query, time.Now().Unix(), deviceId, ipcId, status)
	return err
}

// 根据 ID 查询摄像头
func GetIpcInfoByIpcId(ipcID string) (*model.IpcInfo, error) {
	query := `
		SELECT ipc_id,ipc_ip, ipc_name, device_id, channel_id, manufacturer,
		       transport, stream_type, status, active_time, sip_id,
		       last_heartbeat_time, inner_ip, nogb_username, nogb_password
		FROM gowvp_ipc_info WHERE ipc_id = ?
	`
	row := mysql.MysqlDB.QueryRow(query, ipcID)
	var ipc model.IpcInfo
	if err := row.Scan(
		&ipc.IpcId, &ipc.IpcIP, &ipc.IpcName, &ipc.DeviceID, &ipc.ChannelId, &ipc.Manufacturer,
		&ipc.Transport, &ipc.StreamType, &ipc.Status, &ipc.ActiveTime, &ipc.SipId,
		&ipc.LastHeartbeatTime, &ipc.InnerIP, &ipc.NogbUsername, &ipc.NogbPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ipc, nil
}

// 更新摄像头信息
func UpdateIpcInfoSelective(ipc *model.IpcInfo) error {
	if ipc.IpcId == "" {
		return errors.New("ipc_id is required")
	}

	query := "UPDATE gowvp_ipc_info SET "
	var args []interface{}
	var sets []string

	if ipc.IpcName != "" {
		sets = append(sets, "ipc_name = ?")
		args = append(args, ipc.IpcName)
	}
	if ipc.IpcIP != "" {
		sets = append(sets, "ipc_ip = ?")
		args = append(args, ipc.IpcIP)
	}
	if ipc.DeviceID != "" {
		sets = append(sets, "device_id = ?")
		args = append(args, ipc.DeviceID)
	}
	if ipc.ChannelId != "" {
		sets = append(sets, "channel_id = ?")
		args = append(args, ipc.ChannelId)
	}
	if ipc.Manufacturer != "" {
		sets = append(sets, "manufacturer = ?")
		args = append(args, ipc.Manufacturer)
	}
	if ipc.Transport != "" {
		sets = append(sets, "transport = ?")
		args = append(args, ipc.Transport)
	}
	if ipc.StreamType != "" {
		sets = append(sets, "stream_type = ?")
		args = append(args, ipc.StreamType)
	}
	if ipc.Status != "" {
		sets = append(sets, "status = ?")
		args = append(args, ipc.Status)
	}
	if ipc.ActiveTime > 0 {
		sets = append(sets, "active_time = ?")
		args = append(args, ipc.ActiveTime)
	}
	if ipc.SipId != "" {
		sets = append(sets, "sip_id = ?")
		args = append(args, ipc.SipId)
	}
	if ipc.LastHeartbeatTime > 0 {
		sets = append(sets, "last_heartbeat_time = ?")
		args = append(args, ipc.LastHeartbeatTime)
	}
	if ipc.InnerIP != "" {
		sets = append(sets, "inner_ip = ?")
		args = append(args, ipc.InnerIP)
	}
	if ipc.NogbUsername != "" {
		sets = append(sets, "nogb_username = ?")
		args = append(args, ipc.NogbUsername)
	}
	if ipc.NogbPassword != "" {
		sets = append(sets, "nogb_password = ?")
		args = append(args, ipc.NogbPassword)
	}

	if len(sets) == 0 {
		return errors.New("no fields to update")
	}

	query += strings.Join(sets, ", ") + " WHERE ipc_id = ?"
	args = append(args, ipc.IpcId)

	_, err := mysql.MysqlDB.Exec(query, args...)
	return err
}

// 删除摄像头信息
func DeleteIpcInfo(ipcID string) error {
	ipcInfo, err := GetIpcInfoByIpcId(ipcID)
	if err != nil || ipcInfo == nil {
		return err
	}
	// if ipcInfo.InnerIP == "" {
	// 	return errors.New("国标摄像头不能删除")
	// }
	query := `DELETE FROM gowvp_ipc_info WHERE ipc_id = ?`
	_, err = mysql.MysqlDB.Exec(query, ipcID)
	return err
}
