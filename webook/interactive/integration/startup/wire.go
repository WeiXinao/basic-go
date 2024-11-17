//go:build wireinject

package startup

import (
	"github.com/WeiXinao/basic-go/webook/interactive/grpc"
	repository2 "github.com/WeiXinao/basic-go/webook/interactive/repository"
	cache2 "github.com/WeiXinao/basic-go/webook/interactive/repository/cache"
	dao2 "github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	service2 "github.com/WeiXinao/basic-go/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	InitRedis, InitDB,
	//InitSaramaClient,
	//InitSyncProducer,
	InitLog)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
