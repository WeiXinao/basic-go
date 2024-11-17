//go:build wireinject

package main

import (
	events "github.com/WeiXinao/basic-go/webook/interactive/events"
	"github.com/WeiXinao/basic-go/webook/interactive/grpc"
	"github.com/WeiXinao/basic-go/webook/interactive/ioc"
	repository2 "github.com/WeiXinao/basic-go/webook/interactive/repository"
	cache2 "github.com/WeiXinao/basic-go/webook/interactive/repository/cache"
	dao2 "github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	service2 "github.com/WeiXinao/basic-go/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitSaramaClient,
	ioc.InitRedis, ioc.InitDB,
	ioc.InitLogger)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService)

func InitApp() *App {
	wire.Build(thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
