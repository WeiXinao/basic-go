package ioc

import (
	grpcv3 "github.com/WeiXinao/basic-go/webook/account/grpc"
	"github.com/WeiXinao/basic-go/webook/pkg/grpcx"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGRPCxServer(asc *grpcv3.AccountServiceServer,
	acli *clientv3.Client,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port    int   `yaml:"port"`
		EtcdTTL int64 `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	asc.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "reward",
		L:          l,
		EtcdClient: acli,
		EtcdTTL:    cfg.EtcdTTL,
	}
}
