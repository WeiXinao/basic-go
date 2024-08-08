package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	//db := initDB()
	//redisClient := initRedis()
	//
	//server := initWebServer()
	//
	//u := initUser(db, redisClient)
	//u.RegisterRoutes(server)

	server := InitWebServer()

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，你来了")
	})
	server.Run(":8080")
}

//func initWebServer() *gin.Engine {
//	server := gin.Default()
//
//	server.Use(func(ctx *gin.Context) {
//		println("这是第一个 middleware")
//	})
//
//	server.Use(func(ctx *gin.Context) {
//		println("这是第二个 middleware")
//	})
//
//	//redisClient := redis.NewClient(&redis.Options{
//	//	Addr: config.Config.Redis.Addr,
//	//})
//	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//
//	server.Use(cors.New(cors.Config{
//		//AllowOrigins: []string{"*"},
//		//AllowMethods: []string{"POST", "GET"},
//		AllowHeaders: []string{"Content-Type", "Authorization"},
//		// 你不加这个，前端是拿不到的
//		ExposeHeaders: []string{"x-jwt-token"},
//		// 是否允许你带 cookie 之类的东西
//		AllowCredentials: true,
//		AllowOriginFunc: func(origin string) bool {
//			if strings.HasPrefix(origin, "http://localhost") {
//				// 你的开发环境
//				return true
//			}
//			return strings.Contains(origin, "yourcompany.com")
//		},
//		MaxAge: 12 * time.Hour,
//	}))
//
//	// 步骤1
//	//store := memstore.NewStore(
//	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
//	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"),
//	//)
//	//store, err := redis.NewStore(16, "tcp", "192.168.5.33:6379", "",
//	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
//	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"))
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	//store := memstore.NewStore(
//	//	[]byte("xTEsVjQ6LCdeUkWESjxhWIV2VnsDj5Pq"),
//	//	[]byte("hJLz5UFLxBXfoQ0C0ovf9FqypjpDjnDK"),
//	//)
//
//	//myStore := &sqlx_store.Store{}
//	//server.Use(sessions.Sessions("mysession", store))
//
//	// 步骤3
//	//server.Use(middleware.NewLoginMiddlewareBuilder().
//	//	IgnorePaths("/users/signup").
//	//	IgnorePaths("/users/login").Build())
//
//	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
//		IgnorePaths("/users/signup").
//		IgnorePaths("/users/login").
//		IgnorePaths("/users/login_sms/code/send").
//		IgnorePaths("/users/login_sms").Build())
//	return server
//}

//func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
//	ud := dao.NewUserDAO(db)
//	uc := cache.NewUserCache(rdb)
//	repo := repository.NewUserRepository(ud, uc)
//	svc := service.NewUserService(repo)
//	codeCache := cache.NewCode(rdb)
//	codeRepo := repository.NewCodeRepository(codeCache)
//	smsSvc := memory.NewService()
//	codeSvc := service.NewCodeService(codeRepo, smsSvc)
//	u := web.NewUserHandler(svc, codeSvc)
//	return u
//}
