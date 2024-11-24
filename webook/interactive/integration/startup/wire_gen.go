// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/WeiXinao/basic-go/webook/interactive/grpc"
	"github.com/WeiXinao/basic-go/webook/interactive/repository"
	"github.com/WeiXinao/basic-go/webook/interactive/repository/cache"
	"github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	"github.com/WeiXinao/basic-go/webook/interactive/service"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitInteractiveService() *grpc.InteractiveServiceServer {
	db := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	cmdable := InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	return interactiveServiceServer
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitRedis, InitDB,

	InitLog)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository, service.NewInteractiveService)