package response

//修改成功
func UpdateRequestClient(ts, msg string) *Response {
	return &Response{
		Code:      0,
		Msg:       msg,
		Timestamp: ts,
	}
}
