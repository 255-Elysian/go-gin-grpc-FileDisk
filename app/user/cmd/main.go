package main

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"grpc-todolist-disk/app/user/internal/repository/cache"
	"grpc-todolist-disk/app/user/internal/repository/dao"
	"grpc-todolist-disk/app/user/internal/service"
	"grpc-todolist-disk/conf"
	pb "grpc-todolist-disk/idl/pb/user"
	"grpc-todolist-disk/utils/discovery"
	"net"
)

func main() {
	conf.InitConfig()
	dao.InitDB()
	cache.Init()
	// etcd 地址
	etcdAddress := []string{conf.Conf.Etcd.Endpoints[0]}
	// 注册服务
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	etcdRegister := discovery.NewRegister(etcdAddress, logger)
	grpcAddress := conf.Conf.Services["user"].Addr[0]
	defer etcdRegister.Stop()
	userNode := discovery.Server{
		Name: conf.Conf.Services["user"].Name,
		Addr: grpcAddress,
	}
	server := grpc.NewServer()
	defer server.Stop()
	// 绑定service
	pb.RegisterUserServiceServer(server, service.GetUserSrv())
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	if _, err := etcdRegister.Register(userNode, 10); err != nil {
		panic(fmt.Sprintf("start server failed, err: %v", err))
	}
	logger.Info("gRPC server started",
		zap.String("address", grpcAddress),
		zap.String("service", "user"),
	)
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}
