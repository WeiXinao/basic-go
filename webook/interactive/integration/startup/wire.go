//go:build wireinject

package startup

import (
	repository2 "github.com/WeiXinao/basic-go/webook/interactive/repository"
	cache2 "github.com/WeiXinao/basic-go/webook/interactive/repository/cache"
	dao2 "github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	service2 "github.com/WeiXinao/basic-go/webook/interactive/service"
	artEvent "github.com/WeiXinao/basic-go/webook/internal/events/article"
	"github.com/WeiXinao/basic-go/webook/internal/job"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/repository/article"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	artdao "github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	ijwt "github.com/WeiXinao/basic-go/webook/internal/web/jwt"
	"github.com/WeiXinao/basic-go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLog)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGormJobDAO)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	article.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	artdao.NewGORMArticleDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,
		//interactiveSvcSet,

		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCodeRepository,
		artEvent.NewSaramaSyncProducer,

		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitSMSService,
		service.NewCodeService,
		InitPhantomWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		InitWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	// 随便返回一个
	return gin.Default()
}

//func InitArticleHandler() *web.ArticleHandler {
//	wire.Build(thirdProvider,
//		service.NewArticleService,
//		web.NewArticleHandler,
//		article.NewCachedArticleRepository,
//		artdao.NewGORMArticleDAO,
//	)
//	return &web.ArticleHandler{}
//}

func InitArticleHandler(dao artdao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		//interactiveSvcSet,
		article.NewCachedArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		artEvent.NewSaramaSyncProducer,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcSet)
	return service2.NewInteractiveService(nil)
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdProvider, job.NewScheduler)
	return &job.Scheduler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
