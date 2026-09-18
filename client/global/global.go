package global

import (
	"net/http"
	"remote/model"
	"sync"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

var (
	OpenWatch bool = false

	Upgrader *websocket.Upgrader = &websocket.Upgrader{
		ReadBufferSize:  2 * 1024,
		WriteBufferSize: 2 * 1024,
		// Allow connections from any Origin
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	keyMu             sync.RWMutex
	keyDownList       = make(map[string]chan string)
	keyRelseList      = make(map[string]chan string)
	RemoteServerIP    string
	DeviceInfo        = &model.Device{}
	ClientConn        *swebsocket.ServerConn
	VideoConn         *swebsocket.ServerConn
	KeyboardConn      *swebsocket.ServerConn
	Remote_serverConn *swebsocket.ServerConn
	//文件相关
	File_serverConn *swebsocket.ServerConn
	//键盘相关点对点操作
	KeyboardP2PConn *webrtc.PeerConnection
	KeyboardHandler *webrtc.DataChannel
)

// AddKeyboardSession 为一次键盘会话注册按键 channel。
// 钩子线程与转发 goroutine 会并发访问这两张表，统一在这里加锁。
func AddKeyboardSession(user string) {
	keyMu.Lock()
	defer keyMu.Unlock()
	keyDownList[user] = make(chan string, 1024)
	keyRelseList[user] = make(chan string, 1024)
}

func KeyboardChannels(user string) (down chan string, up chan string, ok bool) {
	keyMu.RLock()
	defer keyMu.RUnlock()
	down, okDown := keyDownList[user]
	up, okUp := keyRelseList[user]
	return down, up, okDown && okUp
}
