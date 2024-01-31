package main

import (
	"fmt"
	"remote_server/config"
	"remote_server/global"
	"remote_server/setup"
	"remote_server/turn"

	logger "github.com/OblivionTime/simple-logger"
)

func main() {
	logger.InitLog(false, "./logs")
	if err := turn.NewServer(); err != nil {
		fmt.Println(err)
		return
	}
	defer global.Turnserver.Svr.Close()
	setup.InitServer(config.Config.ServerAddr)
}
