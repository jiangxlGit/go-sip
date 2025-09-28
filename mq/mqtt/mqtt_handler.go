package mqtt

import (
	"encoding/json"
	"fmt"
	"go-sip/grpc_api"
	grpc_server "go-sip/grpc_api/s"
	. "go-sip/logger"
	"go-sip/m"
	pb "go-sip/signaling"

	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// 非保留发布消息方法
func SimplePublishMessage(topic string, msg interface{}, qos byte) error {
	// 将对象转为 JSON
	var payload []byte
	switch v := msg.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		var err error
		payload, err = json.Marshal(v)
		if err != nil {
			Logger.Error("消息序列化失败",
				zap.Any("msg", msg),
				zap.Error(err))
			return err
		}
	}
	return PublishMessage(topic, payload, qos, false)
}

// 通用发布消息方法
func PublishMessage(topic string, payload []byte, qos byte, retained bool) error {
	if !MqttClient.IsConnected() {
		Logger.Error("MQTT 客户端未连接，消息无法发布",
			zap.String("topic", topic),
			zap.ByteString("payload", payload))
		return fmt.Errorf("mqtt client not connected")
	}

	token := MqttClient.Publish(topic, qos, retained, payload)
	token.Wait() // 等待发送完成

	if token.Error() != nil {
		Logger.Error("MQTT 消息发布失败",
			zap.String("topic", topic),
			zap.ByteString("payload", payload),
			zap.Error(token.Error()))
		return token.Error()
	}

	Logger.Info("MQTT 消息发布成功",
		zap.String("topic", topic),
		zap.ByteString("payload", payload),
		zap.Int("qos", int(qos)),
		zap.Bool("retained", retained))
	return nil
}

// 订阅业务方法
func SubscribeTopics(client mqtt.Client) {
	// 定义业务 topic 和 QoS
	topics := m.SMConfig.MqttConfig.Topics

	// 订阅
	token := client.SubscribeMultiple(topics, func(c mqtt.Client, msg mqtt.Message) {
		// 所有订阅消息统一回调
		go handleMessage(msg)
	})

	if token.Wait() && token.Error() != nil {
		Logger.Error("批量订阅失败", zap.Error(token.Error()))
	} else {
		Logger.Info("批量订阅成功")
	}
}

// 处理不同 topic 的业务逻辑
func handleMessage(msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()
	Logger.Info("处理MQTT业务消息", zap.String("topic", topic), zap.String("payload", string(payload)))

	if strings.HasSuffix(topic, "/firmware/pull/reply") {
		handleFirmwareReply(payload)
	} else if strings.HasPrefix(topic, "xxxx/xxxx") {

	} else {
		Logger.Warn("未知的MQTT业务消息", zap.String("topic", topic), zap.String("payload", string(payload)))
	}
}

func handleFirmwareReply(payload []byte) {
	Logger.Info("处理固件下发回复", zap.String("payload", string(payload)))
	var reply FirmwareReplyMqttMsg
	err := json.Unmarshal(payload, &reply)
	if err != nil {
		Logger.Error("解析MQTT消息失败", zap.Error(err))
		return
	}
	sip_req := &grpc_api.OTA_Event_Req{
		DeviceID:            reply.DeviceID,
		FirmwareID:          reply.FirmwareID,
		FirmwareDownloadURL: reply.URL,
		FirmwareMD5:         reply.Sign,
		FirmwareVersion:     reply.Version,
	}
	d, err := json.Marshal(sip_req)
	if err != nil {
		Logger.Error("json序列化失败")
		return
	}

	sip_server := grpc_server.GetSipServer()
	_, err = sip_server.ExecuteCommand(reply.DeviceID, &pb.ServerCommand{
		MsgID:   m.MsgID_DeviceOTA,
		Method:  m.DeviceOTA,
		Payload: d,
	})
	if err != nil {
		Logger.Error("设备升级失败", zap.Any("device_id", reply.DeviceID), zap.Error(err))
		return
	}
	Logger.Info("设备升级成功", zap.Any("device_id", reply.DeviceID))

}
