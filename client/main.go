package main

import (
	"remote/setup"

	logger "github.com/OblivionTime/simple-logger"
)

func main() {
	logger.InitLog(false, "./logs")
	setup.InitServer(":3002")
}
