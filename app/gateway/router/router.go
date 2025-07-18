package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"grpc-todolist-disk/app/gateway/http"
	"grpc-todolist-disk/app/gateway/middleware"
	"grpc-todolist-disk/utils/logger"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.Cors())

	serverLogger, _ := logger.InitLogger(zap.DebugLevel)
	defer serverLogger.Sync()
	router.Use(logger.GinLogger(serverLogger), logger.GinRecovery(serverLogger, true))

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	v1 := router.Group("/api/v1")
	{
		v1.GET("ping", func(c *gin.Context) {
			c.JSON(200, "pong")
		})

		// 用户服务
		v1.POST("/user/register", http.UserRegister)
		v1.POST("/user/login", http.UserLogin)

		// 需要登录保护
		authed := v1.Group("/")
		authed.Use(middleware.JWT())
		{
			// 用户部分
			authed.PUT("/user/update_password", http.UserChangePassword)
			authed.POST("/user/logout", http.UserLogout)
			authed.DELETE("/user/delete", http.UserDelete)

			// 任务模块
			authed.GET("task", http.GetTaskList)
			authed.POST("task", http.CreateTask)
			authed.PUT("task", http.UpdateTask)
			authed.DELETE("task", http.DeleteTask)
			authed.GET("task/show", http.ShowTask)

			// 文件模块
			authed.POST("file_upload", http.FileUpload)
			authed.POST("big_file_upload", http.BigFileUpload)
			authed.GET("file_list", http.FileList)
			authed.DELETE("file_delete", http.FileDelete)
			authed.GET("file_download", http.FileDownload)
		}
	}

	return router
}
