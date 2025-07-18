package model

import "gorm.io/gorm"

type Files struct {
	gorm.Model
	UserID     uint   `gorm:"index"`
	FileName   string `gorm:"type:varchar(255)"`
	FileSize   int64
	Bucket     string `gorm:"type:varchar(64)"`         // 存储桶名称（如 MinIO 的 bucket）
	ObjectName string `gorm:"type:varchar(255);unique"` // 存储对象名（唯一标识）
}
