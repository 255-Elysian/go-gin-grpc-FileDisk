package mq

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
)

// AsyncFileUploadMsg 表示文件异步上传任务的消息结构
type AsyncFileUploadMsg struct {
	UserID     uint64 `json:"user_id"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	FileHash   string `json:"file_hash"`
	ObjectName string `json:"object_name"`
	Content    []byte `json:"content"` // 小文件内容
	TempPath   string `json:"temp_path"`
}

func SendFileUploadTask(msg *AsyncFileUploadMsg) error {
	value, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Kafka Msg JSON 序列化失败: %v", err)
		return err
	}

	err = KfWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(msg.FileHash), // 可以用 Hash 做 Key 便于排查
		Value: value,
	})
	if err != nil {
		log.Printf("Kafka 消息发送失败: %v", err)
		return err
	}

	log.Printf("Kafka 消息发送成功: %s", msg.Filename)
	return nil
}
