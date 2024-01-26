package main

import (
	"remote_server/setup"

	logger "github.com/OblivionTime/simple-logger"
)

func main() {
	logger.InitLog(false, "./logs")
	setup.InitServer(":9998")
}
