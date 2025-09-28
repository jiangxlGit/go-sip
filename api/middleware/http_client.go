package middleware

import (
	"net/http"
	"time"
)

var HttpClient *http.Client

func GetHttpClient(clientId, secretKey string) *http.Client {
    client := &http.Client{
		Transport: &SignatureTransport{
			ClientId:  clientId,
			SecretKey: secretKey,
			Base: http.DefaultTransport,
		},
		Timeout: 10 * time.Second,
	}
	return client
}