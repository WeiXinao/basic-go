package ioc

import (
	intrv1 "github.com/WeiXinao/basic-go/webook/api/proto/gen/intr/v1"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitIntrClientV1(client *etcdv3.Client) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(etcdResolver)}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	res := intrv1.NewInteractiveServiceClient(cc)
	return res
}

//func InitIntrClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
//	type Config struct {
//		Addr      string `yaml:"addr"`
//		Secure    bool
//		Threshold int32
//	}
//	var cfg Config
//	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
//	if err != nil {
//		panic(err)
//	}
//	var opts []grpc.DialOption
//	if !cfg.Secure {
//		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	}
//
//	cc, err := grpc.Dial(cfg.Addr, opts...)
//	if err != nil {
//		panic(err)
//	}
//	remote := intrv1.NewInteractiveServiceClient(cc)
//	local := client.NewLocalInteractiveServiceAdapter(svc)
//	res := client.NewInteractiveClient(remote, local)
//	viper.OnConfigChange(func(in fsnotify.Event) {
//		cfg = Config{}
//		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
//		if err != nil {
//			panic(err)
//		}
//		res.UpdateThreshold(cfg.Threshold)
//	})
//	return res
//}
