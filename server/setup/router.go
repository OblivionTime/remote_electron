/*
 * @Author: your name
 * @Date: 2021-11-22 09:09:16
 * @LastEditTime: 2022-08-11 11:45:01
 * @LastEditors: solid
 * @Description: 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 * @FilePath: \Yun_Music_Back\setup\router.go
 */
package setup

import (
	"net/http"
	"remote_server/api"
	"remote_server/global"

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
	global.DB = GormSqlite()
	/*
		用户
	*/
	remoteRouter := router.Group("/v1/api/remote")
	{
		remoteRouter.GET("/code", api.GetCode)
		remoteRouter.POST("/online", api.DeviceOnlineStatus)
		remoteRouter.GET("/connect", api.Connect)
		remoteRouter.GET("/device_list", api.GetConnectDeviceList)
		remoteRouter.GET("/file_connect", api.FileConnect)
	}
}
