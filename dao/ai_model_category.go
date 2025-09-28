package dao

import (
	"database/sql"
	"errors"
	"go-sip/db/mysql"
	. "go-sip/model"
)

func InsertAiModelCategory(c *AiModelCategory) (int64, error) {
    query := `
        INSERT INTO gowvp_ai_model_category
        (category_name, category_code, parent_code, remarks)
        VALUES (?, ?, ?, ?)
    `
    res, err := mysql.MysqlDB.Exec(query, c.CategoryName, 
		c.CategoryCode, c.ParentCode, c.Remarks)
    if err != nil {
        return 0, err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
    return id, nil
}

func GetAiModelCategoryByID(id int64) (*AiModelCategory, error) {
    query := `
        SELECT id, category_name, category_code, parent_code, IFNULL(remarks, '')
        FROM gowvp_ai_model_category
        WHERE id = ?
    `
    row := mysql.MysqlDB.QueryRow(query, id)
    var c AiModelCategory
    if err := row.Scan(&c.ID, &c.CategoryName, &c.CategoryCode, &c.ParentCode, &c.Remarks); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &c, nil
}

func GetAiModelCategorieListByParentCode(parentCode, name string) ([]*AiModelCategoryRep, error) {
	if parentCode == "" {
		parentCode = "000"
	}
	query := `
	SELECT
		gamc.id AS id,
		gamc.category_name AS categoryName,
		gamc.category_code AS categoryCode,
		gamc.parent_code AS parentCode,
            IFNULL((SELECT gamc2.category_name FROM 
            gowvp_ai_model_category gamc2
            WHERE gamc2.category_code = ?), '') 
        AS parentCategoryName,
		IFNULL(gamc.remarks, '') AS remarks,
		COUNT(gami.id) as count
	FROM
		gowvp_ai_model_category gamc
    	LEFT JOIN gowvp_ai_model_info gami ON gamc.category_code = gami.mian_category
	WHERE gamc.parent_code = ? AND gamc.category_name LIKE ?
	GROUP BY gamc.category_code
	ORDER BY gamc.category_code
	`
    rows, err := mysql.MysqlDB.Query(query, parentCode, parentCode, "%"+name+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []*AiModelCategoryRep
    for rows.Next() {
        var c AiModelCategoryRep
        if err := rows.Scan(&c.ID, &c.CategoryName, &c.CategoryCode, &c.ParentCode, 
            &c.ParentCategoryName, &c.Remarks, &c.Count); err != nil {
            return nil, err
        }
        list = append(list, &c)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }
    return list, nil
}

func GetAiModelCategorieListByCode(code string) (*AiModelCategory, error) {
	query := `
	SELECT 
		id, 
		category_name, 
		category_code, 
		parent_code, 
		IFNULL(remarks, '')
	FROM 
		gowvp_ai_model_category 
	WHERE category_code = ?
	ORDER BY id DESC 
	`
    row := mysql.MysqlDB.QueryRow(query, code)
    var c AiModelCategory
    if err := row.Scan(&c.ID, &c.CategoryName, &c.CategoryCode, &c.ParentCode, &c.Remarks); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &c, nil
}

func GetAiModelCategorieListByName(name string) (*AiModelCategory, error) {
	query := `
	SELECT 
		id, 
		category_name, 
		category_code, 
		parent_code, 
		IFNULL(remarks, '')
	FROM 
		gowvp_ai_model_category 
	WHERE category_name = ?
	ORDER BY id DESC 
	`
    row := mysql.MysqlDB.QueryRow(query, name)
    var c AiModelCategory
    if err := row.Scan(&c.ID, &c.CategoryName, &c.CategoryCode, &c.ParentCode, &c.Remarks); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &c, nil
}


func GetAiModelCountByMianAndSubCategory(mianCategory, subCategory string) int64 {
    query := `
        SELECT COUNT(*)
        FROM gowvp_ai_model_info
        WHERE mian_category = ? AND sub_category = ?
    `
    var count int64
    err := mysql.MysqlDB.QueryRow(query, mianCategory, subCategory).Scan(&count)
    if err != nil {
        return 0
    }
    return count
}

func UpdateAiModelCategory(c *AiModelCategory) error {
    query := `
        UPDATE gowvp_ai_model_category
        SET category_name = ?, category_code = ?, parent_code = ?, remarks = ?
        WHERE id = ?
    `
    _, err := mysql.MysqlDB.Exec(query, c.CategoryName, c.CategoryCode, c.ParentCode, c.Remarks, c.ID)
    return err
}


func DeleteAiModelCategory(id int64) error {
    query := `DELETE FROM gowvp_ai_model_category WHERE id = ?`
    _, err := mysql.MysqlDB.Exec(query, id)
    return err
}