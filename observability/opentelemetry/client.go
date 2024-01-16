package opentelemetry

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// ClientOteBuilder 构建客户端跟踪拦截器
type ClientOteBuilder struct {
	Tracer trace.Tracer // 追踪器
	Port   int
}

// Build 该函数用于拦截客户端的Unary请求，并添加跟踪功能
func (c *ClientOteBuilder) Build() grpc.UnaryServerInterceptor {
	if c.Tracer == nil {
		c.Tracer = otel.GetTracerProvider().Tracer(instrumentationName) // 从OpenTelemetry获取默认的追踪器
	}
	addr := observability.GetOutboundIP()
	if c.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, c.Port)
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 开始一个新的跟踪span，并设置其名称为FullMethod，并标记为客户端span
		spanCtx, span := c.Tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindClient))
		// 设置span的属性，将客户端的地址添加到span中
		span.SetAttributes(attribute.String("address", addr))
		defer func() {
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			span.End()
		}()
		resp, err = handler(spanCtx, req)
		return
	}
}
