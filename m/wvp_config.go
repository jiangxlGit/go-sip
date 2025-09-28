package m

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config Config
type WVP_Config struct {
	API       string         `yaml:"api"`       // wvp服务暴露ip
	Port      string         `yaml:"port"`      // 监听端口
	Username  string         `yaml:"username"`  // 登录用户名
	Password  string         `yaml:"password"`  // 登录密码
	DataBase  DatabaseConfig `yaml:"database"`  // 数据库配置
	AliYunOss AliOSSConfig   `yaml:"aliyunoss"` // 阿里云OSS
	LogLevel  string         `json:"logLevel" yaml:"logLevel" mapstructure:"logLevel"`
}

var WVPConfig *WVP_Config

func LoadWvpConfig() {
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
	WVPConfig = &WVP_Config{}
	err = viper.Unmarshal(&WVPConfig)
	if err != nil {
		panic(err)
	}
	// 自动设置 API 为 0.0.0.0:<port>（如果未设置）
	api := strings.TrimSpace(WVPConfig.API)
	if api == "" || api == "0.0.0.0" || api == "0.0.0.0:" {
		WVPConfig.API = fmt.Sprintf("0.0.0.0:%s", WVPConfig.Port)
	}
}
