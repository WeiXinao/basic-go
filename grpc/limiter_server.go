package grpc

import (
	"context"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/pkg/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LimiterUserServer struct {
	limiter ratelimit.Limiter
	UserServiceServer
}

func (s *LimiterUserServer) GetById(ctx context.Context, req *GetByIdRequest) (*GetByIdResponse, error) {
	key := fmt.Sprintf("limiter:user:get_by_id:%d", req.Id)
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
	if limited {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
	return s.UserServiceServer.GetById(ctx, req)
}
