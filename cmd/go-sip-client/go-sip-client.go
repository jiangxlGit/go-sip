package main

import (
	"bytes"
	"fmt"
	capi "go-sip/api/c"
	. "go-sip/common"
	"go-sip/db/alioss"
	grpc_client "go-sip/grpc_api/c"
	"go-sip/logger"
	. "go-sip/logger"
	"go-sip/m"
	sipapi "go-sip/sip"
	"go-sip/zlm_api"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ensureEchoCancelModuleLoaded() {
	// 1. 查询已加载模块
	cmd := exec.Command("pactl", "list", "modules", "short")
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// 2. 检查 module-echo-cancel 是否存在
	if strings.Contains(string(output), "module-echo-cancel") {
		// 2.1 获取模块ID
		lines := strings.Split(string(output), "\n")
		var moduleID string
		for _, line := range lines {
			if strings.Contains(line, "module-echo-cancel") {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					moduleID = fields[0]
					break
				}
			}
		}

		// 2.2 卸载现有模块
		if moduleID != "" {
			err := exec.Command("pactl", "unload-module", moduleID).Run()
			if err != nil {
				panic(fmt.Errorf("卸载 module-echo-cancel 失败: %v", err))
			}
			Logger.Info("已卸载 module-echo-cancel ", zap.Any("moduleID", moduleID))
		}
	}

	// 3. 校验配置
	if m.CMConfig.Audio.InputDevice == "" || m.CMConfig.Audio.OutputDevice == "" {
		Logger.Error("音频输入或输出设备未配置，无法加载 module-echo-cancel")
		panic("音频输入或输出设备未配置")
	}

	sink := fmt.Sprintf("sink_master=%s", m.CMConfig.Audio.OutputDevice)
	source := fmt.Sprintf("source_master=%s", m.CMConfig.Audio.InputDevice)

	Logger.Info("加载 module-echo-cancel", zap.String("sink", sink), zap.String("source", source))

	// 4. 加载 module-echo-cancel
	loadCmd := exec.Command("pactl", "load-module",
		"module-echo-cancel",
		"sink_name=echo_cancel_sink",
		"source_name=echo_cancel_source",
		sink,
		source,
		"rate=8000",
		"channels=1",
		"aec_method=webrtc",
	)
	var stderr bytes.Buffer
	loadCmd.Stderr = &stderr

	if err := loadCmd.Run(); err != nil {
		panic(fmt.Errorf("加载 module-echo-cancel 失败: %v\n%s", err, stderr.String()))
	}

	Logger.Info("module-echo-cancel 已加载")

	// 5. 设置默认输入设备（source）为 echo_cancel_source
	if err := exec.Command("pactl", "set-default-source", "echo_cancel_source").Run(); err != nil {
		Logger.Warn("设置默认输入设备失败", zap.Error(err))
	} else {
		Logger.Info("默认输入设备已设置为 echo_cancel_source")
	}

	// 6. 设置默认输出设备（sink）为 echo_cancel_sink
	if err := exec.Command("pactl", "set-default-sink", "echo_cancel_sink").Run(); err != nil {
		Logger.Warn("设置默认输出设备失败", zap.Error(err))
	} else {
		Logger.Info("默认输出设备已设置为 echo_cancel_sink")
	}
}

func GetPlatform() string {

	if model, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		if strings.Contains(string(model), "rk3568") {
			return "rk3568"
		} else if strings.Contains(string(model), "rk3588") {
			return "rk3588"
		} else if strings.Contains(string(model), "rk3576") {
			return "rk3576"
		}
	}

	return m.CMConfig.DeviceType
}

func main() {

	m.LoadClientConfig()
	logger.InitLogger(m.CMConfig.LogLevel)

	ensureEchoCancelModuleLoaded()
	device_id, err := os.Hostname()
	if err != nil || device_id == "" {
		panic(err)
	}

	r := gin.Default()
	r.POST(ZLMWebHookClientURL, capi.ZLMWebHook)

	client := grpc_client.NewSipClient(device_id)
	sipapi.Start()

	os.Setenv("PULSE_PROP", "filter.want=echo-cancel")

	// 关闭本地zlm所有流
	zlm_api.ZlmCloseAllStreams(sipapi.Local_ZLM_Host, m.CMConfig.ZlmSecret)

	// 初始化alioss
	alioss.SipClientInitAliOSS()

	// 下载设备关联的模型列表
	capi.DownloadDeviceRelationAIModel(device_id, GetPlatform())

	// 非国标摄像头推流
	capi.IpcPushStreamHandler(device_id)

	// 启动AI事件触发监听
	go capi.StartAiModelTriggerHandler()

	// 启动 API 服务，放到 goroutine 中
	go func() {
		for {
			err := r.Run(m.CMConfig.API)
			if err != nil {
				Logger.Error("api服务启动失败", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
			// 如果 r.Run() 正常退出了，重新启动
			Logger.Warn("api服务退出，尝试重新启动...")
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		tcp_addr := capi.GetSipServerTcpAddr(m.CMConfig.Gateway, device_id)
		if tcp_addr == "" {
			// 默认获取配置中的tcp地址
			tcp_addr = m.CMConfig.TCP
		}
		if err := client.Connect(tcp_addr); err != nil {
			Logger.Error("连接失败", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		Logger.Info("连接成功")
		client.Run()
		Logger.Info("连接断开，尝试重新连接...")
	}
}
