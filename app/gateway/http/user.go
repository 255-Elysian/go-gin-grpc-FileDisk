package http

import (
	"github.com/gin-gonic/gin"
	"grpc-todolist-disk/app/gateway/rpc"
	"grpc-todolist-disk/conf"
	pb "grpc-todolist-disk/idl/pb/user"
	"grpc-todolist-disk/utils/ctl"
	"grpc-todolist-disk/utils/e"
	"grpc-todolist-disk/utils/token"
	"net/http"
	"time"
)

func UserRegister(ctx *gin.Context) {
	var userReq pb.UserRequest
	if err := ctx.ShouldBind(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	// 两次密码校验放在前端
	r, err := rpc.UserRegister(ctx, &userReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "UserRegister RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func UserLogin(ctx *gin.Context) {
	var userReq pb.UserRequest
	if err := ctx.ShouldBind(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := rpc.UserLogin(ctx, &userReq)
	if err != nil || user == nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "UserLogin RPC服务调用错误"))
		return
	}

	shortDuration := time.Duration(conf.Conf.Token.ShortDuration)
	longDuration := time.Duration(conf.Conf.Token.LongDuration)
	// 1. 生成短期 token（access token）
	shortToken, err := token.IssueRS(uint(user.UserID), time.Now().Add(time.Minute*shortDuration))
	if err != nil {
		ctx.JSON(e.ErrorAuthToken, ctl.RespError(ctx, err, e.GetMsg(e.ErrorAuthToken)))
		return
	}
	// 2. 生成长期 token（refresh token）
	longToken, err := token.IssueRS(uint(user.UserID), time.Now().Add(time.Minute*longDuration))
	if err != nil {
		ctx.JSON(e.ErrorAuthToken, ctl.RespError(ctx, err, e.GetMsg(e.ErrorAuthToken)))
		return
	}
	// 3. 设置响应头：Authorization
	ctx.Header("Authorization", "Bearer "+shortToken)
	// 4. 返回 body：user + refresh token
	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, ctl.TokenData{User: user, Token: longToken}))
}

func UserLogout(ctx *gin.Context) {
	var userReq pb.UserRequest
	if err := ctx.ShouldBind(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	userReq.UserID = uint64(user.ID)

	r, err := rpc.UserLogout(ctx, &userReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "UserLogout RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func UserChangePassword(ctx *gin.Context) {
	var userReq pb.UserRequest
	if err := ctx.ShouldBind(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	userReq.UserID = uint64(user.ID)

	r, err := rpc.UserChangePassword(ctx, &userReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "UserChangePassword RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}

func UserDelete(ctx *gin.Context) {
	var userReq pb.UserRequest
	if err := ctx.ShouldBind(&userReq); err != nil {
		ctx.JSON(http.StatusBadRequest, ctl.RespError(ctx, err, "参数绑定错误"))
		return
	}

	user, err := ctl.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "获取用户信息错误"))
		return
	}
	userReq.UserID = uint64(user.ID)

	r, err := rpc.UserDelete(ctx, &userReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ctl.RespError(ctx, err, "UserDelete RPC服务调用错误"))
		return
	}

	ctx.JSON(http.StatusOK, ctl.RespSuccess(ctx, r))
}
