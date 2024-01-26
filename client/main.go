package main

import (
	"net"
	"remote/setup"
	"sync"

	logger "github.com/OblivionTime/simple-logger"
)

var clients chan *net.UDPAddr
var mutex sync.Mutex

func main() {
	logger.InitLog(false, "./logs")
	setup.InitServer(":3002")
}
