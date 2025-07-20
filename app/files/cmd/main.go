package main

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"grpc-todolist-disk/app/files/dao"
	"grpc-todolist-disk/app/files/internal/service"
	"grpc-todolist-disk/conf"
	pb "grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/utils/discovery"
	"net"
)

func main() {
	conf.InitConfig()
	dao.InitDB()
	// etcd 地址
	etcdAddress := []string{conf.Conf.Etcd.Endpoints[0]}
	// 注册服务
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	etcdRegister := discovery.NewRegister(etcdAddress, logger)
	grpcAddress := conf.Conf.Services["files"].Addr[0]
	defer etcdRegister.Stop()
	filesNode := discovery.Server{
		Name: conf.Conf.Services["files"].Name,
		Addr: grpcAddress,
	}
	server := grpc.NewServer()
	defer server.Stop()
	// 绑定service
	pb.RegisterFilesServiceServer(server, service.GetFilesSrv())
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	if _, err := etcdRegister.Register(filesNode, 10); err != nil {
		panic(fmt.Sprintf("start server failed, err: %v", err))
	}
	logger.Info("gRPC server started",
		zap.String("address", grpcAddress),
		zap.String("service", "files"),
	)
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}
