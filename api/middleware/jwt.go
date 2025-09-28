package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-sip/model"
	"go-sip/utils"
)

func JWT(c *gin.Context) {

	var code int
	var data interface{}

	code = model.StatusSucc
	token := c.GetHeader("token")
	if token == "" {
		code = model.StatusParamsERR
	} else {
		claims, err := utils.ParseToken(token)
		if err != nil {
			code = model.ERROR_AUTH_CHECK_TOKEN_FAIL
		} else if time.Now().Unix() > claims.ExpiresAt {
			code = model.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
		}
	}

	if code != model.StatusSucc {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": code,
			"msg":  "Token鉴权失败",
			"data": data,
		})
		c.Abort()
		return
	}

	c.Next()

}
