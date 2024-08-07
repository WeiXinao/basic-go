package web

// 501001 => 这里代表错误验证码
type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}
