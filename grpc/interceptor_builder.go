package grpc

import (
	"context"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
}

func NewInterceptorBuilder(limiter ratelimit.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{
		limiter: limiter,
		key:     key,
	}
}

func (b InterceptorBuilder) BuildServerUnaryInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if getReq, ok := req.(*GetByIdRequest); ok {
			key := fmt.Sprintf("limiter:user:get_by_id:%d", getReq.Id)
			limited, err := b.limiter.Limit(ctx, key)
			if err != nil {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
			if limited {
				return nil, status.Errorf(codes.Unimplemented, "限流")
			}
		}
		return handler(ctx, req)
	}
}
