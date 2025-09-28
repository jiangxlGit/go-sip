package m

import (
	"strings"
	"time"

	db "go-sip/db/sqlite"
	"go-sip/utils"

	"fmt"

	"github.com/spf13/viper"
)

// Config Config
type C_Config struct {
	API           string         `json:"api" yaml:"api" mapstructure:"api"`
	SipClientPort string         `json:"sip_client_port" yaml:"sip_client_port" mapstructure:"sip_client_port"`
	UDP           string         `json:"udp" yaml:"udp" mapstructure:"udp"`
	TCP           string         `json:"tcp" yaml:"tcp" mapstructure:"tcp"`
	DeviceType    string         `json:"device_type" yaml:"device_type" mapstructure:"device_type"` // rk3568  rk3576  rk3588
	Gateway       string         `json:"gateway" yaml:"gateway" mapstructure:"gateway"`
	ZlmSecret     string         `json:"zlm_secret" yaml:"zlm_secret" mapstructure:"zlm_secret"`
	ZlmInnerIp    string         `json:"zlm_inner_ip" yaml:"zlm_inner_ip" mapstructure:"zlm_inner_ip"`
	LogLevel      string         `json:"logLevel" yaml:"logLevel" mapstructure:"logLevel"`
	Stream        *Stream        `json:"stream" yaml:"stream" mapstructure:"stream"`
	GB28181       *SysInfo       `json:"gb28181" yaml:"gb28181" mapstructure:"gb28181"`
	Audio         *AudioConfig   `json:"audio" yaml:"audio" mapstructure:"audio"`
	OpenApi       *OpenApiConfig `json:"openapi" yaml:"openapi" mapstructure:"openapi"`
	AliYunOss     AliOSSConfig   `yaml:"aliyunoss"` // 阿里云OSS
}

// Stream Stream
type Stream struct {
	HLS  bool `json:"hls" yaml:"hls" mapstructure:"hls"`
	RTMP bool `json:"rtmp" yaml:"rtmp" mapstructure:"rtmp"`
}

type SysInfo struct {
	// Region 当前域
	Region string `json:"region"   yaml:"region" mapstructure:"region"`
	// LID 当前服务id
	LID string `json:"lid" bson:"lid" yaml:"lid" mapstructure:"lid"`
	// 密码
	Passwd string `json:"passwd" bson:"passwd" yaml:"passwd" mapstructure:"passwd"`
}

type AudioConfig struct {
	SampleRate   float64 `json:"sample_rate" yaml:"sample_rate" mapstructure:"sample_rate"`
	Channels     int     `json:"channels" yaml:"channels" mapstructure:"channels"`
	InputDevice  string  `json:"input_device" yaml:"input_device" mapstructure:"input_device"`
	OutputDevice string  `json:"output_device" yaml:"output_device" mapstructure:"output_device"`
}

type OpenApiConfig struct {
	ClientId  string `json:"client_id" yaml:"client_id" mapstructure:"client_id"`
	SecretKey string `json:"secret_key" yaml:"secret_key" mapstructure:"secret_key"`
}

func DefaultInfo() *SysInfo {
	return CMConfig.GB28181
}

var CMConfig *C_Config

func LoadClientConfig() {
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
	CMConfig = &C_Config{}
	err = viper.Unmarshal(&CMConfig)
	if err != nil {
		panic(err)
	}
	db.DBClient, err = db.Open()
	if err != nil {
		panic(err)
	}
	db.DBClient.SetNowFuncOverride(func() interface{} {
		return time.Now().Unix()
	})
	db.DBClient.LogMode(false)
	go db.KeepLive(db.DBClient, time.Minute)

	// 自动设置 API 为 0.0.0.0:<sip_port>（如果未设置）
	api := strings.TrimSpace(CMConfig.API)
	if api == "" || api == "0.0.0.0" || api == "0.0.0.0:" {
		CMConfig.API = fmt.Sprintf("0.0.0.0:%s", CMConfig.SipClientPort)
	}
	localIp := utils.GetPreferredIP()
	fmt.Println("====本机ip:", localIp)
	if localIp != "" {
		CMConfig.ZlmInnerIp = localIp
	}

}
