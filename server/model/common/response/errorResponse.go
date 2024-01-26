/*
 * @Description:错误返回
 * @Version: 1.0
 * @Autor: solid
 * @Date: 2021-12-23 11:31:14
 * @LastEditors: solid
 * @LastEditTime: 2021-12-27 15:07:23
 */
package response

import (
	"net/http"
	"time"
)

func ErrAPIKeyRequired(ts string) *Response {
	return &Response{
		Code:      http.StatusForbidden,
		Msg:       "valid API key required",
		Timestamp: ts,
	}
}

//读数据时出现问题
func ErrReadMessageRequired(ts string) *Response {
	return &Response{
		Code:      4004,
		Msg:       "读取错误",
		Timestamp: ts,
	}
}

//参数解析时出现问题
func ErrParameterUnmarshal(ts string) *Response {
	return &Response{
		Code:      4004,
		Msg:       "参数有误",
		Timestamp: ts,
	}
}
func ErrRequestUSBExit(msg string) *Response {
	ts := time.Now().Format("2006-01-02 15:04:05")
	return &Response{
		Code:      4008,
		Msg:       msg,
		Timestamp: ts,
	}
}

//error
func ErrRequestClient(msg string) *Response {
	ts := time.Now().Format("2006-01-02 15:04:05")
	return &Response{
		Code:      4003,
		Msg:       msg,
		Timestamp: ts,
	}
}
