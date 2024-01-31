package model

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"sync"

	"github.com/pion/turn/v3"
)

// turnserver服务
type TurnServer struct {
	lock   sync.RWMutex
	lookup map[string]*Entry
	Svr    *turn.Server
}
type Entry struct {
	addr     string
	password []byte
}

func (e *Entry) SetIP(addr string) {
	e.addr = addr
}
func NewTurnServer() *TurnServer {
	return &TurnServer{
		lookup: make(map[string]*Entry),
		Svr:    nil,
	}
}
func (a *TurnServer) allow(username, password string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.lookup[username] = &Entry{
		addr:     "",
		password: turn.GenerateAuthKey(username, "oblivionTime", password),
	}
}
func (a *TurnServer) Disallow(username string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.lookup, username)
}
func (a *TurnServer) Credentials(username string) (string, string) {
	password := RandString(20)
	a.allow(username, password)
	return username, password
}
func (a *TurnServer) Authenticate(username, realm string, addr net.Addr) ([]byte, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	entry, ok := a.lookup[username]
	fmt.Println(addr.String(), "连接,信息为:", entry, "状态为:", ok)
	if !ok {
		return nil, false
	}
	if entry.addr == "" {
		a.lookup[username].SetIP(addr.String())
	}
	return entry.password, true
}

var (
	tokenCharacters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_!@#$%^&*()){}\\/=+,.><")
	randReader      = rand.Reader
)

func RandString(length int) string {
	res := make([]byte, length)
	for i := range res {
		index := randIntn(len(tokenCharacters))
		res[i] = tokenCharacters[index]
	}
	return string(res)
}
func randIntn(n int) int {
	max := big.NewInt(int64(n))
	res, err := rand.Int(randReader, max)
	if err != nil {
		panic("random source is not available")
	}
	return int(res.Int64())
}
