package grpc

import (
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	// 这个是 grpc 的 Server
	server := grpc.NewServer()
	defer func() {
		// 优雅退出
		server.GracefulStop()
	}()
	// 我们的业务的 server
	userServer := Server{}
	RegisterUserServiceServer(server, &userServer)
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	server.Serve(l)
}
