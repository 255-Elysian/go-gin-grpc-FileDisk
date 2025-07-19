package utils

import (
	"log"
	"os"
)

func SafeRemove(path string) {
	if err := os.Remove(path); err != nil {
		log.Printf("清理文件失败: %v", err)
	}
}
