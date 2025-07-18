package model

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username    string    `gorm:"unique" json:"username"`             // 唯一用户名，用于登录
	Nickname    string    `gorm:"type: varchar(100)" json:"nickname"` // 昵称，可重复
	Password    string    `gorm:"type: varchar(100); not null;" json:"password" binding:"required"`
	PasswordTry uint      `json:"password_try"` // 登录密码重试次数，超过指定数额锁定账号
	LockedUntil time.Time `json:"locked_until"` // 锁定账号的时间
	Status      string    `json:"status"`       // 当前状态，登录（in）或登出（out）
}

// SetPassword 加密密码
func (user *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return nil
}

// CheckPassword 校验密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
