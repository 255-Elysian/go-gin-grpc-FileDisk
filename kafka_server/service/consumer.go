package service

import (
	"container/heap"
	"context"
	"encoding/json"
	"log"
	"time"
)

var MsgSignal chan int

// MsgConsumer Kafka 消费协程，从 Kafka 中消费消息 → 解析 → 转为延时任务 → 推入堆中 → 通知执行器
func MsgConsumer() {
	for {
		// 读取下一条消息
		ctx := context.Background()
		msg, err := KfReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %s\n", err)
			break
		}

		// 解析消息
		var m Message
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Printf("Error unmarshalling message: %s\n", err)
			continue
		}

		// 创建一个延时任务，并放入堆中
		task := &DelayedTask{
			Name:      m.Name,
			Timestamp: m.Timestamp,
			Msg:       &msg,
		}
		heap.Push(&TaskHeap, task)
		MsgSignal <- 1

		// 打印时间戳
		log.Printf("Message timestamp: %d\n", m.Timestamp)
	}
}

// TaskWorker 任务调度器，定时 + 被动监听 MsgSignal，执行堆中任务（即调用 task()）
func TaskWorker() {
	ticker := time.NewTicker(time.Second * 5) // 5秒一次
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			task()
		case <-MsgSignal:
			task()
		}
	}
}

/*
1、从堆顶取出任务（最早时间戳）

2、当前时间尚未到：放回堆中，等待下次执行

3、时间已到：执行 ClearNameRedisCache，成功则提交 Kafka offset，失败则时间延迟 1 秒后重新入堆（重试）
*/
func task() {
	// 处理堆中的任务
	for TaskHeap.Len() > 0 {
		task := heap.Pop(&TaskHeap).(*DelayedTask)

		if time.Now().Unix() < task.Timestamp {
			heap.Push(&TaskHeap, task)
			MsgSignal <- 1
			log.Println("Not To Time")
			continue
		}

		// 执行清除缓存操作
		ok := ClearNameRedisCache(task.Name)
		if !ok {
			// 删除失败，延迟一秒重新推入队列
			task.Timestamp = time.Now().Add(1 * time.Second).UnixNano()
			heap.Push(&TaskHeap, task)
			log.Printf("Failed to clear cache for %s, retrying in 1 second\n", task.Name)
			continue
		}

		ctx := context.Background()
		if err := KfReader.CommitMessages(ctx, *task.Msg); err != nil {
			log.Printf("Failed to commit messages: %s\n", err)
			continue
		}
		log.Printf("Cleared cache for %s at timestamp: %d\n", task.Name, task.Timestamp)
	}
}
