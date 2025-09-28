package dao

import (
	"database/sql"
	"errors"
	"go-sip/db/mysql"
	"go-sip/model" // 替换为你项目中 model 包的实际路径
)

// 创建记录
func CreateRegion(region *model.ZlmRegionInfo) (int64, error) {
	query := `
		INSERT INTO gowvp_zlm_region_info (region_code, region_name, 
		relation_region_code, relation_region_name, remarks)
		VALUES (?, ?, ?, ?, ?)`
	result, err := mysql.MysqlDB.Exec(query, region.RegionCode, region.RegionName, 
		region.RelationRegionCode, region.RelationRegionName, region.Remarks)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// 查询记录（可选条件）
func GetRegions() ([]model.ZlmRegionInfo, error) {
	query := `SELECT id, region_code, region_name, relation_region_code, 
	relation_region_name, IFNULL(remarks, '') FROM gowvp_zlm_region_info`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ZlmRegionInfo
	for rows.Next() {
		var r model.ZlmRegionInfo
		if err := rows.Scan(&r.ID, &r.RegionCode, &r.RegionName, 
			&r.RelationRegionCode, &r.RelationRegionName, &r.Remarks); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// 根据 ID 查询
func GetRegionByID(id string) (*model.ZlmRegionInfo, error) {
	query := `SELECT id, region_code, region_name, relation_region_code, 
	relation_region_name, IFNULL(remarks, '') FROM gowvp_zlm_region_info WHERE id = ?`
	row := mysql.MysqlDB.QueryRow(query, id)

	var r model.ZlmRegionInfo
	if err := row.Scan(&r.ID, &r.RegionCode, &r.RegionName, 
		&r.RelationRegionCode, &r.RelationRegionName, &r.Remarks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// 根据区域编码查询
func GetRegionByCode(regionCode string) (*model.ZlmRegionInfo, error) {
	query := `
		SELECT id, region_code, region_name, 
		       relation_region_code, relation_region_name, 
		       IFNULL(remarks, '') 
		FROM gowvp_zlm_region_info 
		WHERE region_code = ?
		LIMIT 1`

	row := mysql.MysqlDB.QueryRow(query, regionCode)

	var r model.ZlmRegionInfo
	if err := row.Scan(&r.ID, &r.RegionCode, &r.RegionName, 
		&r.RelationRegionCode, &r.RelationRegionName, &r.Remarks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 没有结果不是异常，返回 nil
		}
		return nil, err
	}
	return &r, nil
}

// 根据关联区域编码查询
func GetRegionByRelationRegionCode(relationRegionCode string) ([]model.ZlmRegionInfo, error) {
	query := `
		SELECT id, region_code, region_name, 
		       relation_region_code, relation_region_name, 
		       IFNULL(remarks, '') 
		FROM gowvp_zlm_region_info 
		WHERE relation_region_code = ?
		LIMIT 1`

	rows, err := mysql.MysqlDB.Query(query, relationRegionCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ZlmRegionInfo
	for rows.Next() {
		var n model.ZlmRegionInfo
		if err := rows.Scan(&n.ID, &n.RegionCode, &n.RegionName, 
			&n.RelationRegionCode, &n.RelationRegionName, &n.Remarks); err != nil {
			return nil, err
		}
		list = append(list, n)
	}

	return list, nil
}

// 更新记录
func UpdateRegion(region *model.ZlmRegionInfo) error {
	query := `
		UPDATE gowvp_zlm_region_info 
		SET region_code = ?, region_name = ?, relation_region_code = ?, relation_region_name = ?, remarks = ?
		WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query, region.RegionCode, region.RegionName, 
		region.RelationRegionCode, region.RelationRegionName, region.Remarks, region.ID)
	return err
}

// 删除记录
func DeleteRegion(id string) error {
	query := `DELETE FROM gowvp_zlm_region_info WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query, id)
	return err
}
