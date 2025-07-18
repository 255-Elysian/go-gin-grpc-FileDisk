package e

const (
	SUCCESS       = 200
	ERROR         = 500
	InvalidParams = 400

	//成员错误
	ErrorExistUser          = 10002
	ErrorNotExistUser       = 10003
	ErrorFailEncryption     = 10006
	ErrorNotCompare         = 10007
	ErrorUserLock           = 10008
	ErrorUserPassword       = 10009
	ErrorUserChangePassword = 10010

	HaveSignUp           = 20001
	ErrorActivityTimeout = 20002

	ErrorAuthCheckTokenFail    = 30001 //token 错误
	ErrorAuthCheckTokenTimeout = 30002 //token 过期
	ErrorAuthToken             = 30003
	ErrorAuth                  = 30004
	ErrorAuthNotFound          = 30005
	ErrorDatabase              = 40001

	ErrorServiceUnavailable = 50003
	ErrorDeadline           = 50004
)
