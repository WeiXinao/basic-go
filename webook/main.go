package main

import (
	"github.com/WeiXinao/basic-go/webook/config"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	"github.com/WeiXinao/basic-go/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRoutes(server)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，你来了")
	})
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(func(ctx *gin.Context) {
		println("这是第一个 middleware")
	})

	server.Use(func(ctx *gin.Context) {
		println("这是第二个 middleware")
	})

	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 你不加这个，前端是拿不到的
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// 步骤1
	//store := memstore.NewStore(
	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"),
	//)
	//store, err := redis.NewStore(16, "tcp", "192.168.5.33:6379", "",
	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"))
	//if err != nil {
	//	panic(err)
	//}

	//store := memstore.NewStore(
	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"),
	//)

	//myStore := &sqlx_store.Store{}
	//server.Use(sessions.Sessions("mysession", store))

	// 步骤3
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgnorePaths("/users/signup").
	//	IgnorePaths("/users/login").Build())
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
