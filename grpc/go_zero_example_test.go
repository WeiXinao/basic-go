package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type GoZeroTestSuite struct {
	suite.Suite
}

func (s *GoZeroTestSuite) TestGoZeroClient() {
	zClient := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"192.168.5.3:2379"},
			Key:   "user",
		},
	},
		zrpc.WithDialOption(
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		))
	client := NewUserServiceClient(zClient.Conn())
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

// TestGoZeroServer 启动 grpc 服务器
func (s *GoZeroTestSuite) TestGoZeroServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

// TestGoZeroServer 启动 grpc 服务端
func (s *GoZeroTestSuite) startServer(addr string) {
	c := zrpc.RpcServerConf{
		ListenOn: addr,
		Etcd: discov.EtcdConf{
			Hosts: []string{"192.168.5.3:2379"},
			Key:   "user",
		},
	}

	//	创建的一个服务器，并且注册服务实例
	server := zrpc.MustNewServer(c, func(server *grpc.Server) {
		RegisterUserServiceServer(server, &Server{
			Name: addr,
		})
	})

	//	这个是往 gRPC 里面添加拦截器（也可以叫做插件）
	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
