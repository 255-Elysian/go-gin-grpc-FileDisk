package model

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	UserID    uint `gorm:"index"`
	Title     string
	Status    int    `gorm:"default:0"`
	Content   string `gorm:"type:longtext"`
	StartTime int64
	EndTime   int64
}
