package dao

import (
	"grpc-todolist-disk/app/task/internal/repository/db/model"
	"log"
)

func migration() {
	// 自动迁移模式
	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(
			&model.Task{},
		)
	if err != nil {
		log.Println("register table failed")
		panic(err)
	}
	log.Println("register table success")
}
