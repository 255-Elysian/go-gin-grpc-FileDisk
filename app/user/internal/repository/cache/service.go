package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/utils/kafka_mq"
	"grpc-todolist-disk/utils/redis_cache"
)

var RDB *redis.Client
var RedSyncLock *redsync.Redsync
var KfWriter *kafka.Writer

func Init() {
	RDB = redis_cache.ConnectRedis()
	RedSyncLock = redis_cache.NewSync(RDB)
	KfWriter = kafka_mq.NewKafkaProducer()
	ClearEntireRedisCache()
}
