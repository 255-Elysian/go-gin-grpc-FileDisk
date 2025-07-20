package service

import (
	"container/heap"
	"github.com/segmentio/kafka-go"
)

type TaskType int

const (
	TaskTypeClearCache TaskType = iota
	TaskTypeFileUpload
)

// DelayedTask 是一个延时任务结构体
type DelayedTask struct {
	Name      string // 任务名称
	Timestamp int64  // 执行任务的时间戳
	Msg       *kafka.Message
	Index     int      // 在堆中的索引
	Type      TaskType // 新增任务类型：区分是清缓存还是文件上传
}

// DelayedTaskHeap 是一个延时任务堆
type DelayedTaskHeap []*DelayedTask

func (h DelayedTaskHeap) Len() int {
	return len(h)
}

func (h DelayedTaskHeap) Less(i, j int) bool {
	return h[i].Timestamp < h[j].Timestamp
}

func (h DelayedTaskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}
func (h *DelayedTaskHeap) Push(x interface{}) {
	n := len(*h)
	task := x.(*DelayedTask)
	task.Index = n
	*h = append(*h, task)
}
func (h *DelayedTaskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	task.Index = -1
	*h = old[0 : n-1]
	return task
}

// 修改堆中的任务，如果任务的时间戳改变了，可能需要重新调整堆
func (h *DelayedTaskHeap) update(task *DelayedTask, name string, timestamp int64) {
	task.Timestamp = timestamp
	task.Name = name
	heap.Fix(h, task.Index)
}
