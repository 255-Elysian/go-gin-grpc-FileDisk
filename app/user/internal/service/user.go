package service

import (
	"context"
	"grpc-todolist-disk/app/user/internal/repository/dao"
	pb "grpc-todolist-disk/idl/pb/user"
	"grpc-todolist-disk/utils/e"
	"sync"
	"time"
)

var UserSrvIns *UserSrv
var UserSrvOnce sync.Once

type UserSrv struct {
	pb.UnimplementedUserServiceServer
}

func GetUserSrv() *UserSrv {
	UserSrvOnce.Do(func() {
		UserSrvIns = &UserSrv{}
	})
	return UserSrvIns
}

func (u *UserSrv) UserRegister(ctx context.Context, req *pb.UserRequest) (*pb.UserCommonResponse, error) {
	resp := &pb.UserCommonResponse{}
	resp.Code = e.SUCCESS
	user := dao.GetUserByName(req.Username)
	if user != nil {
		resp.Code = e.ErrorExistUser
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	//rawPassword, _ := utils.RsaDecode(req.Password)
	//newUser.Password = utils.GetHash(rawPassword)
	err := dao.NewUserDao().CreateUser(req)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		resp.Data = err.Error()
		return resp, nil
	}
	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}

func (u *UserSrv) UserLogin(ctx context.Context, req *pb.UserRequest) (*pb.UserDetailResponse, error) {
	resp := &pb.UserDetailResponse{}
	resp.Code = e.SUCCESS
	user := dao.GetUserByName(req.Username)
	//log.Printf("GetUserByName 返回 user: %+v\n", user)
	if user == nil {
		resp.Code = e.ErrorNotExistUser
		resp.UserDetail = nil
		return resp, nil
	}
	if time.Now().Before(user.LockedUntil) {
		timeTemplate := "2006-01-02 15:04:05"
		resp.Code = e.ErrorUserLock
		resp.Msg = e.GetMsg(int(resp.Code)) + "到：" + user.LockedUntil.Format(timeTemplate)
		return resp, nil
	}
	// 注意：当前使用 SHA256 + 固定盐，建议后期升级为 bcrypt
	//rawPassword, _ := utils.RsaDecode(req.Password)
	//loginPassword := utils.GetHash(rawPassword)
	ok := user.CheckPassword(req.Password)
	if !ok {
		resp.Code = e.ErrorUserPassword
		resp.Msg = e.GetMsg(int(resp.Code))
		dao.RecordPasswordWrong(user, user.PasswordTry+1)
		return resp, nil
	}
	// 设置redis的状态
	dao.RecordPasswordWrong(user, 0)
	dao.SetUserStatus(user, "in")
	resp.Msg = e.GetMsg(int(resp.Code))
	resp.UserDetail = &pb.UserResponse{
		UserID:   uint64(user.ID),
		Nickname: user.Nickname,
		Username: user.Username,
	}
	return resp, nil
}

func (u *UserSrv) UserLogout(ctx context.Context, req *pb.UserRequest) (*pb.UserCommonResponse, error) {
	resp := &pb.UserCommonResponse{}
	resp.Code = e.SUCCESS
	user, err := dao.NewUserDao().GetUserByUserID(uint(req.UserID))
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	if user == nil {
		resp.Code = e.ErrorNotExistUser
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	ok := dao.SetUserStatus(user, "out")
	if !ok {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}

func (u *UserSrv) UserChangePassword(ctx context.Context, req *pb.UserRequest) (*pb.UserCommonResponse, error) {
	resp := &pb.UserCommonResponse{}
	resp.Code = e.SUCCESS
	user, err := dao.NewUserDao().GetUserByUserID(uint(req.UserID))
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	if user == nil {
		resp.Code = e.ErrorNotExistUser
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	if time.Now().Before(user.LockedUntil) {
		timeTemplate := "2006-01-02 15:04:05"
		resp.Code = e.ErrorUserLock
		resp.Msg = e.GetMsg(int(resp.Code)) + "到：" + user.LockedUntil.Format(timeTemplate)
		return resp, nil
	}
	// 注意：当前使用 SHA256 + 固定盐，建议后期升级为 bcrypt
	//rawPassword, _ := utils.RsaDecode(req.Password)
	//password := utils.GetHash(rawPassword)
	ok := dao.SetUserPassword(user, req.Password)
	if !ok {
		resp.Code = e.ErrorUserChangePassword
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}

func (u *UserSrv) UserDelete(ctx context.Context, req *pb.UserRequest) (*pb.UserCommonResponse, error) {
	resp := &pb.UserCommonResponse{}
	resp.Code = e.SUCCESS
	user, err := dao.NewUserDao().GetUserByUserID(uint(req.UserID))
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	if user == nil {
		resp.Code = e.ErrorNotExistUser
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	err = dao.NewUserDao().DeleteUser(user.Username)
	if err != nil {
		resp.Code = e.ERROR
		resp.Msg = e.GetMsg(int(resp.Code))
		return resp, nil
	}
	resp.Msg = e.GetMsg(int(resp.Code))
	return resp, nil
}
