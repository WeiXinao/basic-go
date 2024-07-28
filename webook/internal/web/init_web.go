package web

import "github.com/gin-gonic/gin"

func RegisterRoutes() *gin.Engine {
	server := gin.Default()
	registerUsersRoutes(server)
	return server
}

func registerUsersRoutes(server *gin.Engine) {
	u := &UserHandler{}

	// 这是 REST 风格
	//server.PUT("/user", func(ctx *gin.Context) {
	//
	//})
	server.POST("/users/signup", u.SignUp)

	server.POST("/users/login", u.Login)

	// REST 风格
	//server.POST("/users/:id", func(ctx *gin.Context) {
	//
	//})
	server.POST("/users/edit", u.Edit)

	// REST 风格
	//server.GET("/users/:id", func(ctx *gin.Context) {
	//
	//})
	server.GET("/users/profile", u.Profile)
}
