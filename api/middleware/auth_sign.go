package middleware

import (
	"go-sip/db/redis"
	. "go-sip/logger"
	"go-sip/utils"

	"fmt"
	"net/http"
	"strings"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"strconv"

	"facette.io/natsort"

	redis_util "go-sip/db/redis/redis_gateway_util"
)

func AuthSign(c *gin.Context) {

	method := c.Request.Method
	uri := c.Request.URL.Path
	contentType := c.GetHeader("Content-Type")

	// 处理空 contentType 的情况（通常是 GET）
	if contentType == "" {
		if method != http.MethodGet {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "HTY_IOT_FAIL_CODE_0001",
				"message": "非法访问",
			})
			return
		}
		CheckSign(uri, c) // 签名校验
	} else {
		mediaType := strings.Split(contentType, ";")[0] // 去掉 charset 等参数
		if mediaType == "application/json" || mediaType == "application/x-www-form-urlencoded" {
			CheckSign(uri, c) // 签名校验
		}
	}

	c.Next()

}

// 签名校验函数
func CheckSign(uri string, c *gin.Context) {

	fmt.Println("执行签名校验，URI =", uri)

	clientId := c.GetHeader("client-id")
	timestamp := c.GetHeader("timestamp")
	nonce := c.GetHeader("nonce")
	sign := c.GetHeader("sign")

	if clientId == "" || timestamp == "" || nonce == "" || sign == "" {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code":    "HTY_IOT_FAIL_CODE_0002",
			"message": "头部参数非法",
		})
		return
	}

	if !CheckTime(timestamp) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code":    "HTY_IOT_FAIL_CODE_0002",
			"message": "时间戳非法或过期",
		})
		return
	}

	if !CheckNonce(c, nonce) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code":    "HTY_IOT_FAIL_CODE_0002",
			"message": "随机数非法",
		})
		return
	}

	// 从 Redis 获取签名密钥
	signSecret, err := redis_util.Get_4(fmt.Sprintf(redis.IOT_OPEN_API_KEY, clientId))
	if err != nil || signSecret == "" {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code":    "HTY_IOT_FAIL_CODE_0003",
			"message": "获取签名密钥失败",
		})
		return
	}

	unquotedSignSecret, err := strconv.Unquote(signSecret)
	if err != nil {
		// 如果不是包裹的引号，会报错，可直接使用原值
		unquotedSignSecret = signSecret
	}

	// 构建签名原文
	headerStr := fmt.Sprintf("client-id=%s&nonce=%s&timestamp=%s", clientId, nonce, timestamp)
	queryUri := uri + GetQueryParam(c.Request.URL.Query())

	realSign := CalculatorSign(clientId, queryUri, c, headerStr, unquotedSignSecret)

	if realSign != sign {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code":    "HTY_IOT_FAIL_CODE_0005",
			"message": "签名错误",
		})
		return
	}
}

// 获取排序后的查询参数字符串，格式如 "?a=1&b=2"
func GetQueryParam(queryParams url.Values) string {
	if len(queryParams) == 0 {
		return ""
	}

	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	natsort.Sort(keys)

	var builder strings.Builder
	builder.WriteString("?")

	for _, key := range keys {
		values := queryParams[key]
		if len(values) == 1 {
			builder.WriteString(key)
			builder.WriteString("=")
			builder.WriteString(values[0])
		} else {
			// 多个值拼成字符串（如 [a b]），可自定义格式
			builder.WriteString(key)
			builder.WriteString("=")
			builder.WriteString(strings.Join(values, ","))
		}
		builder.WriteString("&")
	}

	result := builder.String()
	return strings.TrimSuffix(result, "&")
}

func Sha256HMAC(message, secret string) string {
	defer func() {
		if r := recover(); r != nil {
			Logger.Error("error in Sha256HMAC panic", zap.Any("recover", r))
		}
	}()

	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(message))
	if err != nil {
		Logger.Error("error in Sha256HMAC", zap.Error(err))
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func CalculatorSign(clientId, queryUri string, c *gin.Context, headerStr, signSecret string) string {
	method := strings.ToLower(c.Request.Method)
	ori := fmt.Sprintf("method=%s,clientId=%s,headerStr=%s,queryUri=%s", method, clientId, headerStr, queryUri)
	sign := Sha256HMAC(ori, signSecret)
	return sign
}

func CheckTime(timestamp string) bool {
	timeMillis, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		Logger.Error("CheckTime 解析时间戳失败", zap.String("err", err.Error()))
		return false
	}

	now := time.Now().UnixMilli()
	duration := now - timeMillis
	fiveMinutes := int64(120 * 1000) // 120 秒

	if duration > fiveMinutes || -duration > fiveMinutes {
		nowStr := time.UnixMilli(now).Format("2006-01-02 15:04:05")
		timeStr := time.UnixMilli(timeMillis).Format("2006-01-02 15:04:05")
		Logger.Error("时间戳超时",
			zap.String("now", nowStr),
			zap.String("time", timeStr),
		)
		return false
	}

	return true
}

// CheckNonce 检查随机串 nonce 是否有效
func CheckNonce(c *gin.Context, nonce string) bool {
	if strings.TrimSpace(nonce) == "" {
		return false
	}

	nonce = strings.TrimSpace(nonce)
	if len(nonce) < 10 {
		Logger.Warn("随机串nonce长度最少为10位", zap.String("nonce", nonce))
		return false
	}
	redisNonce, err := redis_util.Get_2(redis.OPEN_API_KEY_NONCE)
	if err != nil {
		// 如果是 key 不存在则允许继续
		if err.Error() != "redis: nil" {
			Logger.Error("获取 nonce 错误", zap.Error(err))
			return false
		}
	}

	if redisNonce == "" {
		// 不存在，说明可以设置
		err := redis_util.Set_2(redis.OPEN_API_KEY_NONCE, nonce, 60*time.Second)
		if err != nil {
			Logger.Error("设置 nonce 错误", zap.Error(err))
			return false
		}
		return true
	} else {
		if redisNonce == nonce {
			Logger.Warn("60秒内不允许重复请求", zap.String("nonce", nonce))
			return false
		}
	}

	return true
}

type SignatureTransport struct {
	ClientId  string
	SecretKey string
	Base      http.RoundTripper
}

func (st *SignatureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	method := req.Method
	queryUri := req.URL.RequestURI()

	nonce, timestamp, signature, err := utils.BuildSignature(method, st.ClientId, st.SecretKey, queryUri)
	if err != nil {
		return nil, fmt.Errorf("签名构建失败: %w", err)
	}

	// 注入签名头
	req.Header.Set("client-id", st.ClientId)
	req.Header.Set("nonce", nonce)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", signature)
	// Logger.Info("签名成功", zap.Any("queryUri", queryUri) ,zap.String("client-id", st.ClientId), zap.String("nonce", nonce), 
	// zap.String("timestamp", timestamp), zap.String("sign", signature))

	return st.Base.RoundTrip(req)
}

