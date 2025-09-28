package alioss

import (
	. "go-sip/logger"
	"go-sip/m"

	"sync"

	gooss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

// AliyunOSS 封装 OSS 客户端
type AliyunOSS struct {
	Client *gooss.Client
	Bucket *gooss.Bucket
}

var (
	once      sync.Once
	globalOSS *AliyunOSS
	initErr   error
)

func SipClientInitAliOSS() {
	endpoint := m.CMConfig.AliYunOss.Endpoint
	accessKeyID := m.CMConfig.AliYunOss.AccessKeyID
	accessKeySecret := m.CMConfig.AliYunOss.AccessKeySecret
	bucketName := m.CMConfig.AliYunOss.BucketName
	InitAliOSS(endpoint, accessKeyID, accessKeySecret, bucketName)
}
func WvpInitAliOSS() {
	endpoint := m.WVPConfig.AliYunOss.Endpoint
	accessKeyID := m.WVPConfig.AliYunOss.AccessKeyID
	accessKeySecret := m.WVPConfig.AliYunOss.AccessKeySecret
	bucketName := m.WVPConfig.AliYunOss.BucketName
	InitAliOSS(endpoint, accessKeyID, accessKeySecret, bucketName)
}

// InitOSS 初始化全局 OSS 客户端（只执行一次）
func InitAliOSS(endpoint, accessKeyID, accessKeySecret, bucketName string) {

	once.Do(func() {
		client, err := gooss.New(endpoint, accessKeyID, accessKeySecret)
		if err != nil {
			initErr = err
			return
		}

		bucket, err := client.Bucket(bucketName)
		if err != nil {
			initErr = err
			return
		}

		globalOSS = &AliyunOSS{
			Client: client,
			Bucket: bucket,
		}
	})
	if initErr != nil {
		Logger.Error("初始化 OSS 客户端失败", zap.Any("error", initErr))
		return
	}
	Logger.Info("初始化 OSS 客户端成功")
}

func GetAliOSS() *AliyunOSS {
	if globalOSS == nil {
		Logger.Error("OSS 实例未初始化")
		return nil
	}
	Logger.Info("获取 OSS 实例")
	return globalOSS
}
