package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	. "go-sip/db/alioss"
	. "go-sip/logger"
	"go-sip/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 上传文件到阿里oss
// @Router  /wvp/file/alioss/upload [post]
func FileUploadAliOSSHandler(c *gin.Context) {
	// 从表单中获取文件，字段名前端需与这里一致（比如 file）
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件获取失败: " + err.Error()})
		return
	}
	mianCategory := c.Query("mianCategory")
	subCategory := c.Query("subCategory")
	if mianCategory == "" || subCategory == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不能为空"})
		return
	}

	fileName := fileHeader.Filename
	// 文件id，这里可根据需求生成唯一名（比如加 UUID）
	fileID := generateAiModelID()
	// 获取fileHeader.Filename的后缀
	fileExt := filepath.Ext(fileName)
	fileNamePre := strings.TrimSuffix(fileName, fileExt)
	objectKey := "ai_model/" + mianCategory + "/" + subCategory + "/" + fileNamePre + "_" + fileID + fileExt

	// 上传到 OSS
	url, md5, err := UploadBytes(objectKey, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OSS 上传失败: " + err.Error()})
		return
	}
	Logger.Info("文件上传成功", zap.Any("fileName", fileName), zap.Any("url", url), zap.Any("md5", md5))
	var aiModelFileInfo model.AiModelFileInfo
	aiModelFileInfo.FileId = fileID
	aiModelFileInfo.FileName = fileHeader.Filename
	aiModelFileInfo.FileKey = objectKey
	aiModelFileInfo.FileDownloadUrl = url
	aiModelFileInfo.FileMd5 = md5

	model.JsonResponseSucc(c, aiModelFileInfo)
}

// 生成带前缀、日期、随机数的 ID
func generateAiModelID() string {
	prefix := "AI"
	date := time.Now().Format("20060102")

	// 创建私有随机数生成器，避免全局污染
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := r.Intn(900000) + 100000 // 保证六位

	return fmt.Sprintf("%s%s%d", prefix, date, randomNumber)
}

// 上传本地文件到oss
func UploadAliOSSFile(c *gin.Context) {
	objectKey := c.Query("objectKey")
	if objectKey == "" {
		model.JsonResponseSysERR(c, "objectKey不能为空")
		return
	}
	filePath := c.Query("filePath")
	if filePath == "" {
		model.JsonResponseSysERR(c, "filePath不能为空")
		return
	}
	fileUrl, err := UploadFile(objectKey, filePath)
	if err != nil {
		model.JsonResponseSysERR(c, err.Error())
		return
	}
	model.JsonResponseSucc(c, fileUrl)
}

// @Summary 删除阿里OSS文件
// @Router  /wvp/file/alioss/delete [delete]
func FileDeleteAliOSSHandler(c *gin.Context) {
	objectKey := c.Query("objectKey")
	if objectKey == "" {
		model.JsonResponseSysERR(c, "objectKey不能为空")
		return
	}

	err := DeleteObject(objectKey)
	if err != nil {
		Logger.Error("OSS 文件删除失败", zap.String("objectKey", objectKey), zap.Error(err))
		model.JsonResponseSysERR(c, "OSS 文件删除失败")
		return
	}

	Logger.Info("OSS 文件删除成功", zap.String("objectKey", objectKey))
	model.JsonResponseSucc(c, "删除成功")
}
