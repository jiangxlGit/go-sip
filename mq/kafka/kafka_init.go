package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	. "go-sip/logger"
	"go-sip/m"

	"github.com/IBM/sarama"
	ka "github.com/IBM/sarama"
	"github.com/xdg-go/scram"
	"go.uber.org/zap"
)

const (
	HTY_IOT_DOOR_STATE_TOPIC    = "HTY_IOT_DOOR_STATE"
	HTY_IOT_DEVICE_ONLINE_TOPIC = "HTY_IOT_DEVICE_ONLINE"
)

var (
	SHA256        scram.HashGeneratorFcn = sha256.New
	SHA512        scram.HashGeneratorFcn = sha512.New
	KafkaProducer ka.SyncProducer
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}
func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}
func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

func InitKafkaProducer() error {
	cfg := m.SMConfig.KafkaCfg
	config := ka.NewConfig()

	config.Producer.Return.Successes = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = cfg.User
	config.Net.SASL.Password = cfg.Password
	config.Net.SASL.Mechanism = ka.SASLTypeSCRAMSHA512
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
	}

	p, err := ka.NewSyncProducer(cfg.BrokerList, config)
	if err != nil {
		Logger.Error("init kafka producer failed", zap.Error(err))
		return fmt.Errorf("init kafka producer failed: %w", err)
	}
	Logger.Info("init kafka producer success")
	KafkaProducer = p
	return nil
}

func CloseKafkaProducer() {
	if KafkaProducer != nil {
		_ = KafkaProducer.Close()
	}
}
