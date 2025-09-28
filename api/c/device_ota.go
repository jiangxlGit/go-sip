package api

import (
	. "go-sip/logger"
	"go-sip/utils"
	"os/exec"
	"path/filepath"
	"time"

	"fmt"
	"os"

	"github.com/moby/sys/reexec"
	"go.uber.org/zap"
)

func init() {
	fmt.Fprintln(os.Stderr, "init runOtaTask")
	reexec.Register("runOtaTask", runOtaTask)
	if reexec.Init() {
		os.Exit(0)
	}
}

// 设备升级
func DeviceOTA(DeviceID, firmwareId, firmwareDownloadUrl, firmwareMd5, firmwareVersion string) error {
	// 项目根目录
	rootPath, err := os.Getwd()
	if err != nil {
		Logger.Error("获取当前目录失败", zap.Any("err", err))
		return fmt.Errorf("获取当前目录失败: %v", err)
	}
	// 指定下载目录
	otaFirmwareDir, err := utils.FileDirHandler(fmt.Sprintf("%s/%s", rootPath, "ota_firmware"), firmwareVersion)
	if err != nil || otaFirmwareDir == "" {
		Logger.Error("创建目录失败", zap.Any("err", err))
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// files, err := os.ReadDir(otaFirmwareDir)
	// if err != nil {
	// 	Logger.Error("读取目录失败", zap.Any("err", err))
	// 	return fmt.Errorf("读取目录失败: %v", err)
	// }

	// otaFirmwareMd5Map := make(map[string]string)
	// for _, file := range files {
	// 	if file.IsDir() {
	// 		continue // 跳过目录
	// 	}
	// 	fullPath := filepath.Join(otaFirmwareDir, file.Name())
	// 	md5Str, err := utils.ComputeFileMD5(fullPath)
	// 	if err != nil {
	// 		Logger.Error("计算文件MD5失败", zap.Any("fileName", file.Name()), zap.Error(err))
	// 		return fmt.Errorf("计算文件MD5失败: %v", err)
	// 	}
	// 	otaFirmwareMd5Map[file.Name()] = md5Str
	// }

	otaFirmwarePath := filepath.Join(otaFirmwareDir, "go-sip-client")
	// if md5, ok := otaFirmwareMd5Map[firmwareId]; ok {
	// 	if md5 != firmwareMd5 {
	// 		Logger.Info("固件MD5不匹配,重新下载", zap.Any("firmwareVersion", firmwareVersion), zap.Any("newMd5", md5), zap.Any("oldMd5", firmwareMd5))
	// 		_, err := utils.DownloadFileTo(otaFirmwarePath, firmwareDownloadUrl)
	// 		if err != nil {
	// 			Logger.Error("下载固件失败", zap.Any("firmwareVersion", firmwareVersion), zap.Any("firmwareDownloadUrl", firmwareDownloadUrl), zap.Error(err))
	// 			return fmt.Errorf("下载固件失败: %v", err)
	// 		}
	// 		newMd5, err := utils.ComputeFileMD5(otaFirmwarePath)
	// 		if newMd5 != firmwareMd5 {
	// 			Logger.Error("固件MD5不匹配", zap.Any("firmwareVersion", firmwareVersion), zap.Any("newMd5", newMd5), zap.Any("oldMd5", firmwareMd5))
	// 			return fmt.Errorf("固件MD5不匹配")
	// 		}
	// 	}
	// } else {
	Logger.Info("固件开始下载", zap.Any("firmwareVersion", firmwareVersion), zap.Any("firmwareDownloadUrl", firmwareDownloadUrl))
	_, err = utils.DownloadFileTo(otaFirmwarePath, firmwareDownloadUrl)
	if err != nil {
		Logger.Error("下载固件失败", zap.Any("firmwareVersion", firmwareVersion), zap.Any("firmwareDownloadUrl", firmwareDownloadUrl), zap.Error(err))
		return fmt.Errorf("下载固件失败: %v", err)
	}
	newMd5, err := utils.ComputeFileMD5(otaFirmwarePath)
	if newMd5 != firmwareMd5 {
		Logger.Error("固件MD5不匹配", zap.Any("firmwareVersion", firmwareVersion), zap.Any("newMd5", newMd5), zap.Any("oldMd5", firmwareMd5))
		return fmt.Errorf("固件MD5不匹配")
	}
	// }
	Logger.Info("固件下载成功", zap.Any("otaFirmwareDir", otaFirmwareDir))
	// 备份老固件
	lastOtaFirmwareDir, err := utils.FileDirHandler(rootPath, "last_ota_firmware")
	if err != nil || lastOtaFirmwareDir == "" {
		Logger.Error("创建目录失败", zap.Any("err", err))
		return fmt.Errorf("创建目录失败: %v", err)
	} else {
		Logger.Info("备份固件", zap.Any("lastOtaFirmwareDir", lastOtaFirmwareDir))
		err = utils.CopyFile(fmt.Sprintf("%s/%s", rootPath, "go-sip-client"), lastOtaFirmwareDir)
		if err != nil {
			Logger.Error("备份固件失败", zap.Any("err", err))
			return fmt.Errorf("备份固件失败: %v", err)
		}
	}
	Logger.Info("备份固件成功", zap.Any("lastOtaFirmwareDir", lastOtaFirmwareDir))

	Logger.Info("开始升级固件", zap.Any("otaFirmwareDir", otaFirmwareDir))

	// fork 子进程
	cmd := reexec.Command("runOtaTask", rootPath, otaFirmwareDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		Logger.Error("fork子进程失败", zap.Any("err", err))
		return fmt.Errorf("fork子进程失败: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		Logger.Error("固件升级失败", zap.Any("err", err))
		return fmt.Errorf("固件升级失败: %v", err)
	}

	if err != nil {
		Logger.Error("固件升级失败", zap.Any("err", err))
		return fmt.Errorf("固件升级失败: %v", err)
	}
	Logger.Info("固件升级成功", zap.Any("otaFirmwareDir", otaFirmwareDir))

	return nil
}

func runOtaTask() {
	// os.Args[0] 是程序本身路径
	// os.Args[1] 就是 "runOtaTask"
	// os.Args[2:] 才是真正的传参
	if len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "正式开始固件升级 runOtaTask")
		os.Exit(2) // 明确告诉父进程是失败
	}
	rootPath := os.Args[2]
	otaFirmwareDir := os.Args[3]
	// 停止运行的go-sip-client进程
	cmd := exec.Command("supervisorctl", "stop", "sip")
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "停止运行中的旧固件失败")
		os.Exit(2) // 明确告诉父进程是失败
	}
	time.Sleep(5 * time.Second)

	// 回滚
	err = rollbackOtaTask(rootPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "回滚固件失败")
		os.Exit(2) // 明确告诉父进程是失败
	}

	// 新固件覆盖旧固件
	err = utils.CopyFile(fmt.Sprintf("%s/%s", otaFirmwareDir, "go-sip-client"), fmt.Sprintf("%s/%s", rootPath, "go-sip-client"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "覆盖旧固件失败")
		// 回滚
		err := rollbackOtaTask(rootPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "回滚固件失败")
			os.Exit(2) // 明确告诉父进程是失败
		}
	}
	// 执行supervisorctl start sip命令
	cmd = exec.Command("supervisorctl", "start", "sip")
	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "启动sip服务失败")
		os.Exit(2) // 明确告诉父进程是失败
	}
	fmt.Println(os.Stderr, "升级成功")
	os.Exit(0)
}

func rollbackOtaTask(rootPath string) error {
	const maxRetries = 3
	src := fmt.Sprintf("%s/%s", rootPath, "last_ota_firmware/go-sip-client")
	dst := fmt.Sprintf("%s/%s", rootPath, "go-sip-client")

	var err error
	for i := 1; i <= maxRetries; i++ {
		err = utils.CopyFile(src, dst)
		if err == nil {
			// 执行supervisorctl start sip命令
			cmd := exec.Command("supervisorctl", "start", "sip")
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("执行supervisorctl start sip命令失败: %v", err)
			}
			return nil
		}
		time.Sleep(1 * time.Second) // 间隔重试，避免瞬间三连失败
	}
	return fmt.Errorf("回滚固件最终失败，超过最大重试次数: %v", err)
}
