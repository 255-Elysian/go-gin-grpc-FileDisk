package redis_cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"grpc-todolist-disk/conf"
)

// ConnectRedis 连接 redis
func ConnectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     conf.Conf.Redis.Address,
		Password: conf.Conf.Redis.Password,
		DB:       0,
	})
}

// NewSync 创建一个 redis 分布式锁实例
func NewSync(rdb *redis.Client) *redsync.Redsync {
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	return rs
}
