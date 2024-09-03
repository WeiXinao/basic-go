// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	"github.com/WeiXinao/basic-go/webook/internal/web/jwt"
	"github.com/WeiXinao/basic-go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	limiter := ioc.NewRateLimiter(cmdable)
	loggerV1 := InitLog()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitMiddlewares(cmdable, limiter, loggerV1, handler)
	gormDB := InitTestDB()
	userDao := dao.NewUserDAO(gormDB)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCode(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	feishuService := InitPhantomWechatService(loggerV1)
	feishuHandlerConfig := ioc.NewFeishuHandlerConfig()
	oAuth2FeishuHandler := web.NewOAuth2FeishuHandler(feishuService, userService, handler, feishuHandlerConfig)
	engine := ioc.InitGin(v, userHandler, oAuth2FeishuHandler)
	return engine
}

func InitUserSvc() service.UserService {
	gormDB := InitTestDB()
	userDao := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	loggerV1 := InitLog()
	userService := service.NewUserService(userRepository, loggerV1)
	return userService
}

func InitJwtHdl() jwt.Handler {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	return handler
}

// wire.go:

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewUserCache, repository.NewUserRepository, service.NewUserService)
