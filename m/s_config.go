package m

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type MqttConfig struct {
	Brokers              []string         `yaml:"brokers"`
	Username             string           `yaml:"username"`
	Password             string           `yaml:"password"`
	AutoReconnect        bool             `yaml:"autoReconnect"`
	ConnectRetry         bool             `yaml:"connectRetry"`
	KeepAlive            time.Duration    `yaml:"keepAlive"`
	PingTimeout          time.Duration    `yaml:"pingTimeout"`
	ConnectRetryInterval time.Duration    `yaml:"connectRetryInterval"`
	CleanSession         bool             `yaml:"cleanSession"`
	Topics               map[string]uint8 `yaml:"topics"` // topic → QoS
}

type KafkaConfig struct {
	BrokerList []string `json:"brokers" yaml:"brokers" mapstructure:"brokers"`
	User       string   `json:"user" yaml:"user" mapstructure:"user"`
	Password   string   `json:"passwd" yaml:"passwd" mapstructure:"passwd"`
	Topic      string   `json:"topic" yaml:"topic" mapstructure:"topic"`
}

// Config Config
type S_Config struct {
	SipID      string      `json:"sip_id" yaml:"sip_id" mapstructure:"sip_id"`
	SipInnerIp string      `json:"sip_inner_ip" yaml:"sip_inner_ip" mapstructure:"sip_inner_ip"`
	SipOutIp   string      `json:"sip_out_ip" yaml:"sip_out_ip" mapstructure:"sip_out_ip"`
	API        string      `json:"api" yaml:"api" mapstructure:"api"`
	SipPort    string      `json:"sip_port" yaml:"sip_port" mapstructure:"sip_port"`
	TcpIp      string      `json:"tcp_ip" yaml:"tcp_ip" mapstructure:"tcp_ip"`
	TcpPort    string      `json:"tcp_port" yaml:"tcp_port" mapstructure:"tcp_port"`
	UDP        string      `json:"udp" yaml:"udp" mapstructure:"udp"`
	Secret     string      `json:"secret" yaml:"secret" mapstructure:"secret"`
	Sign       string      `json:"sign" yaml:"sign" mapstructure:"sign"`
	DataBase   RedisConfig `json:"database" yaml:"database" mapstructure:"database"`
	KafkaCfg   KafkaConfig `json:"kafka" yaml:"kafka" mapstructure:"kafka"`
	MqttConfig MqttConfig  `json:"mqtt" yaml:"mqtt" mapstructure:"mqtt"`
	LogLevel   string      `json:"logLevel" yaml:"logLevel" mapstructure:"logLevel"`
}

var SMConfig *S_Config

func LoadServerConfig() {
	viper.SetConfigType("yml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.SetDefault("mod", "release")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	SMConfig = &S_Config{}
	err = viper.Unmarshal(&SMConfig)
	if err != nil {
		panic(err)
	}
	// 自动设置 API 为 0.0.0.0:<sip_port>（如果未设置）
	api := strings.TrimSpace(SMConfig.API)
	if api == "" || api == "0.0.0.0" || api == "0.0.0.0:" {
		SMConfig.API = fmt.Sprintf("0.0.0.0:%s", SMConfig.SipPort)
	}
}
