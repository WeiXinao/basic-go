//go:build wireinject

package main

import (
	"github.com/WeiXinao/basic-go/webook/internal/events/article"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	artRepo "github.com/WeiXinao/basic-go/webook/internal/repository/article"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	artDao "github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	"github.com/WeiXinao/basic-go/webook/internal/web"
	ijwt "github.com/WeiXinao/basic-go/webook/internal/web/jwt"
	"github.com/WeiXinao/basic-go/webook/ioc"
	"github.com/google/wire"
)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepositoryV1,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,

		// 初始化 DAO
		dao.NewUserDAO,
		artDao.NewGORMArticleDAO,

		interactiveSvcSet,
		rankingSvcSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		//	cache 部分
		cache.NewUserCache,
		cache.NewCodeCache,
		//cache.NewMemcachedCodeCache,
		cache.NewArticleRedisCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		artRepo.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
