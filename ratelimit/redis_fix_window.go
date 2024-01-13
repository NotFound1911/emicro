package ratelimit

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixWindow string

type RedisFixWindowLimiter struct {
	client   redis.Cmdable
	service  string
	interval time.Duration
	rate     int // 阈值
}

func NewRedisFixWindowLimiter(client redis.Cmdable, service string,
	interval time.Duration, rate int) *RedisFixWindowLimiter {
	return &RedisFixWindowLimiter{
		client:   client,
		service:  service,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisFixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
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

func (r *RedisFixWindowLimiter) limit(ctx context.Context) (bool, error) {
	return r.client.Eval(ctx, luaFixWindow, []string{r.service}, r.interval.Milliseconds(), r.rate).Bool()
}
