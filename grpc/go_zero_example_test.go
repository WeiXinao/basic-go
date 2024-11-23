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
	})
	client := NewUserServiceClient(zClient.Conn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

// TestGoZeroServer 启动 grpc 服务端
func (s *GoZeroTestSuite) TestGoZeroServer() {
	c := zrpc.RpcServerConf{
		ListenOn: ":8090",
		Etcd: discov.EtcdConf{
			Hosts: []string{"192.168.5.3:2379"},
			Key:   "user",
		},
	}

	//	创建的一个服务器，并且注册服务实例
	server := zrpc.MustNewServer(c, func(server *grpc.Server) {
		RegisterUserServiceServer(server, &Server{})
	})

	//	这个是往 gRPC 里面添加拦截器（也可以叫做插件）
	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
