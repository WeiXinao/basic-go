package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default()
	// 如果一个 HTTP 请求，用 GET 方法访问的时候，如果访问路径是 /hello,
	server.GET("/hello", func(c *gin.Context) {
		// 就执行这段代码
		c.JSON(http.StatusOK, "hello, go")
	})

	server.POST("/post", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, post 方法")
	})

	// /users/delete?name=xiaoxin
	// get /users/xiaoxin 查询
	// delete /users/xiaxiao 删除
	// put /users/xiaoxin 注册
	// post /users/xiaoxin 修改
	server.GET("/users/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.String(http.StatusOK, "hello, 这是参数路由"+name)
	})

	server.GET("/views/*.html", func(ctx *gin.Context) {
		page := ctx.Param(".html")
		ctx.String(http.StatusOK, "hello, 这是通配符路由"+page)
	})

	server.GET("/order", func(ctx *gin.Context) {
		oid := ctx.Query("id")
		ctx.String(http.StatusOK, "hello, 这是查询参数"+oid)
	})

	//server.GET("/items/", func(ctx *gin.Context) {
	//	ctx.String(http.StatusOK, "hello, 这是 items")
	//})

	server.GET("/items/*abc", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, 这是 items")
	})

	//server.GET("/users/*/", func(ctx *gin.Context) {
	//
	//})

	server.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}
