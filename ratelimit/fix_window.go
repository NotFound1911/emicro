package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	timestamp int64 // 窗口起始时间
	interval  int64 // 窗口大小
	rate      int64 // 窗口内允许通过的最大请求数量
	cnt       int64 // 计数
}

func NewFixWindowLimiter(interval time.Duration, rate int64) *FixWindowLimiter {
	return &FixWindowLimiter{
		interval:  interval.Nanoseconds(),
		timestamp: time.Now().UnixNano(),
		rate:      rate,
	}
}

func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 需要考虑重置cnt
		current := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)
		if timestamp+f.interval < current {
			// 新窗口 需要重置
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, current) {
				atomic.CompareAndSwapInt64(&f.cnt, cnt, 0) // 重置cnt
			}
		}
		cnt = atomic.AddInt64(&f.cnt, 1)
		if cnt > f.rate {
			err = errors.New("超出阈值")
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}
