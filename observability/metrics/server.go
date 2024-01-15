package metrics

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro/observability"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"time"
)

// ServerMetricsBuilder 服务器度量指标所需的字段
type ServerMetricsBuilder struct {
	Namespace string
	Subsystem string
	Port      int
}

func (s *ServerMetricsBuilder) Build() grpc.UnaryServerInterceptor {
	addr := observability.GetOutboundIP() // 获取出站IP地址
	if s.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, s.Port)
	}
	// 创建活跃请求计数的Prometheus指标
	reqGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数量",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
			// ...
		},
	}, []string{"service"})
	// 注册该指标到Prometheus中
	prometheus.MustRegister(reqGauge)
	// 创建错误计数和响应时长的Prometheus指标
	errCnt := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数量",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
			// ...
		},
	}, []string{"service"})
	// 记录响应的摘要信息（如平均值、标准差等）
	response := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "active_request_cnt",
		Help:      "当前正在处理的请求数量",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   addr,
			// ...
		},
	}, []string{"service"})

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		startTime := time.Now()
		// 增加当前正在处理的请求计数（对于特定FullMethod）
		reqGauge.WithLabelValues(info.FullMethod).Add(1)
		// 当请求处理完成时，执行以下操作：减少当前正在处理的请求计数、增加错误计数（如果有错误）并记录响应时长
		defer func() {
			reqGauge.WithLabelValues(info.FullMethod).Add(-1) // 减少当前正在处理的请求计数
			if err != nil {
				errCnt.WithLabelValues(info.FullMethod).Add(1) // 增加错误计数
			}
			// 记录响应时长
			response.WithLabelValues(info.FullMethod).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		}()
		resp, err = handler(ctx, req)
		return
	}
}
