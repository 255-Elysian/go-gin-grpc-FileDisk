package rpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"grpc-todolist-disk/conf"
	"grpc-todolist-disk/idl/pb/files"
	"grpc-todolist-disk/idl/pb/task"
	"grpc-todolist-disk/idl/pb/user"
	"grpc-todolist-disk/utils/discovery"
	"log"
	"time"
)

var (
	Register   *discovery.Resolver
	ctx        context.Context
	CancelFunc context.CancelFunc

	UserClient  user.UserServiceClient
	TaskClient  task.TaskServiceClient
	FilesClient files.FilesServiceClient
)

func Init() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	Register = discovery.NewResolver([]string{conf.Conf.Etcd.Endpoints[0]}, logger)
	resolver.Register(Register)
	ctx, CancelFunc = context.WithTimeout(context.Background(), 10*time.Second)
	defer Register.Close()

	initClient(conf.Conf.Domain["user"].Name, &UserClient)
	initClient(conf.Conf.Domain["task"].Name, &TaskClient)
	initClient(conf.Conf.Domain["files"].Name, &FilesClient)
}

func initClient(serverName string, client interface{}) {
	conn, err := connectServer(serverName)
	if err != nil {
		panic(err)
	}

	switch c := client.(type) {
	case *user.UserServiceClient:
		*c = user.NewUserServiceClient(conn)
	case *task.TaskServiceClient:
		*c = task.NewTaskServiceClient(conn)
	case *files.FilesServiceClient:
		*c = files.NewFilesServiceClient(conn)
	default:
		panic("invalid client type")
	}
}

func connectServer(serverName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf("%s:///%s", Register.Scheme(), serverName)

	// 负载均衡
	if conf.Conf.Services[serverName].LoadBalance {
		log.Printf("load balance enabled for %s\n", serverName)
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, "round_robin")))
	}

	conn, err = grpc.DialContext(ctx, addr, opts...)
	return
}
