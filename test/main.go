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
	//g()  // 本地流式上传
	//k()  // 异步上传
	//qiniuFileUpload() // 七牛云表单上传
	qiniuBigFileUpload() // 七牛云流式上传
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

// 七牛云流式上传测试
func qiniuBigFileUpload() {
	// 连接到 Files 服务
	conn, err := grpc.Dial("localhost:10004", grpc.WithInsecure())
	if err != nil {
		log.Fatal("连接 Files 服务失败:", err)
	}
	defer conn.Close()

	client := pb.NewFilesServiceClient(conn)

	// 创建七牛云流式上传流
	stream, err := client.QiniuBigFileUpload(context.Background())
	if err != nil {
		log.Fatal("创建七牛云上传流失败:", err)
	}

	// 打开要上传的文件
	filePath := "E:\\download\\单依纯 - 永不失联的爱 (Live).flac"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("打开文件失败:", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal("获取文件信息失败:", err)
	}

	fmt.Printf("开始上传文件: %s\n", filePath)
	fmt.Printf("文件大小: %d bytes\n", fileInfo.Size())

	// 分片上传
	buf := make([]byte, 1024*1024) // 1MB 分片
	totalSize := int64(0)
	chunkCount := 0
	isFirst := true

	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}

		chunkCount++
		totalSize += int64(n)
		isLast := err == io.EOF

		req := &pb.BigFileUploadRequest{
			Content: buf[:n],
			IsLast:  isLast,
		}

		// 第一个分片包含文件信息
		if isFirst {
			req.UserID = 2 // 测试用户ID
			req.Filename = "单依纯 - 永不失联的爱 (Live).flac"
			isFirst = false
		}

		fmt.Printf("发送分片 %d: %d bytes, isLast: %v\n", chunkCount, n, isLast)

		if err := stream.Send(req); err != nil {
			log.Fatal("发送分片失败:", err)
		}

		if isLast {
			break
		}
		if err != nil && err != io.EOF {
			log.Fatal("读取文件失败:", err)
		}
	}

	// 接收响应
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("接收响应失败:", err)
	}

	fmt.Printf("\n=== 上传完成 ===\n")
	fmt.Printf("响应码: %d\n", resp.Code)
	fmt.Printf("消息: %s\n", resp.Msg)
	fmt.Printf("文件ID: %d\n", resp.FileID)
	fmt.Printf("访问URL: %s\n", resp.ObjectUrl)
	fmt.Printf("总共发送: %d 个分片\n", chunkCount)
	fmt.Printf("总大小: %d bytes\n", totalSize)
}

// 七牛云表单上传测试
func qiniuFileUpload() {
	filePath := "C:\\Users\\elisia\\Desktop\\杂物\\Linux 安装及使用.md"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", "Linux 安装及使用.md")
	if err != nil {
		log.Fatalf("创建表单文件字段失败: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		log.Fatalf("复制文件内容失败: %v", err)
	}

	if err := writer.Close(); err != nil {
		log.Fatalf("关闭表单写入器失败: %v", err)
	}

	// 构造 HTTP 请求
	req, err := http.NewRequest("POST", "http://localhost:4000/api/v1/qiniu_file_upload", body)
	if err != nil {
		log.Fatalf("构建请求失败: %v", err)
	}

	// 添加授权头（请替换为有效的 token）
	// 注意：这个 token 可能已过期，请使用有效的 JWT token
	req.Header.Set("Authorization", "Bearer "+"YOUR_VALID_JWT_TOKEN_HERE")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("开始七牛云表单上传: %s\n", filePath)

	// 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 打印响应
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n=== 七牛云表单上传完成 ===\n")
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(respBody))
}
