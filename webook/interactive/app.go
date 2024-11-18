package main

import (
	"github.com/WeiXinao/basic-go/webook/internal/events"
	"github.com/WeiXinao/basic-go/webook/pkg/ginx"
	"github.com/WeiXinao/basic-go/webook/pkg/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
}
