package service

import (
	"container/heap"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/utils/kafka_mq"
	"grpc-todolist-disk/utils/redis_cache"

	"log"
)

// Message 解析Kafka消息中的JSON
type Message struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

var RDB *redis.Client
var KfReader *kafka.Reader
var TaskHeap DelayedTaskHeap

func ClearNameRedisCache(name string) bool {
	// 连接redis
	rdb := RDB
	ctx := context.Background()
	key := fmt.Sprintf("user:%s", name)
	_, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to clear Redis cache for user ID %s: %v", name, err)
		return false
	}
	return true
}

func Init() {
	RDB = redis_cache.ConnectRedis()
	KfReader = kafka_mq.NewKafkaConsumer()
	heap.Init(&TaskHeap)
	MsgSignal = make(chan int)
}
