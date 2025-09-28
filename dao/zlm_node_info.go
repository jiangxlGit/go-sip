package dao

import (
	"database/sql"
	"errors"
	"go-sip/db/mysql"
	"go-sip/model" // 替换为你项目实际路径
)

// 创建节点
func CreateZlmNode(node *model.ZlmNodeInfo) (int64, error) {
	query := `
		INSERT INTO gowvp_zlm_node_info (
			zlm_ip, zlm_port, zlm_domain, zlm_secret, 
			zlm_node_region, region_code, zlm_node_status, remarks
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := mysql.MysqlDB.Exec(query, node.ZlmIP, node.ZlmPort, node.ZlmDomain, node.ZlmSecret,
		node.ZlmNodeRegion, node.RegionCode, node.ZlmNodeStatus, node.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// 查询全部节点
func GetAllZlmNodes() ([]model.ZlmNodeInfo, error) {
	query := `SELECT id, zlm_ip, zlm_port, zlm_domain, zlm_secret, 
	zlm_node_region, region_code, zlm_node_status, IFNULL(remarks, '') 
	FROM gowvp_zlm_node_info`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ZlmNodeInfo
	for rows.Next() {
		var n model.ZlmNodeInfo
		if err := rows.Scan(&n.ID, &n.ZlmIP, &n.ZlmPort, &n.ZlmDomain, &n.ZlmSecret,
			&n.ZlmNodeRegion, &n.RegionCode, &n.ZlmNodeStatus, &n.Remarks); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

// 根据 ID 查询节点
func GetZlmNodeByID(id string) (*model.ZlmNodeInfo, error) {
	query := `SELECT id, zlm_ip, zlm_port, zlm_domain, zlm_secret, zlm_node_region, 
	region_code, zlm_node_status, IFNULL(remarks, '') FROM gowvp_zlm_node_info WHERE id = ?`
	row := mysql.MysqlDB.QueryRow(query, id)

	var n model.ZlmNodeInfo
	if err := row.Scan(&n.ID, &n.ZlmIP, &n.ZlmPort, &n.ZlmDomain, &n.ZlmSecret,
		&n.ZlmNodeRegion, &n.RegionCode, &n.ZlmNodeStatus, &n.Remarks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// 根据zlm_ip+zlm_port 查询节点
func GetZlmNodeByZlmIPAndZlmPort(zlmIP string, zlmPort int) (*model.ZlmNodeInfo, error) {
	query := `SELECT id, zlm_ip, zlm_port, zlm_domain, zlm_secret, zlm_node_region, 
	region_code, zlm_node_status, IFNULL(remarks, '') FROM gowvp_zlm_node_info WHERE zlm_ip = ? AND zlm_port = ?`
	row := mysql.MysqlDB.QueryRow(query, zlmIP, zlmPort)
	var n model.ZlmNodeInfo
	if err := row.Scan(&n.ID, &n.ZlmIP, &n.ZlmPort, &n.ZlmDomain, &n.ZlmSecret,
		&n.ZlmNodeRegion, &n.RegionCode, &n.ZlmNodeStatus, &n.Remarks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

// 更新节点
func UpdateZlmNode(node *model.ZlmNodeInfo) error {
	query := `
		UPDATE gowvp_zlm_node_info 
		SET zlm_ip = ?, zlm_port = ?, zlm_domain = ?, zlm_secret = ?, 
		    zlm_node_region = ?, region_code = ?, zlm_node_status = ?, remarks = ?
		WHERE id = ?
	`
	_, err := mysql.MysqlDB.Exec(query, node.ZlmIP, node.ZlmPort, node.ZlmDomain, node.ZlmSecret,
		node.ZlmNodeRegion, node.RegionCode, node.ZlmNodeStatus, node.Remarks, node.ID)
	return err
}

// 删除节点
func DeleteZlmNode(id string) error {
	query := `DELETE FROM gowvp_zlm_node_info WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query, id)
	return err
}

// 根据 region_code 查询节点列表
func GetZlmNodesByCode(regionCode string) ([]model.ZlmNodeInfo, error) {
	query := `SELECT id, zlm_ip, zlm_port, zlm_domain, zlm_secret, zlm_node_region, 
	region_code, zlm_node_status, IFNULL(remarks, '') FROM gowvp_zlm_node_info 
	WHERE zlm_node_status = 'enable' and region_code = ?`
	rows, err := mysql.MysqlDB.Query(query, regionCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ZlmNodeInfo
	for rows.Next() {
		var n model.ZlmNodeInfo
		if err := rows.Scan(&n.ID, &n.ZlmIP, &n.ZlmPort, &n.ZlmDomain, &n.ZlmSecret,
			&n.ZlmNodeRegion, &n.RegionCode, &n.ZlmNodeStatus, &n.Remarks); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}
