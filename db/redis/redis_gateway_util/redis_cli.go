package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	. "go-sip/db/redis"
	. "go-sip/logger"
	"go-sip/model"

	"encoding/json"
	"time"
)

var ctx = context.Background()

// /////////// gateway_2 /////////////

func HSetStruct_2[T any](hash_key, key string, val T) error {
	rdb := GetRedisClientByName("gateway_2")
	data, err := json.Marshal(val)
	if err != nil {
		Logger.Error("redis HSetStruct 序列化错误")
		return err
	}
	err = rdb.HSet(ctx, hash_key, key, data).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
		return err
	}
	return nil
}

func HSet_2(hash_key string, key string, val string) error {
	rdb := GetRedisClientByName("gateway_2")
	err := rdb.HSet(ctx, hash_key, key, val).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
	}
	return err
}

func HGet_2(hash_key, key string) (string, error) {
	rdb := GetRedisClientByName("gateway_2")
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
	rdb := GetRedisClientByName("gateway_2")

	err := rdb.HDel(ctx, hash_key, key).Err()
	if err != nil {
		Logger.Error("redis hdel 错误")
	}
	return err
}

func HGetAll_2(hash_key string) (map[string]string, error) {
	rdb := GetRedisClientByName("gateway_2")
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
	rdb := GetRedisClientByName("gateway_2")

	err := rdb.Set(ctx, key, val, expiration).Err()
	if err != nil {
		Logger.Error("redis set 错误")
	}
	return err
}

func Get_2(key string) (string, error) {
	rdb := GetRedisClientByName("gateway_2")

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
	rdb := GetRedisClientByName("gateway_2")

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		Logger.Error("redis del 错误")
	}
	return err
}

func SAdd_2(key string, val string) error {
	rdb := GetRedisClientByName("gateway_2")
	err := rdb.SAdd(ctx, key, val).Err()
	if err != nil {
		Logger.Error("redis sadd 错误")
	}
	return err
}

func SMembers_2(key string) ([]string, error) {
	rdb := GetRedisClientByName("gateway_2")
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

func ZAdd_2(key string, items []model.IpcPlaybackRecordData, expiration time.Duration) error {
	rdb := GetRedisClientByName("gateway_2")
	zMembers := make([]redis.Z, 0, len(items))

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			Logger.Error("IpcPlaybackRecordData 序列化失败", zap.Error(err))
			return err
		}
		zMembers = append(zMembers, redis.Z{
			Score:  float64(item.StartTime), // 作为 score
			Member: string(data),            // 序列化 JSON 存入
		})
	}

	pipe := rdb.TxPipeline()
	pipe.ZAdd(ctx, key, zMembers...)
	pipe.Expire(ctx, key, expiration) // 设置 key 的过期时间
	_, err := pipe.Exec(ctx)
	if err != nil {
		Logger.Error("redis zadd+expire 错误", zap.Error(err))
	}
	return err
}

func ZRangeByScore_2(key string, min, max string) ([]model.IpcPlaybackRecordData, error) {
	rdb := GetRedisClientByName("gateway_2")

	rawMembers, err := rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err == redis.Nil {
		return []model.IpcPlaybackRecordData{}, nil
	}
	if err != nil {
		Logger.Error("redis zrangebyscore 错误", zap.Error(err))
		return nil, err
	}

	result := make([]model.IpcPlaybackRecordData, 0, len(rawMembers))
	for _, raw := range rawMembers {
		var item model.IpcPlaybackRecordData
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			Logger.Error("IpcPlaybackRecordData 反序列化失败", zap.String("raw", raw), zap.Error(err))
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// ScanPlaybackRecords 扫描匹配的 redis key，并反序列化成结构体
func ScanPlaybackRecords(pattern, min, max string) ([]model.IpcPlaybackRecordData, error) {
	rdb := GetRedisClientByName("gateway_2")
	var (
		cursor uint64
		result []model.IpcPlaybackRecordData
	)
	for {
		// SCAN 游标迭代
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			Logger.Error("redis scan 错误", zap.Error(err))
			return nil, err
		}
		// 遍历拿到的 keys
		for _, key := range keys {
			rawMembers, err := rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
				Min: min,
				Max: max,
			}).Result()
			if err != nil {
				Logger.Error("redis zrange 错误", zap.String("key", key), zap.Error(err))
				continue
			}
			for _, raw := range rawMembers {
				var item model.IpcPlaybackRecordData
				if err := json.Unmarshal([]byte(raw), &item); err != nil {
					Logger.Error("IpcPlaybackRecordData 反序列化失败", zap.String("raw", raw), zap.Error(err))
					continue
				}
				result = append(result, item)
			}
		}
		// 游标归零说明扫描结束
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return result, nil
}

///////////// gateway_2 /////////////

///////////// gateway_4 /////////////

func HSet_4(hash_key string, key string, val string) error {
	rdb := GetRedisClientByName("gateway_4")
	err := rdb.HSet(ctx, hash_key, key, val).Err()
	if err != nil {
		Logger.Error("redis hset 错误")
	}
	return err
}

func HGet_4(hash_key, key string) (string, error) {
	rdb := GetRedisClientByName("gateway_4")
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
	rdb := GetRedisClientByName("gateway_4")

	err := rdb.HDel(ctx, hash_key, key).Err()
	if err != nil {
		Logger.Error("redis hdel 错误")
	}
	return err
}

func HGetAll_4(hash_key string) (map[string]string, error) {
	rdb := GetRedisClientByName("gateway_4")
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
	rdb := GetRedisClientByName("gateway_4")

	err := rdb.Set(ctx, key, val, expiration).Err()
	if err != nil {
		Logger.Error("redis set 错误")
	}
	return err
}

func Get_4(key string) (string, error) {
	rdb := GetRedisClientByName("gateway_4")

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
	rdb := GetRedisClientByName("gateway_4")

	err := rdb.Del(ctx, key).Err()
	if err != nil {
		Logger.Error("redis del 错误")
	}
	return err
}

func SAdd_4(key string, val string) error {
	rdb := GetRedisClientByName("gateway_4")
	err := rdb.SAdd(ctx, key, val).Err()
	if err != nil {
		Logger.Error("redis sadd 错误")
	}
	return err
}

///////////// gateway_4 /////////////
