package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"grpc-todolist-disk/app/user/internal/repository/model"
	"log"
	"time"
)

// ClearEntireRedisCache 清除redis缓存
func ClearEntireRedisCache() {
	ctx := context.Background()

	log.Println("清除redis")

	// 使用FlushDB命令清空当前数据库
	err := RDB.FlushDB(ctx).Err()
	if err != nil {
		log.Printf("Failed to flush redis cache: %v", err)
	} else {
		log.Printf("Successfully flush redis cache")
	}
}

// GetUserFromRedisByName 获取redis缓存的用户
func GetUserFromRedisByName(name string) *model.User {
	ctx := context.Background()
	var user *model.User
	key := fmt.Sprintf("user:%s", name)
	result, err := RDB.Get(ctx, key).Result()
	if err == nil {
		// 如果Redis中存在缓存，则直接返回
		//fmt.Println("从Redis缓存中获取数据")
		if err := json.Unmarshal([]byte(result), &user); err == nil {
			return user
		}
	}
	return nil
}

// SetNameToRedis 将查询结果缓存到redis
func SetNameToRedis(name string, user *model.User) bool {
	ctx := context.Background()
	key := fmt.Sprintf("user:%s", name)
	data, err := json.Marshal(user)
	err = RDB.Set(ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		log.Println("Redis缓存失败:", err)
		return false
	}
	return true
}

// ClearNameRedisCache 清除redis键值对
func ClearNameRedisCache(name string) {
	ctx := context.Background()
	rdb := RDB
	key := fmt.Sprintf("user:%s", name)
	_, err := rdb.Del(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to clear Redis cache for user ID %s: %v", name, err)
	}
}
