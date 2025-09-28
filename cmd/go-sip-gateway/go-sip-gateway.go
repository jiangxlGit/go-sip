package main

import (
	_ "net/http/pprof"

	"go-sip/api"
	"go-sip/api/middleware"
	"go-sip/db/mysql"
	"go-sip/db/redis"
	"go-sip/logger"
	"go-sip/m"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置、初始化日志、Redis
	m.LoadGatewsyConfig()
	logger.InitLogger(m.GatewayConfig.LogLevel)
	redis.InitGatewayRedisMulti(2, 4)
	// 初始化 mysql
	mysql.GatewayInitMysql()

	// 初始化 Gin
	r := gin.Default()
	r.Use(middleware.Recovery)

	api.GatewayApiInit(r)

	r.Run(m.GatewayConfig.GatewayAPI)
}
