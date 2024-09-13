package web

import (
	"github.com/WeiXinao/basic-go/webook/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type handler interface {
	RegisterRoutes(server *gin.Engine)
}

// 重构小技巧
type Result = ginx.Result

//type Result struct {
//	// 这个叫做业务错误码
//	Code int    `json:"code"`
//	Msg  string `json:"msg"`
//	Data any    `json:"data"`
//}
