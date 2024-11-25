package grpc

import (
	"context"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &Server{}

func (s *Server) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		log.Println(rest.String())
	}
	time.Sleep(time.Millisecond * 100)
	return &GetByIdResponse{
		User: &User{
			Id:   123,
			Name: "from " + s.Name,
		},
	}, nil
}
