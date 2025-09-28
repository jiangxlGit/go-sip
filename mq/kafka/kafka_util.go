package kafka

import (
	"encoding/json"
	"fmt"
	. "go-sip/logger"
	"go-sip/m"

	ka "github.com/IBM/sarama"
	"go.uber.org/zap"
)

// SendKafkaMessage 发送任意结构的消息体到默认 Kafka topic
func SendKafkaMessage(deviceId string, message interface{}) error {
	if KafkaProducer == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := &ka.ProducerMessage{
		Topic: m.SMConfig.KafkaCfg.Topic,
		Key:   ka.StringEncoder(deviceId),
		Value: ka.ByteEncoder(data),
	}

	partition, offset, err := KafkaProducer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send kafka message: %w", err)
	}

	Logger.Info("sent kafka message", zap.Any("topic", m.SMConfig.KafkaCfg.Topic), zap.Any("partition", partition), zap.Any("offset", offset))
	return nil
}

// SendKafkaMessageByTopic 发送kafka消息
func SendKafkaMessageByTopic(topic string, deviceId string, message interface{}) error {
	if KafkaProducer == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := &ka.ProducerMessage{
		Topic: topic, // 使用外部传入的 topic
		Key:   ka.StringEncoder(deviceId),
		Value: ka.ByteEncoder(data),
	}

	partition, offset, err := KafkaProducer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send kafka message to topic %s: %w", topic, err)
	}

	Logger.Info("sent kafka message",
		zap.String("topic", topic),
		zap.Any("partition", partition),
		zap.Any("offset", offset),
	)

	return nil
}
