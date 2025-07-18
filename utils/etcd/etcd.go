package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"grpc-todolist-disk/conf"
	"log"
	"math/rand"
	"time"
)

func NewEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Conf.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Error connect to etcd client: %v", err)
	}
	return cli
}

func GetAddrFromEtcd(clientName string, etcdClient *clientv3.Client) string {
	// 从 etcd 获取服务地址
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := etcdClient.Get(ctx, clientName)
	cancel()
	if err != nil {
		log.Printf("failed to get service address from etcd: %v \n", err)
	}

	if len(resp.Kvs) == 0 {
		log.Println("service address not found in etcd")
	}

	// 随机选择一个服务地址，负载均衡
	rand.Seed(time.Now().UnixNano())
	serviceAddr := string(resp.Kvs[rand.Intn(len(resp.Kvs))].Value)

	return serviceAddr
}

func RegisterAddrToEtcd(serviceName string, serviceAddr string, etcdClient *clientv3.Client) {
	// 服务注册到etcd
	leaseResp, err := etcdClient.Grant(context.Background(), 10)
	if err != nil {
		log.Printf("failed to grant lease: %v\n", err)
	}

	// 设置键值对，其中键通常是服务名称，值是服务地址
	putResp, err := etcdClient.Put(context.Background(), serviceName, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		log.Printf("failed to put lease: %v\n", err)
	}
	log.Println(putResp, serviceName, serviceAddr)

	// 保持心跳，以续约租约
	keepAlive, err := etcdClient.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		log.Printf("failed to keep alive lease: %v\n", err)
	}

	for {
		select {
		case ka := <-keepAlive:
			if ka == nil {
				log.Println("Lease expired or KeepAlive channel closed")
				return
			}
		}
	}
}
