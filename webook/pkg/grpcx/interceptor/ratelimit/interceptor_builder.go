package ratelimit

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
}

// NewInterceptorBuilder key1: limiter:interceptor-service => 整个点赞的应用限流
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{
		limiter: limiter,
		key:     key,
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil || limited {
			ctx = context.WithValue(ctx, "downgrade", "true")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			//	这个 key，limiter:UserService
			limited, err := b.limiter.Limit(ctx, b.key)
			if err != nil {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}

			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)
	}
}
