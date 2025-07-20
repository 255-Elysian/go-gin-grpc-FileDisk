package service

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/app/files/dao"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/kafka_mq"
	"grpc-todolist-disk/utils/redis_cache"
	"log"
	"os"
	"path/filepath"
)

// Message 解析Kafka消息中的JSON
type Message struct {
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

var RDB *redis.Client
var KfReader *kafka.Reader
var TaskHeap DelayedTaskHeap
var KfFileReader *kafka.Reader

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

type AsyncFileUploadMsg struct {
	UserID     uint64 `json:"user_id"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	FileHash   string `json:"file_hash"`
	ObjectName string `json:"object_name"`
	Content    []byte `json:"content"` // 小文件内容
	TempPath   string `json:"temp_path"`
}

// HandleAsyncFileUpload 异步启动上传文件的消费者（表单）
func HandleAsyncFileUpload(msg *kafka.Message) error {
	var m AsyncFileUploadMsg
	if err := json.Unmarshal(msg.Value, &m); err != nil {
		return fmt.Errorf("解析文件上传消息失败: %w", err)
	}

	log.Println("开始异步处理文件：", m.Filename)

	savePath := filepath.Join("stores/uploaded_files", m.ObjectName)
	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(savePath, m.Content, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 写入数据库
	if dao.DB == nil {
		log.Fatal("dao.DB 未初始化")
	}
	if _, err := dao.NewFilesDao().CreateFile(&pb.FileUploadRequest{
		UserID:     m.UserID,
		Filename:   m.Filename,
		FileSize:   m.FileSize,
		ObjectName: m.ObjectName,
		FileHash:   m.FileHash,
	}); err != nil {
		return fmt.Errorf("数据库写入失败: %w", err)
	}

	log.Println("文件处理成功: ", m.Filename)
	return nil
}

func Init() {
	dao.InitDB()
	RDB = redis_cache.ConnectRedis()
	KfReader = kafka_mq.NewKafkaConsumer()
	KfFileReader = kafka_mq.NewFileKafkaConsumer()
	heap.Init(&TaskHeap)
	MsgSignal = make(chan int)
}
