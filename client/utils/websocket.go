package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"remote/model/common/response"

	logger "github.com/OblivionTime/simple-logger"
	"github.com/gorilla/websocket"
)

func GenerateRandomString() string {
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(randomBytes)[:16]
}

const (
	// 写入时间
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = time.Second * 55

	//ping时间
	pingPeriod = (pongWait * 9) / 10
)

type ServerConn struct {
	WSMutex     sync.Mutex      // WS写锁
	WS          *websocket.Conn // websocket连接
	Exit        chan bool       //退出
	Flag        bool            //判断websocket是否已经关闭
	Receive     chan []byte     //接收客户端的数据
	Send        chan []byte     //发送数据
	MessageType int
}

// 创建连接
func CreateConn(wsConn *websocket.Conn, MessageType int) (*ServerConn, error) {
	//初始化客户端
	return &ServerConn{
		WS:          wsConn,
		Exit:        make(chan bool, 1),
		Flag:        false,
		Send:        make(chan []byte, 2*1024),
		Receive:     make(chan []byte, 2*1024),
		MessageType: MessageType,
	}, nil
}

// 用给定的消息类型（MT）和有效载荷写入消息。
func (*ServerConn) wsWrite(ws *websocket.Conn, mt int, msg interface{}) error {
	var bits []byte
	if msg != nil {
		bits = msg.([]byte)
	} else {
		bits = []byte{}
	}
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.WriteMessage(mt, bits)
}

// 发送消息
func (c *ServerConn) SendMessage(msg []byte) {
	if err := c.wsWrite(c.WS, c.MessageType, msg); err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure) {
			logger.Log.Error("websocket发送消息失败", err)
		}
	}
}

// CloseConn
func (conn *ServerConn) CloseConn() {
	if conn.Flag {
		return
	}
	logger.Log.Info("app websocket close")
	//关闭连接
	conn.WS.Close()
	conn.WSMutex.Lock()
	conn.Flag = true
	conn.Exit <- true
	conn.WSMutex.Unlock()
}
func (conn *ServerConn) WriteReadLoop() {
	conn.WS.SetCloseHandler(func(code int, text string) error {
		conn.CloseConn()
		return nil
	})
	go conn.writeLoop()
	conn.readLoop()
}

// 写处理
func (conn *ServerConn) writeLoop() {
	//定义定时器
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// Break readLoop.
		conn.CloseConn()
	}()
	for {
		select {
		case msg := <-conn.Send:
			//发送消息
			conn.SendMessage(msg)
		case <-ticker.C:
			if err := conn.wsWrite(conn.WS, websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					logger.Log.Error("ws: writeLoop ping", err)
				}
				return
			}
		case flag := <-conn.Exit:
			if flag {
				return
			}
		}
	}
}

// 读处理
func (conn *ServerConn) readLoop() {
	//异常处理
	defer func() {
		conn.CloseConn()
		//异常处理
		RecoverPanic()
	}()

	for {
		_, data, err := conn.WS.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				logger.Log.Error("ws: DoConn:", err)
			} else {
				errinfo, _ := json.Marshal(response.ErrReadMessageRequired(time.Now().Format("2006-01-02 15:04:05")))
				conn.Send <- errinfo
			}
			return
		}
		conn.Receive <- data
	}
}
