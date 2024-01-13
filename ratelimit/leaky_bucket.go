package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	producer *time.Ticker
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		producer: time.NewTicker(interval),
	}
}

func (t *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.producer.C: // 如果收到令牌，继续处理请求
			resp, err = handler(ctx, err)
			return
		}
		return
	}
}

func (t *LeakyBucketLimiter) Close() error {
	t.producer.Stop()
	return nil
}
