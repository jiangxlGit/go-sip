package middleware

import (
	"strings"

	. "go-sip/common"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"

	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Restful API sign 鉴权
func ApiAuth(c *gin.Context) {
	if c.GetString("msgid") == "" {
		c.Set("msgid", utils.RandString(32))
	}
	Logger.Debug("=== openapi auth ===", zap.String("method", c.Request.Method), zap.String("url", c.Request.URL.Path), zap.Any("query", c.Request.URL.RawQuery))
	if strings.Contains(c.Request.URL.Path, "/zlm/webhook") {
		c.Next()
		return
	}
	// 采用签名方式进行鉴权
	if strings.HasPrefix(c.Request.URL.Path, "/open") {
		if !strings.HasPrefix(c.Request.URL.Path, OpenSipServerInfoURL) {
			AuthSign(c)
		}
	}
	// 登录鉴权
	if strings.HasPrefix(c.Request.URL.Path, "/wvp") {
		JWT(c)
	}
	c.Next()
}

type Auth struct {
	Username string `json:"username" validate:"required,min=1,max=20"`
	Password string `json:"password" validate:"required,min=8,max=20"`
}

func GetAuth(c *gin.Context) {

	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		Logger.Error("read body error", zap.Error(err))
		model.JsonResponseSysERR(c, "read body error")
		return
	}
	req := &Auth{}
	if err := utils.JSONDecode(bodyBytes, &req); err != nil {
		Logger.Error("bind json error", zap.Error(err))
		model.JsonResponseSysERR(c, "invalid request")
		return
	}
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		model.JsonResponseSysERR(c, "账号或密码错误")
		return
	}
	username := req.Username
	password := req.Password

	isExist, err := check(username, password)
	if err != nil || !isExist {
		model.JsonResponseSysERR(c, "账号或密码错误")
		return
	}

	token, err := utils.GenerateToken(username, password)
	if err != nil {
		model.JsonResponseSysERR(c, "登录失败")
		return
	}
	model.JsonResponseSucc(c, token)
}

func check(username, password string) (bool, error) {
	if username == m.WVPConfig.Username && password == m.WVPConfig.Password {
		return true, nil
	}
	return false, errors.New("账号或密码错误")
}
