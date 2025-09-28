package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	. "go-sip/logger"
	"go-sip/m"

	"go.uber.org/zap"
)

var MysqlDB *sql.DB

func WvpInitMysql () {
	InitMysql(m.WVPConfig.DataBase.Mysql.User, m.WVPConfig.DataBase.Mysql.Password, 
		m.WVPConfig.DataBase.Mysql.Host, m.WVPConfig.DataBase.Mysql.DBName,
		m.WVPConfig.DataBase.Mysql.MaxOpenConns, m.WVPConfig.DataBase.Mysql.MaxIdleConns, 
		m.WVPConfig.DataBase.Mysql.ConnMaxLifetime)
}

func GatewayInitMysql () {
	InitMysql(m.GatewayConfig.DataBase.Mysql.User, m.GatewayConfig.DataBase.Mysql.Password, 
		m.GatewayConfig.DataBase.Mysql.Host, m.GatewayConfig.DataBase.Mysql.DBName, 
		m.GatewayConfig.DataBase.Mysql.MaxOpenConns, m.GatewayConfig.DataBase.Mysql.MaxIdleConns,
		 m.GatewayConfig.DataBase.Mysql.ConnMaxLifetime)
}

func InitMysql(user, password, host, dbName string, maxOpenConns, maxIdleConns, connMaxLifetime int) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
		user, password, host, dbName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		Logger.Error("mysql connect error", zap.Error(err))
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		Logger.Error("mysql ping error", zap.Error(err))
		return nil, err
	}

	MysqlDB = db
	Logger.Info("mysql connect success")
	return db, nil
}
