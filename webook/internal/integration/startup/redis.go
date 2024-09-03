package startup

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

var redisClient redis.Cmdable

func InitRedis() redis.Cmdable {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "192.168.5.3:6379",
		})

		for err := redisClient.Ping(context.Background()).Err(); err != nil; {
			panic(err)
		}
	}
	return redisClient
}

func NewRateLimiter(redisClient redis.Cmdable) ratelimit.Limiter {
	return ratelimit.NewRedisSlidingWindowLimiter(
		redisClient, time.Second, 100,
	)
}
