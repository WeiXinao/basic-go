package ioc

import (
	"context"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	ijwt "github.com/WeiXinao/basic-go/webook/internal/web/jwt"
	"github.com/WeiXinao/basic-go/webook/internal/web/middleware"
	"github.com/WeiXinao/basic-go/webook/pkg/ginx"
	"github.com/WeiXinao/basic-go/webook/pkg/ginx/middlewares/logger"
	"github.com/WeiXinao/basic-go/webook/pkg/ginx/middlewares/prometheus"
	logger2 "github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	fmt.Println("InitWebServer run")
	return server
}

func InitMiddlewares(redisClient redis.Cmdable,
	l logger2.LoggerV1,
	jwtHdl ijwt.Handler) []gin.HandlerFunc {
	bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
		l.Debug("HTTP请求", logger2.Field{Key: "al", Value: al})
	}).AllowReqBody(true).AllowRespBody()
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})
	pb := &prometheus.Builder{
		Namespace: "xiaoxin",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的 HTTP 接口数据",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "xiaoxin",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})

	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/signup").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/login").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 你不加这个，前端是拿不到的
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
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
	})
}
