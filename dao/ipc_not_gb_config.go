package dao

import (
	"database/sql"
	"errors"
	"go-sip/db/mysql"
	"go-sip/model"
)

// 新增非国标配置
func CreateNotGBConfig(cfg *model.NotGBConfig) (int64, error) {
	query := `
		INSERT INTO gowvp_not_gb_config (
			manufacturer, rtsp_sub_suffix, rtsp_main_suffix
		) VALUES (?, ?, ?)
	`
	res, err := mysql.MysqlDB.Exec(query,
		cfg.Manufacturer, cfg.RtspSubSuffix, cfg.RtspMainSuffix,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// 查询所有配置
func GetAllNotGBConfigs() ([]model.NotGBConfig, error) {
	query := `
		SELECT id, manufacturer, rtsp_sub_suffix, rtsp_main_suffix
		FROM gowvp_not_gb_config
	`
	rows, err := mysql.MysqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.NotGBConfig
	for rows.Next() {
		var cfg model.NotGBConfig
		if err := rows.Scan(
			&cfg.ID, &cfg.Manufacturer,
			&cfg.RtspSubSuffix, &cfg.RtspMainSuffix,
		); err != nil {
			return nil, err
		}
		list = append(list, cfg)
	}
	return list, nil
}

// 根据 ID 查询配置
func GetNotGBConfigByID(id int64) (*model.NotGBConfig, error) {
	query := `
		SELECT id, manufacturer, rtsp_sub_suffix, rtsp_main_suffix
		FROM gowvp_not_gb_config WHERE id = ?
	`
	row := mysql.MysqlDB.QueryRow(query, id)
	var cfg model.NotGBConfig
	if err := row.Scan(
		&cfg.ID, &cfg.Manufacturer,
		&cfg.RtspSubSuffix, &cfg.RtspMainSuffix,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// 根据 manufacturer 查询配置（唯一）
func GetNotGBConfigByManufacturer(manufacturer string) (*model.NotGBConfig, error) {
	query := `
		SELECT id, manufacturer, rtsp_sub_suffix, rtsp_main_suffix
		FROM gowvp_not_gb_config WHERE manufacturer = ?
	`
	row := mysql.MysqlDB.QueryRow(query, manufacturer)

	var cfg model.NotGBConfig
	if err := row.Scan(
		&cfg.ID, &cfg.Manufacturer,
		&cfg.RtspSubSuffix, &cfg.RtspMainSuffix,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// 更新配置
func UpdateNotGBConfig(cfg *model.NotGBConfig) error {
	query := `
		UPDATE gowvp_not_gb_config
		SET manufacturer = ?, rtsp_sub_suffix = ?, rtsp_main_suffix = ?
		WHERE id = ?
	`
	_, err := mysql.MysqlDB.Exec(query,
		cfg.Manufacturer,
		cfg.RtspSubSuffix, cfg.RtspMainSuffix, cfg.ID,
	)
	return err
}

// 删除配置
func DeleteNotGBConfig(id int64) error {
	query := `DELETE FROM gowvp_not_gb_config WHERE id = ?`
	_, err := mysql.MysqlDB.Exec(query, id)
	return err
}
