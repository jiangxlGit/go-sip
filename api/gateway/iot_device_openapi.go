package gateway

import (
	. "go-sip/common"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"

	"fmt"
	"io"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeviceAudioVloume struct {
	DeviceId string  `json:"deviceId"`
	Volume   float64 `json:"volume"`
}

// @Summary		设置设备音量接口
// @Router		/open/device/setDevcieAudioVloume [post]
func DevcieSetAudioVloume(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("read body error", zap.Error(err))
		model.JsonResponseSysERR(c, "read body error")
		return
	}
	req := &DeviceAudioVloume{}
	if err := utils.JSONDecode(bodyBytes, &req); err != nil {
		Logger.Error("bind json error", zap.Error(err))
		model.JsonResponseSysERR(c, "invalid request")
		return
	}
	device_id := req.DeviceId
	volume := req.Volume
	// 不能全部为空
	if device_id == "" {
		Logger.Error("device_id不能为空")
		model.JsonResponseSysERR(c, "device_id不能为空")
		return
	}
	if volume < 100.0 || volume > 300.0 {
		Logger.Error("音量范围100-300")
		model.JsonResponseSysERR(c, "音量范围100-300")
		return
	}

	params := url.Values{}
	params.Add("deviceId", device_id)
	params.Add("volume", fmt.Sprintf("%.2f", volume))

	response := GatewayDeviceGetRequestHandler(c, device_id, DevcieSetAudioVloumeURL, params)
	if response["code"] != m.StatusSucc {
		return
	}
	data := response["data"]
	model.JsonResponseSucc(c, data)

}
