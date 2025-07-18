package cache

import (
	"github.com/go-redsync/redsync/v4"
	"log"
	"time"
)

func GetRedLock(name string) *redsync.Mutex {
	lock := RedSyncLock.NewMutex("lock:user:"+name,
		redsync.WithExpiry(time.Second*10),
		redsync.WithTries(15), // 设置尝试获取锁的最大次数为5次
		redsync.WithRetryDelay(time.Second*1))
	return lock
}

// ContinueLock 续期锁，每 5 秒尝试一次，直到收到 stop 信号或续租失败
func ContinueLock(lock *redsync.Mutex, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ok, err := lock.Extend()
			if !ok {
				log.Println("锁续期失败:", err)
				return
			}
		case <-stop:
			return
		}
	}
}
