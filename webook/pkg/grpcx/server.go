package grpcx

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/WeiXinao/basic-go/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	EtcdAddr string
	Port     int
	Name     string
	L        logger.LoggerV1

	client   *etcdv3.Client
	kaCancel func()
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 我们要在这里完成注册
	s.register()
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = client
	target := "service/" + s.Name
	em, err := endpoints.NewManager(client, target)
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//	租期
	var ttl int64 = 30
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		//	定位信息，客户端怎么连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return err
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.client != nil {
		//	依赖注入，你就不要关
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
