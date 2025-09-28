package gateway

import (
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_gateway_util"
	. "go-sip/logger"
	"go-sip/model"
	"go-sip/utils"
	"strings"

	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		获取sip服务信息接口
// @Router			/open/sip/getSipServerInfo [post]
func GetSipServerInfo(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("read body error", zap.Error(err))
		model.JsonResponseSysERR(c, "read body error")
		return
	}
	req := &model.SipServerInfoBody{}
	if err := utils.JSONDecode(bodyBytes, &req); err != nil {
		Logger.Error("bind json error", zap.Error(err))
		model.JsonResponseSysERR(c, "invalid request")
		return
	}
	device_id := req.DeviceId
	// 不能全部为空
	if device_id == "" {
		Logger.Error("device_id不能为空")
		model.JsonResponseSysERR(c, "device_id不能为空")
		return
	}

	tcp_addr := ""
	// 查出所有sip服务, hash取模获取一个
	sip_server_tcp_map, err := redis_util.HGetAll_2(redis.SIP_SERVER_PUBLIC_TCP_HOST)
	if err != nil || sip_server_tcp_map == nil || len(sip_server_tcp_map) == 0 {
		Logger.Error("未找到任何sip服务")
		model.JsonResponseSysERR(c, "未找到任何sip服务")
		return
	} else {
		sip_server_id, sip_server_tcp_addr, err := SelectPollMapValue("client_invoke_server", FilterLoopbackSipServers(sip_server_tcp_map))
		if err != nil || sip_server_tcp_addr == "" {
			Logger.Error("未找到任何sip服务")
			model.JsonResponseSysERR(c, "未找到任何sip服务")
			return
		} else {
			Logger.Info("选择sip服务", zap.String("sip_server_id", sip_server_id), zap.String("sip_server_tcp_addr", sip_server_tcp_addr))
			tcp_addr = sip_server_tcp_addr
		}
	}
	if strings.Contains(tcp_addr, "127.0.0.1") || strings.Contains(tcp_addr, "localhost") {
		Logger.Error("sip服务地址错误")
		model.JsonResponseSysERR(c, "sip服务地址错误")
		return
	}

	// 返回zlm_info
	model.JsonResponseSucc(c, tcp_addr)
}
