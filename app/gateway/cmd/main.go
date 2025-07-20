package main

import (
	"context"
	"errors"
	"grpc-todolist-disk/app/gateway/router"
	"grpc-todolist-disk/app/gateway/rpc"
	"grpc-todolist-disk/app/gateway/utils/mq"
	"grpc-todolist-disk/conf"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conf.InitConfig()
	mq.Init()
	rpc.Init()

	// 创建 Gin 路由和 HTTP Server 实例
	r := router.NewRouter()
	server := &http.Server{
		Addr:           conf.Conf.Server.Port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动 HTTP 监听（子协程）
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server start failed: %v", err)
		}
	}()
	log.Printf("gateway listen on: %s", conf.Conf.Server.Port)

	// 捕获退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// 创建带超时的 context 用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
