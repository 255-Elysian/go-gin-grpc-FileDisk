package http

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"grpc-todolist-disk/app/files/dao"
	pb "grpc-todolist-disk/idl/pb/files"
	"log"
	"os"
	"path/filepath"
)

type AsyncFileUploadMsg struct {
	UserID     uint64 `json:"user_id"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	FileHash   string `json:"file_hash"`
	ObjectName string `json:"object_name"`
	Content    []byte `json:"content"` // 小文件内容
}

// HandleAsyncFileUpload 异步启动上传文件的消费者
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
