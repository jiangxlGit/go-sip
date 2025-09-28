package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// 判断目录是否存在
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err // 其他错误，如权限问题
}

func FileDirHandler(rootPath, subPath string) (string, error) {

	// 拼接目标子目录
	saveDir := filepath.Join(rootPath, subPath)

	// 判断是否存在
	exists, err := DirExists(saveDir)
	if err != nil {
		fmt.Println("检查目录失败:", err)
		return "", err
	}

	if !exists {
		// 不存在则创建
		err = os.MkdirAll(saveDir, os.ModePerm)
		if err != nil {
			fmt.Println("创建目录失败:", err)
			return "", err
		}
	}
	return saveDir, nil
}

// downloadFileTo 下载文件到指定路径，并返回文件的 MD5 值
func DownloadFileTo(saveDir string, firmwareDownloadUrl string) (string, error) {
	// 解码 URL
	decodedURL, err := DecodeURLFromJSON(firmwareDownloadUrl)
	if err != nil {
		return "", fmt.Errorf("URL 解码失败: %w", err)
	}

	// 发起请求
	resp, err := http.Get(decodedURL)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件并保存内容
	outFile, err := os.Create(saveDir)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer outFile.Close()

	// 计算 MD5
	hash := md5.New()
	tee := io.TeeReader(resp.Body, hash)

	if _, err = io.Copy(outFile, tee); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回 hex 编码的 MD5 值
	md5sum := hex.EncodeToString(hash.Sum(nil))
	return md5sum, nil
}

// 计算文件的MD5
func ComputeFileMD5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
