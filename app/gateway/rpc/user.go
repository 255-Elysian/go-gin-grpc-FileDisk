package rpc

import (
	"context"
	"errors"
	pb "grpc-todolist-disk/idl/pb/user"
	"grpc-todolist-disk/utils/e"
)

func UserRegister(ctx context.Context, req *pb.UserRequest) (resp *pb.UserCommonResponse, err error) {
	resp, err = UserClient.UserRegister(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func UserLogin(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	r, err := UserClient.UserLogin(ctx, req)
	if r.UserDetail == nil {
		err = errors.New("返回的用户为空")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if r.Code != e.SUCCESS {
		err = errors.New("登录失败")
		return nil, err
	}
	return r.UserDetail, nil
}

func UserLogout(ctx context.Context, req *pb.UserRequest) (resp *pb.UserCommonResponse, err error) {
	resp, err = UserClient.UserLogout(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func UserChangePassword(ctx context.Context, req *pb.UserRequest) (resp *pb.UserCommonResponse, err error) {
	resp, err = UserClient.UserChangePassword(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}

func UserDelete(ctx context.Context, req *pb.UserRequest) (resp *pb.UserCommonResponse, err error) {
	resp, err = UserClient.UserDelete(ctx, req)
	if err != nil {
		return
	}

	if resp.Code != e.SUCCESS {
		err = errors.New(resp.Msg)
		return
	}
	return
}
