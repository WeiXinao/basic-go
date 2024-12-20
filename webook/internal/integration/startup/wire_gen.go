// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	article3 "github.com/WeiXinao/basic-go/webook/internal/events/article"
	"github.com/WeiXinao/basic-go/webook/internal/job"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	article2 "github.com/WeiXinao/basic-go/webook/internal/repository/article"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
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
	loggerV1 := InitLog()
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitMiddlewares(cmdable, loggerV1, handler)
	gormDB := InitDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler, loggerV1, cmdable)
	wechatService := InitPhantomWechatService(loggerV1)
	wechatHandlerConfig := InitWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler, wechatHandlerConfig)
	articleDAO := article.NewGORMArticleDAO(gormDB)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := article2.NewCachedArticleRepository(articleDAO, userRepository, articleCache)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article3.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler(dao2 article.ArticleDAO) *web.ArticleHandler {
	gormDB := InitDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := article2.NewCachedArticleRepository(dao2, userRepository, articleCache)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article3.NewSaramaSyncProducer(syncProducer)
	loggerV1 := InitLog()
	articleService := service.NewArticleService(articleRepository, producer, loggerV1)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	return articleHandler
}

func InitInteractiveService() service.InteractiveService {
	gormDB := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(gormDB)
	cmdable := InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	return interactiveService
}

func InitJobScheduler() *job.Scheduler {
	gormDB := InitDB()
	jobDAO := dao.NewGormJobDAO(gormDB)
	cronJobRepository := repository.NewPreemptJobRepository(jobDAO)
	loggerV1 := InitLog()
	cronJobService := service.NewCronJobService(cronJobRepository, loggerV1)
	scheduler := job.NewScheduler(cronJobService, loggerV1)
	return scheduler
}

func InitUserSvc() service.UserService {
	gormDB := InitDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
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

var thirdProvider = wire.NewSet(
	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLog)

var jobProviderSet = wire.NewSet(service.NewCronJobService, repository.NewPreemptJobRepository, dao.NewGormJobDAO)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewUserCache, repository.NewUserRepository, service.NewUserService)

var articleSvcProvider = wire.NewSet(article2.NewCachedArticleRepository, cache.NewArticleRedisCache, article.NewGORMArticleDAO, service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository, service.NewInteractiveService)
