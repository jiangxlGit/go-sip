package m

type DatabaseConfig struct {
	Redis RedisConfig `yaml:"redis"` // Redis 配置
	Mysql MysqlConfig `yaml:"mysql"` // MySQL 配置
}

type RedisConfig struct {
	Dialect  string `json:"dialect" yaml:"dialect" mapstructure:"dialect"`
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
}

type MysqlConfig struct {
	Host            string `yaml:"host"`              // MySQL 主机
	User            string `yaml:"user"`              // MySQL 用户名
	Password        string `yaml:"password"`          // MySQL 密码
	DBName          string `yaml:"dbname"`            // MySQL 数据库名称
	MaxOpenConns    int    `yaml:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int    `yaml:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 连接最大生命周期 单位秒
}

type AliOSSConfig struct { 
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	BucketName      string `yaml:"bucketName"`
}

