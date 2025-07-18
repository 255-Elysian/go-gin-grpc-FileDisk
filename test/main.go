package main

import (
	"context"
	"google.golang.org/grpc"
	pb "grpc-todolist-disk/idl/pb/files"
	"io"
	"log"
	"os"
	"time"
)

// 流式上传
func main() {
	conn, err := grpc.Dial("localhost:10004", grpc.WithInsecure())
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer conn.Close()

	client := pb.NewFilesServiceClient(conn)

	stream, err := client.BigFileUpload(context.Background())
	if err != nil {
		log.Fatal("创建流失败:", err)
	}

	file, err := os.Open("C:\\Users\\elisia\\Desktop\\杂物\\图片\\微信图片_20250718231505.png")
	if err != nil {
		log.Fatal("打开文件失败:", err)
	}
	defer file.Close()

	buf := make([]byte, 1024*1024) // 1MB
	total := int64(0)
	objectName := "3/" + time.Now().Format("20060102150405") + "_your_test_file.png"
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("读取失败:", err)
		}
		total += int64(n)
		req := &pb.BigFileUploadRequest{
			UserID:     3,
			Filename:   "your_test_file.png",
			ObjectName: objectName,
			Content:    buf[:n],
			IsLast:     false,
		}
		if err := stream.Send(req); err != nil {
			log.Fatal("发送失败:", err)
		}
	}
	// 最后发送 IsLast = true 的包
	lastReq := &pb.BigFileUploadRequest{
		UserID:     3,
		Filename:   "your_test_file.png",
		ObjectName: objectName,
		IsLast:     true,
		FileSize:   total,
	}
	if err := stream.Send(lastReq); err != nil {
		log.Fatal("发送最后分片失败:", err)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("接收响应失败:", err)
	}
	log.Printf("上传完成: %+v\n", resp)
}
