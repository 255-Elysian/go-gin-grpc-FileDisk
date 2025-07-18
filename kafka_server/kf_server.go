package main

import (
	"grpc-todolist-disk/conf"
	"grpc-todolist-disk/kafka_server/service"
	"log"
	"sync"
)

func main() {
	conf.InitConfig()
	service.Init()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		service.MsgConsumer()
	}()

	go func() {
		defer wg.Done()
		service.TaskWorker()
	}()

	log.Println("kafka consumer running")
	wg.Wait()
}
