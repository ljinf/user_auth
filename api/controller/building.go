package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ljinf/user_auth/common/app"
	"github.com/ljinf/user_auth/common/errcode"
	"github.com/ljinf/user_auth/common/logger"
	"github.com/ljinf/user_auth/config"
	"github.com/ljinf/user_auth/logic/appservice"
	"net/http"
)

// 存放一些项目搭建过程中验证效果用的接口Handler, 之前搭建过程中在main包中写的测试接口也挪到了这里

func TestPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
	return
}

func TestConfigRead(c *gin.Context) {
	database := config.Database
	c.JSON(http.StatusOK, gin.H{
		"type":     database.Type,
		"max_life": database.Master.MaxLifeTime,
	})
	return
}

func TestLogger(c *gin.Context) {
	logger.New().Info(c, "logger test", "key", "keyName", "val", 2)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
	return
}

func TestAccessLog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
	return
}

func TestPanicLog(c *gin.Context) {
	var a map[string]string
	a["k"] = "v"
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   a,
	})
	return
}

func TestAppError(c *gin.Context) {

	// 使用 Wrap 包装原因error 生成 项目error
	err := errors.New("a dao error")
	appErr := errcode.Wrap("包装错误", err)
	bAppErr := errcode.Wrap("再包装错误", appErr)
	logger.New().Error(c, "记录错误", "err", bAppErr)

	// 预定义的ErrServer, 给其追加错误原因的error
	err = errors.New("a domain error")
	apiErr := errcode.ErrServer.WithCause(err)
	logger.New().Error(c, "API执行中出现错误", "err", apiErr)

	c.JSON(apiErr.HttpStatusCode(), gin.H{
		"code": apiErr.Code(),
		"msg":  apiErr.Msg(),
	})
	return
}

func TestResponseObj(c *gin.Context) {
	data := map[string]int{
		"a": 1,
		"b": 2,
	}
	app.NewResponse(c).Success(data)
	return
}

func TestResponseList(c *gin.Context) {

	pagination := app.NewPagination(c)
	// Mock fetch list data from db
	data := []struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		{
			Name: "Lily",
			Age:  26,
		},
		{
			Name: "Violet",
			Age:  25,
		},
	}
	pagination.SetTotalRows(2)
	app.NewResponse(c).SetPagination(pagination).Success(data)
	return
}

func TestResponseError(c *gin.Context) {
	baseErr := errors.New("a dao error")
	// 这一步正式开发时写在service层
	err := errcode.Wrap("encountered an error when xxx service did xxx", baseErr)
	app.NewResponse(c).Error(errcode.ErrServer.WithCause(err))
	return
}

func TestMakeToken(c *gin.Context) {
	userSvc := appservice.NewUserAppSvc(c)
	token, err := userSvc.GenToken()
	if err != nil {
		if errors.Is(err, errcode.ErrUserInvalid) {
			logger.New().Error(c, "invalid user is unable to generate token", err)
			//app.NewResponse(c).Error(errcode.ErrUserInvalid.WithCause(err)) 第一版的AppError会导Error循环引用，现在已解决
			app.NewResponse(c).Error(errcode.ErrUserInvalid)
		} else {
			appErr := err.(*errcode.AppError)
			app.NewResponse(c).Error(appErr)
		}
		return
	}
	app.NewResponse(c).Success(token)
}

func TestRefreshToken(c *gin.Context) {
	refreshToken := c.Query("refresh_token")
	if refreshToken == "" {
		app.NewResponse(c).Error(errcode.ErrParams)
		return
	}
	userSvc := appservice.NewUserAppSvc(c)
	token, err := userSvc.TokenRefresh(refreshToken)
	if err != nil {
		if errors.Is(err, errcode.ErrTooManyRequests) {
			// 客户端有并发刷新token
			app.NewResponse(c).Error(errcode.ErrTooManyRequests)
			return
		} else {
			appErr := err.(*errcode.AppError)
			app.NewResponse(c).Error(appErr)
		}
		return
	}
	app.NewResponse(c).Success(token)
}

func TestAuthToken(c *gin.Context) {
	app.NewResponse(c).Success(gin.H{
		"user_id":    c.GetInt64("userId"),
		"session_id": c.GetString("sessionId"),
	})
	return
}
