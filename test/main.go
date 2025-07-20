package main

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-todolist-disk/idl/pb/files"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// 流式上传
func main() {
	//g()
	k()
}

// grpc同步处理
func g() {
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

// kafka异步处理
func k() {
	filePath := "C:\\Users\\elisia\\Desktop\\杂物\\图片\\120753427_p0.jpg"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", "test_async_upload.png")
	if err != nil {
		log.Fatalf("创建表单文件字段失败: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		log.Fatalf("复制文件内容失败: %v", err)
	}

	// 可选：添加其他表单字段（如 token 等）
	// _ = writer.WriteField("token", "xxx")

	if err := writer.Close(); err != nil {
		log.Fatalf("关闭表单写入器失败: %v", err)
	}

	// 构造 HTTP 请求
	req, err := http.NewRequest("POST", "http://localhost:4000/api/v1/big_upload", body)
	if err != nil {
		log.Fatalf("构建请求失败: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTMxMTIxNjgsIm5iZiI6MTc1MzAyNTc2OCwiaWF0IjoxNzUzMDI1NzY4LCJ1c2VyX2lkIjo1fQ.WRF_I8R9r107WqvdvJ5PKEQrepMBYJ6ImumxElRSwvKjdoxgr56clWqC-ljkG9fn-hICd9Qq7ItASrQ21DSaDkjWZKTM_Mxw9ExrmMZjgXHfDXxKh0lWw37aLb9VRf7OrIW8RIOwjdAwYwwhtQB9KUFBdal6mei5mhEofAfx00H02XxPjuvROUfADk3uybIsY7RoeJ8mJOgu0j98OOKIXT6diM4OkBp47jwclBpdzyVqgg50ZhR-Av6sdtcP0E0B-e4yV4MPfFx_n-cL6awjv9ay-rY2B0qLuz6FPCHNZ61dVcvP4iJeT3qMXAFL3U9pZ49sc3ja0l2a8-Hd9fJJLw") // 添加授权头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 打印响应
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("响应状态码:", resp.StatusCode)
	fmt.Println("响应内容:", string(respBody))
}
