package rpc

import (
	"context"
	"errors"
	pb "grpc-todolist-disk/idl/pb/task"
	"grpc-todolist-disk/utils/e"
)

func TaskCreate(ctx context.Context, req *pb.TaskRequest) (resp *pb.TaskCommonResponse, err error) {
	r, err := TaskClient.TaskCreate(ctx, req)

	if err != nil {
		return
	}

	if r.Code != e.SUCCESS {
		err = errors.New(r.Msg)
		return
	}

	return r, nil
}

func TaskUpdate(ctx context.Context, req *pb.TaskRequest) (resp *pb.TaskCommonResponse, err error) {
	r, err := TaskClient.TaskUpdate(ctx, req)

	if err != nil {
		return
	}

	if r.Code != e.SUCCESS {
		err = errors.New(r.Msg)
		return
	}

	return r, nil
}

func TaskDelete(ctx context.Context, req *pb.TaskRequest) (resp *pb.TaskCommonResponse, err error) {
	r, err := TaskClient.TaskDelete(ctx, req)

	if err != nil {
		return
	}

	if r.Code != e.SUCCESS {
		err = errors.New(r.Msg)
		return
	}

	return r, nil
}

func TaskShow(ctx context.Context, req *pb.TaskRequest) (resp *pb.TasksDetailResponse, err error) {
	r, err := TaskClient.TaskShow(ctx, req)

	if err != nil {
		return
	}

	if r.Code != e.SUCCESS {
		err = errors.New("获取失败")
		return
	}

	return r, nil
}

func TaskShowOne(ctx context.Context, req *pb.TaskRequest) (resp *pb.TasksDetailResponse, err error) {
	r, err := TaskClient.TaskShowOne(ctx, req)
	if err != nil {
		return
	}

	if r.Code != e.SUCCESS {
		err = errors.New("获取失败")
		return
	}

	return r, nil
}
