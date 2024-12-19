package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ljinf/user_auth/common/app"
	"github.com/ljinf/user_auth/common/errcode"
	"github.com/ljinf/user_auth/logic/domainservice"
)

// 用户认证相关的中间件

func AuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("go-mall-token")
		if len(token) != 40 { // 我们生成的token长度为40
			app.NewResponse(c).Error(errcode.ErrToken)
			c.Abort()
			return
		}
		tokenVerify, err := domainservice.NewUserDomainSvc(c).VerifyAccessToken(token)
		if err != nil { // 验证Token时服务出错
			app.NewResponse(c).Error(errcode.ErrServer)
			c.Abort()
			return
		}
		if !tokenVerify.Approved { // Token未通过验证
			app.NewResponse(c).Error(errcode.ErrToken)
			c.Abort()
			return
		}
		c.Set("userId", tokenVerify.UserId)
		c.Set("sessionId", tokenVerify.SessionId)
		c.Next()
	}
}
