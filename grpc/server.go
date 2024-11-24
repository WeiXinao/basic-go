package grpc

import (
	"context"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &Server{}

func (s *Server) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	return &GetByIdResponse{
		User: &User{
			Id:   123,
			Name: "from" + s.Name,
		},
	}, nil
}
