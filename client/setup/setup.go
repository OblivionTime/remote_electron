/*
 * @Description:
 * @Version: 2.0
 * @Autor: solid
 * @Date: 2021-11-28 11:37:41 +0800
 * @LastEditors: solid
 * @LastEditTime: 2022-08-11 12:17:32
 */
package setup

import (
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
)

func InitServer(host string) {
	// router := gin.New()
	router := gin.Default()
	//设置路由
	setupRouter(router)
	err := router.Run(host)
	if err != nil {
		logger.Log.Infof("Init http server. Error :", err)
	}
}
