package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	MaxConnsPerHost: 8,
}

// 请求Client
var client = &http.Client{
	Transport: tr,
	Timeout:   15 * time.Second, // 设置超时时间为 30 秒

}

// 发送请求
func SendMessageServer(RemoteURL string, data []byte) ([]byte, error) {
	var req *http.Request
	var err error
	if data == nil {
		req, err = http.NewRequest("GET", RemoteURL, nil)
	} else {
		req, err = http.NewRequest("POST", RemoteURL, bytes.NewBuffer(data))
	}
	if err != nil {
		fmt.Println("http.GET error: ", err.Error())
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http.GET error: ", err.Error())
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}
