package middleware

import (
	"github.com/gin-gonic/gin"
	"grpc-todolist-disk/utils/ctl"
	"grpc-todolist-disk/utils/token"
	"net/http"
	"strings"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		//log.Println("进入 JWT 中间件")

		// 从请求头中获取 Authorization 字段
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "缺少 Authorization 头",
				"code": "401",
			})
			c.Abort()
			return
		}

		// 拆解 Authorization 格式：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "Authorization 格式错误，应为 Bearer {token}",
				"code": "401",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 校验 token 是否有效
		if err := token.CheckRS(tokenStr); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "token 无效或已过期",
				"code": "401",
			})
			c.Abort()
			return
		}

		// 解析 token，提取 claims
		var claims token.UserClaims
		if err := token.Rs.Decode(tokenStr, &claims); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "token 解码失败",
				"code": "401",
			})
			c.Abort()
			return
		}
		//log.Println("用户ID：", claims.UserID)
		c.Request = c.Request.WithContext(ctl.NewContext(c.Request.Context(), &ctl.UserInfo{ID: claims.UserID}))
		c.Next()
	}
}
