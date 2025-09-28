package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

func BuildSignature(method, clientId, secretKey, queryUri string) (nonce string, timestamp string, signature string, err error) {
	// nonce: 用 crypto/rand
	nonce, err = generateNonce()
	if err != nil {
		return "", "", "", fmt.Errorf("生成 nonce 失败: %w", err)
	}

	// timestamp: 毫秒时间戳
	timestamp = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)

	headerStr := fmt.Sprintf("client-id=%s&nonce=%s&timestamp=%s", clientId, nonce, timestamp)
	ori := fmt.Sprintf("method=%s,clientId=%s,headerStr=%s,queryUri=%s",
		strings.ToLower(method), clientId, headerStr, queryUri)

	mac := hmac.New(sha256.New, []byte(secretKey))
	_, err = mac.Write([]byte(ori))
	if err != nil {
		return "", "", "", fmt.Errorf("签名写入失败: %w", err)
	}
	signature = hex.EncodeToString(mac.Sum(nil))
	return nonce, timestamp, signature, nil
}


// 返回 [min, max] 范围的随机整数
func cryptoRandInt(min, max int64) (int64, error) {
	if min > max {
		return 0, fmt.Errorf("min 大于 max")
	}
	diff := max - min + 1
	nBig, err := rand.Int(rand.Reader, big.NewInt(diff))
	if err != nil {
		return 0, err
	}
	return nBig.Int64() + min, nil
}

// 生成 10 位随机数字符串
func generateNonce() (string, error) {
	num, err := cryptoRandInt(1000000000, 9999999999)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", num), nil
}