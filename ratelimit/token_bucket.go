package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{} // 令牌通道，用于发送令牌
	close  chan struct{}
}

// NewTokenBucketLimiter  创建一个新的令牌桶限流器实例 capacity容量、 interval间隔
func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	producer := time.NewTicker(interval) // 启动一个协程来持续生产令牌
	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C: // 时间间隔到达, 发送令牌
				select {
				case ch <- struct{}{}:
				default: // 通道已满
					// 没有取得令牌
				}
			case <-closeCh: // 关闭
				return
			}
		}
	}()
	return &TokenBucketLimiter{
		tokens: ch,
		close:  closeCh,
	}
}

// BuildServerInterceptor 服务器构建一个拦截器，用于请求限流
func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 这里拿到令牌
		select {
		case <-t.close:
			// 关掉故障检测
			// resp, err := handler(ctx, req)
			err = errors.New("缺乏保护，拒绝请求")
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.tokens:
			resp, err = handler(ctx, req)
		}
		return
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}
