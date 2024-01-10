package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List
	interval int64
	rate     int
	mu       sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		queue:    list.New(),
		interval: interval.Nanoseconds(),
		rate:     rate,
	}
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		current := time.Now().UnixNano()
		boundary := current - s.interval
		// 快路径
		s.mu.Lock()
		length := s.queue.Len()
		if length < s.rate {
			resp, err = handler(ctx, req)
			// 记录请求的时间戳
			s.queue.PushBack(current)
			s.mu.Unlock()
			return
		}
		// 慢路径
		timestamp := s.queue.Front()
		// 删除不再窗口内人数据
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			s.queue.Remove(timestamp)
			timestamp = s.queue.Front()
		}
		length = s.queue.Len()
		s.mu.Unlock()
		if length >= s.rate {
			err = errors.New("达到上限")
			return
		}
		resp, err = handler(ctx, req)
		// 记录当前的时间戳
		s.queue.PushBack(current)
		return
	}
}
