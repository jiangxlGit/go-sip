package model

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusSucc      = 200
	StatusParamsERR = 400
	StatusAuthERR   = 401
	StatusSysERR    = 500

	ERROR_AUTH_CHECK_TOKEN_FAIL    = 20001
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT = 20002
)

const (
	CodeSucc   = "success"
	CodeSysERR = "fail"
)

var HTTPCode = map[string]int{
	CodeSucc: http.StatusOK,
}

type ApiResult struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
	Result any    `json:"result"`
}

type IpcAiModelResult struct {
	Code   string                    `json:"code"`
	Status int                       `json:"status"`
	Result map[string][]*AiModelInfo `json:"result"`
}

type IotDeviceRelationAiModelResult struct {
	Code   string                          `json:"code"`
	Status int                             `json:"status"`
	Result []*IotDeviceRelationAiModelInfo `json:"result"`
}

type IotIpcInfoResult struct {
	Code   string     `json:"code"`
	Status int        `json:"status"`
	Result []*IpcInfo `json:"result"`
}
type IotNotGbIpcInfoResult struct {
	Code   string             `json:"code"`
	Status int                `json:"status"`
	Result []*IotNotGbIpcInfo `json:"result"`
}

type PageResult struct {
	Total    int64 `json:"total"`    // 总记录数
	Page     int   `json:"page"`     // 当前页码
	PageSize int   `json:"pageSize"` // 每页大小
	Data     any   `json:"data"`     // 列表数据
}

func JsonResponse(c *gin.Context, code string, status int, data any) {
	switch d := data.(type) {
	case error:
		data = d.Error()
	}
	c.JSON(HTTPCode[code], ApiResult{Code: code, Status: status, Result: data})
}

func JsonResponseSucc(c *gin.Context, data any) {
	JsonResponse(c, CodeSucc, StatusSucc, data)
}

// 分页查询成功后调用
func JsonResponsePageSucc(c *gin.Context, total int64, page int, pageSize int, list any) {
	result := PageResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     list,
	}
	JsonResponseSucc(c, result)
}

func JsonResponseSysERR(c *gin.Context, data any) {
	JsonResponse(c, CodeSysERR, StatusSysERR, data)
}
