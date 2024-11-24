package grpc

import (
	"context"
	etcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/random"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

type KratosTestSuite struct {
	suite.Suite
	etcdClient *etcdv3.Client
}

func (s *KratosTestSuite) SetupSuite() {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"192.168.5.3:2379"},
	})
	require.NoError(s.T(), err)
	s.etcdClient = cli
}

func (s *KratosTestSuite) TestClient() {
	// 默认是 WRR 负载均衡算法
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	for _ = range 10 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *KratosTestSuite) TestClientLoadBalancer() {
	selector.SetGlobalSelector(random.NewBuilder())
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	for _ = range 10 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *KratosTestSuite) TestServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

func (s *KratosTestSuite) startServer(addr string) {
	grpcSrv := grpc.NewServer(
		grpc.Address(addr),
		grpc.Middleware(recovery.Recovery()),
	)
	RegisterUserServiceServer(grpcSrv, &Server{
		Name: addr,
	})
	//	注册中心
	r := etcd.New(s.etcdClient)
	app := kratos.New(
		kratos.Name("user"),
		kratos.Server(
			grpcSrv,
		),
		kratos.Registrar(r),
	)
	app.Run()
}

func TestKratos(t *testing.T) {
	suite.Run(t, new(KratosTestSuite))
}
