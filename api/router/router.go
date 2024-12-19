package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ljinf/user_auth/common/middleware"
)

func RegisterRoutes(engine *gin.Engine) {
	// use global middlewares
	engine.Use(middleware.StartTrace(), middleware.LogAccess(), middleware.GinPanicRecovery())
	routeGroup := engine.Group("")
	registerBuildingRoutes(routeGroup)
}
