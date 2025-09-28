package alioss

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	. "go-sip/logger"
	"go-sip/utils"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	gooss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

// 以字节数组形式上传文件到OSS
func UploadBytes(objectKey string, fileHeader *multipart.FileHeader) (string, string, error) {
	aliyunoss := GetAliOSS()
	if aliyunoss == nil {
		Logger.Error("阿里云OSS客服端获取失败")
		return "", "", fmt.Errorf("阿里云OSS客服端获取失败")
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		Logger.Error("文件打开失败: " + err.Error())
		return "", "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	md5sum, err := utils.MD5FromReader(tee)
	if err != nil {
		Logger.Error("MD5计算失败: " + err.Error())
	}

	// bytes.NewReader(data) 返回 io.Reader
	err = aliyunoss.Bucket.PutObject(objectKey, &buf)
	if err != nil {
		Logger.Error("OSS上传文件失败", zap.Any("objectKey", objectKey))
		return "", "", fmt.Errorf("上传字节流失败: %w", err)
	}

	signedURL, err := aliyunoss.Bucket.SignURL(objectKey, gooss.HTTPGet, 10*365*24*3600) // 10年有效期
	if err != nil {
		Logger.Error("OSS生成签名 URL 失败", zap.Any("objectKey", objectKey))
		return "", "", fmt.Errorf("生成签名 URL 失败: %w", err)
	}
	Logger.Info("OSS上传文件成功", zap.Any("objectKey", objectKey), zap.Any("signedURL", signedURL))
	return signedURL, md5sum, nil
}

// 上传本地文件到OSS
func UploadFile(objectKey, filePath string) (string, error) {
	// 获取阿里云OSS客服端
	aliyunoss := GetAliOSS()
	if aliyunoss == nil {
		Logger.Error("阿里云OSS客服端获取失败")
		return "", fmt.Errorf("阿里云OSS客服端获取失败")
	}
	// 校验文件是否存在
	exists := utils.FileExists(filePath)
	if !exists {
		Logger.Error("文件不存在")
		return "", fmt.Errorf("文件不存在")
	}
	err := aliyunoss.Bucket.PutObjectFromFile(objectKey, filePath)
	if err != nil {
		Logger.Error("上传文件失败", zap.Error(err))
		return "", fmt.Errorf("上传文件失败")
	}
	signedURL, err := aliyunoss.Bucket.SignURL(objectKey, gooss.HTTPGet, 10*365*24*3600)
	if err != nil {
		Logger.Error("生成签名 URL 失败", zap.Error(err))
		return "", fmt.Errorf("生成签名 URL 失败")
	}
	return signedURL, nil
}

// 下载oss文件
func DownloadFile(objectKey, filePath string) (bool, error) {
	aliyunoss := GetAliOSS()
	if aliyunoss == nil {
		Logger.Error("阿里云OSS客服端获取失败")
		return false, fmt.Errorf("阿里云OSS客服端获取失败")
	}
	// 参数不能为空
	if objectKey == "" || filePath == "" {
		Logger.Error("参数不能为空")
		return false, fmt.Errorf("参数不能为空")
	}
	// 判断文件是否存在
	exist, err := aliyunoss.Bucket.IsObjectExist(objectKey)
	if err != nil || !exist {
		return false, fmt.Errorf("文件不存在")
	}
	err = aliyunoss.Bucket.GetObjectToFile(objectKey, filePath)
	if err != nil {
		return false, fmt.Errorf("从OSS下载文件失败: %w", err)
	}
	return true, nil
}

// DeleteObjectFromOSS 删除 OSS 中的对象
func DeleteObject(objectKey string) error {
	aliyunoss := GetAliOSS()
	if aliyunoss == nil {
		Logger.Error("阿里云OSS客户端获取失败")
		return fmt.Errorf("阿里云OSS客户端获取失败")
	}

	// 调用删除
	err := aliyunoss.Bucket.DeleteObject(objectKey)
	if err != nil {
		Logger.Error("OSS 删除对象失败", zap.String("objectKey", objectKey), zap.Error(err))
		return fmt.Errorf("OSS 删除对象失败: %w", err)
	}
	Logger.Info("OSS 删除对象成功", zap.String("objectKey", objectKey))
	return nil
}

// DeleteOldObjects 删除 OSS 中指定目录下创建时间早于指定天数的文件
func DeleteOldObjects(prefix string, days int) error {
	aliyunoss := GetAliOSS()
	if aliyunoss == nil {
		Logger.Error("阿里云OSS客户端获取失败")
		return fmt.Errorf("阿里云OSS客户端获取失败")
	}

	// 计算时间阈值
	expireTime := time.Now().AddDate(0, 0, -days)

	// 分页列举对象
	marker := oss.Marker("")
	for {
		lsRes, err := aliyunoss.Bucket.ListObjects(oss.Prefix(prefix), marker, oss.MaxKeys(1000))
		if err != nil {
			Logger.Error("列举 OSS 文件失败", zap.Error(err))
			return fmt.Errorf("列举 OSS 文件失败: %w", err)
		}

		for _, object := range lsRes.Objects {
			objectKey := object.Key
			if !strings.HasSuffix(object.Key, ".mp4") {
				continue // 跳过非 .mp4 文件
			}
			// 获取对象的元信息
			meta, err := aliyunoss.Bucket.GetObjectDetailedMeta(object.Key)
			if err != nil {
				Logger.Warn("获取对象元信息失败", zap.String("key", object.Key), zap.Error(err))
				continue
			}

			// 从 HTTP Header 中获取 Last-Modified 字段（标准 HTTP 头）
			lastModifiedStr := meta.Get("Last-Modified")
			if lastModifiedStr == "" {
				Logger.Warn("Last-Modified 字段缺失", zap.String("key", objectKey))
				continue
			}

			// 解析时间格式（RFC1123格式）
			lastModified, err := time.Parse(time.RFC1123, lastModifiedStr)
			if err != nil {
				Logger.Warn("Last-Modified 时间解析失败", zap.String("key", objectKey), zap.String("value", lastModifiedStr), zap.Error(err))
				continue
			}
			if lastModified.Before(expireTime) {
				err = aliyunoss.Bucket.DeleteObject(object.Key)
				if err != nil {
					Logger.Warn("删除过期文件失败", zap.String("key", object.Key), zap.Error(err))
					continue
				}
				Logger.Info("已删除过期文件", zap.String("key", object.Key), zap.Time("lastModified", lastModified))
			}
		}

		if !lsRes.IsTruncated {
			break
		}
		marker = oss.Marker(lsRes.NextMarker)
	}
	return nil
}
