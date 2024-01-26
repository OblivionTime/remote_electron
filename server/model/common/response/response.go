package response

import (
	"encoding/json"
	"net/http"
	"time"

	"gitee.com/solidone/sutils/sencryption"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
	Msg       string      `json:"msg"`
	Timestamp string      `json:"ts"`
}

const (
	ERROR   = 7
	SUCCESS = 0
)

var X = string([]byte{224, 130, 44, 118, 90, 64, 194, 254, 132, 203, 16, 162, 60, 98, 138, 210, 133, 127, 81, 251, 99, 33, 185, 109, 228, 56, 236, 246, 188, 40, 236, 45})
var Y = string([]byte{35, 94, 180, 165, 65, 11, 166, 41, 215, 187, 165, 69, 202, 219, 158, 136})

func Result(code int, data interface{}, msg string, c *gin.Context) {
	// 开始时间
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
		time.Now().Format("2006-01-02 15:04:05"),
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	res, err := json.Marshal(data)
	if err != nil {
		Fail(c)
		return
	}
	data, _ = sencryption.AesEncrypt(res, X, Y)
	Result(SUCCESS, data, "查询成功", c)
}

func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}
func FailWithAuthority(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, nil)
}
