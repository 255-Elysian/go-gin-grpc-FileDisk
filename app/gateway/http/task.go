package http

import (
	"github.com/gin-gonic/gin"
	"grpc-todolist-disk/app/gateway/rpc"
	pb "grpc-todolist-disk/idl/pb/task"
	"grpc-todolist-disk/utils/ctl"
	"net/http"
)

func GetTaskList(ctx *gin.Context) {
	var req pb.TaskRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "绑定参数错误"))
		return
	}
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	r, err := rpc.TaskShow(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "TaskShow RPC调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func CreateTask(ctx *gin.Context) {
	var req pb.TaskRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "绑定参数错误"))
		return
	}
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	r, err := rpc.TaskCreate(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "TaskShow RPC调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func UpdateTask(ctx *gin.Context) {
	var req pb.TaskRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "绑定参数错误"))
		return
	}
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	r, err := rpc.TaskUpdate(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "TaskShow RPC调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func DeleteTask(ctx *gin.Context) {
	var req pb.TaskRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "绑定参数错误"))
		return
	}
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	r, err := rpc.TaskDelete(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "TaskShow RPC调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func ShowTask(ctx *gin.Context) {
	var req pb.TaskRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "绑定参数错误"))
		return
	}
	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	req.UserID = uint64(user.ID)
	r, err := rpc.TaskShowOne(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "TaskShow RPC调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}
