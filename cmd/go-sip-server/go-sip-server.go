package main

import (
	"fmt"
	"net"
	_ "net/http/pprof"
	"time"

	"go-sip/api"
	"go-sip/api/middleware"
	sapi "go-sip/api/s"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_server_util"
	_ "go-sip/docs"
	grpc_server "go-sip/grpc_api/s"
	"go-sip/logger"
	"go-sip/m"
	"go-sip/mq/kafka"
	"go-sip/mq/mqtt"
	pb "go-sip/signaling"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.uber.org/zap"
)

func main() {
	m.LoadServerConfig()
	logger.InitLogger(m.SMConfig.LogLevel)
	redis.InitServerRedisMulti(2, 4)
	r := gin.Default()
	r.Use(middleware.Recovery)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	api.ServerApiInit(r)
	// 定时监控ipc状态
	sapi.IpcStatusSync()
	// 初始化kafka
	go kafka.InitKafkaProducer()
	// 初始化mqtt
	go mqtt.InitMqttClient()

	lis, _ := net.Listen("tcp", fmt.Sprintf("%s:%s", m.SMConfig.TcpIp, m.SMConfig.TcpPort))
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 15 * time.Minute, // 空闲连接最大保持时间
			Time:              30 * time.Second, // PING发送间隔（服务端->客户端）
			Timeout:           20 * time.Second, // PING响应超时
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second, // 允许客户端的最小PING间隔
			PermitWithoutStream: true,             // 允许无活跃流的PING
		}),
	)
	sip_server := grpc_server.GetSipServer()
	pb.RegisterSipServiceServer(grpcServer, sip_server)
	go grpcServer.Serve(lis)

	// sip服务id对应sip公网url存入redis
	redis_util.HSet_2(redis.SIP_SERVER_HOST, m.SMConfig.SipID, fmt.Sprintf("%s:%s", m.SMConfig.SipInnerIp, m.SMConfig.SipPort))
	redis_util.HSet_2(redis.SIP_SERVER_PUBLIC_TCP_HOST, m.SMConfig.SipID, fmt.Sprintf("%s:%s", m.SMConfig.SipOutIp, m.SMConfig.TcpPort))

	err := r.Run(m.SMConfig.API)
	if err != nil {
		logger.Logger.Error("sip server启动失败", zap.Error(err))
		return
	}

}
