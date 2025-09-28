package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	. "go-sip/logger"

	"time"

	. "go-sip/db/redis"

	"encoding/json"
)

var ctx = context.Background()

// /////////// server_2 /////////////

func HSetStruct_2[T any](hash_key, key string, val T) error {
	rdb := GetRedisClientByName("server_2")
	data, err := json.Marshal(val)
	if err != nil {
		Logger.Error("redis HSetStruct 序列化错误")
		return err
	}
	err = rdb.HSet(ctx, hash_key, key, string(data)).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
		return err
	}
	return nil
}

func HSet_2(hash_key string, key string, val string) error {
	rdb := GetRedisClientByName("server_2")
	err := rdb.HSet(ctx, hash_key, key, val).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
	}
	return err
}

// 判断是否存在hash
func HExists_2(hash_key, key string) bool {
	rdb := GetRedisClientByName("server_2")
	return rdb.HExists(ctx, hash_key, key).Val()
}

// 不存在则hset的方法
func HSetIfNotExist_2(hash_key, key string, val string) error {
	rdb := GetRedisClientByName("server_2")
	err := rdb.HSetNX(ctx, hash_key, key, val).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
	}
	return err
}

func HGet_2(hash_key, key string) (string, error) {
	rdb := GetRedisClientByName("server_2")
	val, err := rdb.HGet(ctx, hash_key, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		Logger.Error("redis hget 错误")
		return "", err
	}
	return val, err
}

func HDel_2(hash_key string, key string) error {
	rdb := GetRedisClientByName("server_2")

	err := rdb.HDel(ctx, hash_key, key).Err()
	if err != nil {
		Logger.Error("redis hdel 错误")
	}
	return err
}

func HGetAll_2(hash_key string) (map[string]string, error) {
	rdb := GetRedisClientByName("server_2")
	val, err := rdb.HGetAll(ctx, hash_key).Result()
	if err == redis.Nil {
		return map[string]string{}, nil
	}
	if err != nil {
		Logger.Error("redis hget 错误")
		return nil, err
	}
	return val, err
}

func Set_2(key string, val string, expiration time.Duration) error {
	rdb := GetRedisClientByName("server_2")

	err := rdb.Set(ctx, key, val, expiration).Err()
	if err != nil {
		Logger.Error("redis set 错误")
	}
	return err
}

func Get_2(key string) (string, error) {
	rdb := GetRedisClientByName("server_2")

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		if err == redis.Nil {
			Logger.Warn("redis get 警告: key 不存在", zap.String("key", key))
			return "", nil
		}
		Logger.Error("redis get 错误")
		return "", err
	}
	return val, nil
}

func Del_2(key string) error {
	rdb := GetRedisClientByName("server_2")

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		Logger.Error("redis del 错误")
	}
	return err
}

func SAdd_2(key string, val string) error {
	rdb := GetRedisClientByName("server_2")
	err := rdb.SAdd(ctx, key, val).Err()
	if err != nil {
		Logger.Error("redis sadd 错误")
	}
	return err
}

func SMembers_2(key string) ([]string, error) {
	rdb := GetRedisClientByName("server_2")
	members, err := rdb.SMembers(ctx, key).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		Logger.Error("redis SMembers 错误")
		return nil, err
	}
	return members, nil
}

// SetNX
func SetNX(key string, val string, expiration time.Duration) (bool, error) {
	rdb := GetRedisClientByName("server_2")
	cmd := rdb.SetNX(ctx, key, val, expiration)
	err := cmd.Err()
	if err == redis.Nil {
		return false, err
	}
	if err != nil {
		Logger.Error("redis SetNX 错误")
		return false, err
	}
	result := cmd.Val()
	return result, nil
}

///////////// server_2 /////////////

///////////// server_4 /////////////

func HSet_4(hash_key string, key string, val string) error {
	rdb := GetRedisClientByName("server_4")
	err := rdb.HSet(ctx, hash_key, key, val).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
	}
	return err
}

func HGet_4(hash_key, key string) (string, error) {
	rdb := GetRedisClientByName("server_4")
	val, err := rdb.HGet(ctx, hash_key, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		Logger.Error("redis hget 错误")
		return "", err
	}
	return val, err
}

func HDel_4(hash_key string, key string) error {
	rdb := GetRedisClientByName("server_4")

	err := rdb.HDel(ctx, hash_key, key).Err()
	if err != nil {
		Logger.Error("redis hdel 错误")
	}
	return err
}

func HGetAll_4(hash_key string) (map[string]string, error) {
	rdb := GetRedisClientByName("server_4")
	val, err := rdb.HGetAll(ctx, hash_key).Result()
	if err == redis.Nil {
		return map[string]string{}, nil
	}
	if err != nil {
		Logger.Error("redis hget 错误")
		return nil, err
	}
	return val, err
}

func Set_4(key string, val string, expiration time.Duration) error {
	rdb := GetRedisClientByName("server_4")

	err := rdb.Set(ctx, key, val, expiration).Err()
	if err != nil {
		Logger.Error("redis set 错误")
	}
	return err
}

func Get_4(key string) (string, error) {
	rdb := GetRedisClientByName("server_4")

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		if err == redis.Nil {
			Logger.Warn("redis get 警告: key 不存在", zap.String("key", key))
			return "", nil
		}
		Logger.Error("redis get 错误")
		return "", err
	}
	return val, nil
}

func Del_4(key string) error {
	rdb := GetRedisClientByName("server_4")

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		Logger.Error("redis del 错误")
	}
	return err
}

func SAdd_4(key string, val string) error {
	rdb := GetRedisClientByName("server_4")
	err := rdb.SAdd(ctx, key, val).Err()
	if err != nil {
		Logger.Error("redis sadd 错误")
	}
	return err
}

///////////// server_4 /////////////
