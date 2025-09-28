package main

import (
	_ "net/http/pprof"

	"go-sip/api"
	"go-sip/api/middleware"
	wvp "go-sip/api/wvp"
	"go-sip/db/alioss"
	"go-sip/db/mysql"
	"go-sip/db/redis"
	_ "go-sip/docs"
	"go-sip/logger"
	"go-sip/m"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func main() {
	m.LoadWvpConfig()
	logger.InitLogger(m.WVPConfig.LogLevel)
	// 初始化wvp redis
	redis.InitWvpRedisMulti(2, 4)

	// 初始化 mysql
	mysql.WvpInitMysql()

	// 初始化alioss
	alioss.WvpInitAliOSS()

	// 初始化ipc列表
	wvp.IpcInfoInit()

	r := gin.Default()
	r.Use(middleware.Recovery)

	wvp.ZlmNodeInfoInit()
	wvp.ZlmNodeRegionInfoInit()
	api.WvpApiInit(r)

	err := r.Run(m.WVPConfig.API)
	if err != nil {
		logger.Logger.Error("go wvp启动失败", zap.Error(err))
		return
	}

}
