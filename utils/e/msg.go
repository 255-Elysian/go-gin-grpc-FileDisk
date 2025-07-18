package e

var MsgFlags = map[int]string{
	SUCCESS: "ok",
	ERROR:   "failed",

	InvalidParams:              "请求参数错误",
	HaveSignUp:                 "已经报名了",
	ErrorActivityTimeout:       "活动过期了",
	ErrorAuthCheckTokenFail:    "Token鉴权失败",
	ErrorAuthCheckTokenTimeout: "Token已超时",
	ErrorAuthToken:             "Token生成失败",
	ErrorAuth:                  "Token错误",
	ErrorNotCompare:            "不匹配",
	ErrorDatabase:              "数据库操作出错,请重试",
	ErrorAuthNotFound:          "Token不能为空",

	ErrorServiceUnavailable: "过载保护，服务暂时不可用",
	ErrorDeadline:           "服务调用超时",

	ErrorExistUser:          "用户已存在",
	ErrorNotExistUser:       "用户不存在",
	ErrorFailEncryption:     "用户密码加密失败",
	ErrorUserLock:           "用户被锁定",
	ErrorUserPassword:       "用户密码错误",
	ErrorUserChangePassword: "用户修改密码错误",
}

// GetMsg 获取状态码对应信息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
