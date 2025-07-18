package ctl

import (
	"context"
	"errors"
)

type ctxUserKey struct{} // 定义一个结构体类型作为唯一 key

var userKey = ctxUserKey{}

type UserInfo struct {
	ID uint `json:"id"`
}

func GetUserInfo(ctx context.Context) (*UserInfo, error) {
	user, ok := FromContext(ctx)
	//log.Println("GetUserInfo:", user, ok)
	if !ok {
		return nil, errors.New("获取用户信息失败")
	}
	return user, nil
}

func FromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(userKey).(*UserInfo)
	return user, ok
}

func NewContext(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, userKey, user)
}
