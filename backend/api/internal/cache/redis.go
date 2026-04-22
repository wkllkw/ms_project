package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(addr, password string, db int) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		fmt.Printf("redis connect error: %v\n", err)
		return err
	}

	fmt.Println("redis connected successfully")
	return nil
}

// Prefix 所有缓存 key 统一前缀，避免冲突
const Prefix = "ms:"

// Get 获取缓存值
func Get(key string) (string, error) {
	return Client.Get(context.Background(), Prefix+key).Result()
}

// Set 设置缓存（默认 1 小时过期）
func Set(key string, value interface{}, expiration ...time.Duration) error {
	exp := time.Hour
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return Client.Set(context.Background(), Prefix+key, value, exp).Err()
}

// SetJSON 序列化为 JSON 后存入缓存
func SetJSON(key string, value interface{}, expiration ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	exp := time.Hour
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return Client.Set(context.Background(), Prefix+key, data, exp).Err()
}

// GetJSON 从缓存读取并反序列化到 target
func GetJSON(key string, target interface{}) error {
	data, err := Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), target)
}

// Del 删除缓存
func Del(keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = Prefix + k
	}
	return Client.Del(context.Background(), prefixedKeys...).Err()
}

// Exists 判断 key 是否存在
func Exists(key string) (bool, error) {
	n, err := Client.Exists(context.Background(), Prefix+key).Result()
	return n > 0, err
}

// SAdd 向集合添加成员（用于在线状态等）
func SAdd(key string, members ...interface{}) error {
	return Client.SAdd(context.Background(), Prefix+key, members...).Err()
}

// SRem 从集合移除成员
func SRem(key string, members ...interface{}) error {
	return Client.SRem(context.Background(), Prefix+key, members...).Err()
}

// SMembers 获取集合所有成员
func SMembers(key string) ([]string, error) {
	return Client.SMembers(context.Background(), Prefix+key).Result()
}

// HSet 哈希表设置字段
func HSet(key string, values ...interface{}) error {
	return Client.HSet(context.Background(), Prefix+key, values...).Err()
}

// HGet 哈希表获取字段
func HGet(key, field string) (string, error) {
	return Client.HGet(context.Background(), Prefix+key, field).Result()
}

// IsAvailable 检查 Redis 是否可用
func IsAvailable() bool {
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := Client.Ping(ctx).Err()
	return err == nil
}
