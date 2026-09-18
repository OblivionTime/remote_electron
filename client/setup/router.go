package setup

import (
	"net/http"

	api "remote/api/server"

	"github.com/gin-gonic/gin"
)

func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404 ,page not exists!",
	})
}

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}

// 设置路由
func setupRouter(router *gin.Engine) {
	//跨域
	router.Use(Cors())
	//未找到路由
	router.NoRoute(NotFound)
	/*
		用户
	*/
	remoteRouter := router.Group("/v1/api/remote")
	{
		serverRouter := remoteRouter.Group("/server")
		{
			serverRouter.GET("/connect", api.Connect)
			serverRouter.GET("/file_connect", api.FileConnect)
			serverRouter.GET("/device_list", api.ConnectedList)
			serverRouter.GET("/video_connect", api.ConnectVideo)
			serverRouter.GET("/keyboard_connect", api.ConnectKeyboard)
			serverRouter.POST("/connect_server", api.ConnectServer)
			serverRouter.POST("/remote", api.CheckDeviceOnline)
		}
	}
}
