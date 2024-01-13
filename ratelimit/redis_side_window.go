package ratelimit

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

var luaSlideWindow string

type RedisSlideWindowLimiter struct {
	client   redis.Cmdable
	service  string
	interval time.Duration
	rate     int
}

func NewRedisSlideWindowLimiter(client redis.Cmdable, service string,
	interval time.Duration, rate int) *RedisSlideWindowLimiter {
	return &RedisSlideWindowLimiter{
		client:   client,
		service:  service,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 使用 FullMethod，那就是单一方法上限流，比如说 GetById
		// 使用服务名来限流，那就是在单一服务上 users.UserService
		// 使用应用名，user-service
		limit, err := r.limit(ctx)
		if err != nil {
			return
		}
		if limit {
			err = errors.New("超出阈值")
			return
		}
		resp, err = handler(ctx, err)
		return
	}
}

func (r *RedisSlideWindowLimiter) limit(ctx context.Context) (bool, error) {
	return r.client.Eval(ctx, luaFixWindow, []string{r.service}, r.interval.Milliseconds(), r.rate).Bool()
}
