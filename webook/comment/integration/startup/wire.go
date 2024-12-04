//go:build wireinject

package startup

import (
	"github.com/WeiXinao/basic-go/webook/comment/grpc"
	"github.com/WeiXinao/basic-go/webook/comment/ioc"
	"github.com/WeiXinao/basic-go/webook/comment/repository"
	"github.com/WeiXinao/basic-go/webook/comment/repository/dao"
	"github.com/WeiXinao/basic-go/webook/comment/service"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	InitTestDB,
)

func InitGRPCServer() *grpc.CommentServiceServer {
	wire.Build(thirdProvider, serviceProviderSet)
	return new(grpc.CommentServiceServer)
}
