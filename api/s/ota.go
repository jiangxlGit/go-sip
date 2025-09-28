package api

import (
	"fmt"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/mq/mqtt"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

// @Summary 设备固件拉取
// @Router /device/ota/firmwarePull [GET]
func OTAFirmwarePull(c *gin.Context) {

	productId := c.Query("productId")
	deviceId := c.Query("deviceId")
	requestVersion := c.Query("requestVersion")
	force := c.Query("force")
	latest := c.Query("latest")

	if productId == "" || deviceId == "" || requestVersion == "" {
		m.JsonResponse(c, m.StatusParamsERR, "productId和deviceId和requestVersion参数都不能为空")
		return
	}

	var forceBool bool
	var latestBool bool

	if force == "" || force == "false" {
		forceBool = false
	} else {
		forceBool = true
	}
	if latest == "" || latest == "false" {
		latestBool = false
	} else {
		latestBool = true
	}

	// ota固件拉取消息发送给kafka
	firmwarePullMqttMsg := mqtt.OTAFirmwarePullMqttMsg{
		DeviceID:       deviceId,
		RequestVersion: requestVersion,
		Headers: mqtt.OTAFirmwarePullHeaders{
			Force:  forceBool,
			Latest: latestBool,
		},
		MessageID: uuid.Must(uuid.NewV4()).String(),
		Timestamp: time.Now().UnixMilli(),
	}
	err := mqtt.SimplePublishMessage(fmt.Sprintf("%s/%s/firmware/pull", productId, deviceId), firmwarePullMqttMsg, 2)
	if err != nil {
		Logger.Error("ota固件拉取消息发送mqtt失败", zap.Any("firmwarePullMqttMsg", firmwarePullMqttMsg), zap.Error(err))
		m.JsonResponse(c, m.StatusSysERR, "发布mqtt消息失败")
		return
	}
	m.JsonResponse(c, m.StatusSucc, "OTA固件拉取发布mqtt成功")

}
