package startup

import (
	pmtv1 "github.com/WeiXinao/basic-go/webook/api/proto/gen/payment/v1"
	"github.com/WeiXinao/basic-go/webook/reward/grpc"
	"github.com/WeiXinao/basic-go/webook/reward/repository"
	"github.com/WeiXinao/basic-go/webook/reward/repository/cache"
	"github.com/WeiXinao/basic-go/webook/reward/repository/dao"
	"github.com/WeiXinao/basic-go/webook/reward/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) *grpc.RewardServiceServer {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		cache.NewRewardRedisCache,
		repository.NewRepository, dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer)
	return new(grpc.RewardServiceServer)
}
