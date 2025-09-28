package mqtt

import (
	. "go-sip/logger"
	"go-sip/m"

	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

var MqttClient mqtt.Client

func InitMqttClient() {
	Logger.Info("初始化MQTT客户端", zap.Any("brokers", m.SMConfig.MqttConfig.Brokers))

	opts := mqtt.NewClientOptions().
		SetClientID(fmt.Sprintf("gosip-server-mqtt-%s", m.SMConfig.SipID)).
		SetCleanSession(m.SMConfig.MqttConfig.CleanSession).                 // 持久会话，避免断线丢订阅
		SetAutoReconnect(m.SMConfig.MqttConfig.AutoReconnect).               // 自动重连
		SetConnectRetry(m.SMConfig.MqttConfig.ConnectRetry).                 // 连接失败自动重试
		SetConnectRetryInterval(m.SMConfig.MqttConfig.ConnectRetryInterval). // 尝试重连间隔
		SetKeepAlive(m.SMConfig.MqttConfig.KeepAlive).                       // 心跳间隔
		SetPingTimeout(m.SMConfig.MqttConfig.PingTimeout).                   // PING 响应超时
		SetUsername(m.SMConfig.MqttConfig.Username).
		SetPassword(m.SMConfig.MqttConfig.Password)

	// 添加多个 Broker 地址
	for _, broker := range m.SMConfig.MqttConfig.Brokers {
		opts.AddBroker(broker)
	}

	// ====== 回调函数 ======
	opts.OnConnect = func(c mqtt.Client) {
		Logger.Info("Connected to MQTT broker, 开始订阅")
		// 订阅业务单独调用方法
		SubscribeTopics(c)
	}

	// ====== 连接丢失 ======
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		Logger.Error("连接丢失，等待自动重连", zap.Error(err))
	}

	// ====== 建立连接 ======
	MqttClient = mqtt.NewClient(opts)

	// 同步连接，打印首次连接失败
	token := MqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		Logger.Error("MQTT 首次连接失败", zap.Error(token.Error()))
	} else {
		Logger.Info("MQTT Connect() 已调用，等待 OnConnect 回调订阅")
	}

	// 启动后台监控
	go monitorConnection(MqttClient, 10*time.Second)
}

// monitorConnection 定期检查连接状态
func monitorConnection(c mqtt.Client, interval time.Duration) {
	for {
		time.Sleep(interval)
		if !c.IsConnected() {
			Logger.Warn("MQTT 未连接，等待自动重连中")
		}
	}
}
