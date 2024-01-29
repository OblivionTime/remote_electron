package global

import (
	"net/http"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
)

var (
	OpenWatch bool                = false
	Upgrader  *websocket.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  2 * 1024,
		WriteBufferSize: 2 * 1024,
		// Allow connections from any Origin
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	DB         *gorm.DB
	DeviceList = make(map[string]*swebsocket.ServerConn)
	VideoRooms = make(map[string][]string, 0)
)
