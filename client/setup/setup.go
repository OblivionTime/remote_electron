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
	if err := router.Run(host); err != nil {
		logger.Log.Info("Init http server. Error :", err)
	}
}
