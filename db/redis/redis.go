package redis

import (
	"context"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	. "go-sip/logger"
	"go-sip/m"

	"time"
)

const (
	REDIS_POOL_SIZE         = 100
	REDIS_MIN_IDLE_CONNS    = 10
	REDIS_POOL_TIMEOUT_SEC  = 5
	REDIS_DIAL_TIMEOUT_SEC  = 5
	REDIS_READ_TIMEOUT_SEC  = 3
	REDIS_WRITE_TIMEOUT_SEC = 3
)

var (
	redisClients = make(map[string]*redis.Client)
	clientsMu    sync.RWMutex
	RDB          *redis.Client
)

func InitRedis(name, addr, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,

		PoolSize:     REDIS_POOL_SIZE,
		MinIdleConns: REDIS_MIN_IDLE_CONNS,
		PoolTimeout:  time.Duration(REDIS_POOL_TIMEOUT_SEC) * time.Second,
		DialTimeout:  time.Duration(REDIS_DIAL_TIMEOUT_SEC) * time.Second,
		ReadTimeout:  time.Duration(REDIS_READ_TIMEOUT_SEC) * time.Second,
		WriteTimeout: time.Duration(REDIS_WRITE_TIMEOUT_SEC) * time.Second,
	})

	// 测试连接
	ctx := context.Background()
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	clientsMu.Lock()
	redisClients[name] = RDB
	clientsMu.Unlock()
}

func InitGatewayRedisMulti(dbs ...int) {
	for _, db := range dbs {
		InitRedis("gateway"+"_"+strconv.Itoa(db), m.GatewayConfig.DataBase.Redis.Host, m.GatewayConfig.DataBase.Redis.Password, db)
		Logger.Info("gateway redis 连接成功", zap.String("addr", m.GatewayConfig.DataBase.Redis.Host), zap.Int("db", db))
	}
	SetCurrentRedisClient("gateway_2")
}

func InitServerRedisMulti(dbs ...int) {
	for _, db := range dbs {
		InitRedis("server"+"_"+strconv.Itoa(db), m.SMConfig.DataBase.Host, m.SMConfig.DataBase.Password, db)
		Logger.Info("server redis 连接成功", zap.String("addr", m.SMConfig.DataBase.Host), zap.Int("db", db))
	}
	SetCurrentRedisClient("server_2")
}

func InitWvpRedisMulti(dbs ...int) {
	for _, db := range dbs {
		InitRedis("wvp"+"_"+strconv.Itoa(db), m.WVPConfig.DataBase.Redis.Host, m.WVPConfig.DataBase.Redis.Password, db)
		Logger.Info("wvp redis 连接成功", zap.String("addr", m.WVPConfig.DataBase.Redis.Host), zap.Int("db", db))
	}
	SetCurrentRedisClient("wvp_2")
}

func SetCurrentRedisClient(name string) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	client, ok := redisClients[name]
	if !ok {
		panic("redis client not found: " + name)
	}
	RDB = client

}

func CloseAllRedisClients() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for name, client := range redisClients {
		_ = client.Close()
		Logger.Info("redis client closed", zap.String("name", name))
	}
	redisClients = make(map[string]*redis.Client)
	RDB = nil
}

func GetRedisClient() *redis.Client {
	if RDB == nil {
		panic("current redis client is not set")
	}
	return RDB
}

func GetRedisClientByName(name string) *redis.Client {
	client, ok := redisClients[name]
	if !ok {
		panic("redis client not found: " + name)
	}
	RDB = client
	return client
}
