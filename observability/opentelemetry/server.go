package opentelemetry

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const instrumentationName = "github.com/NotFound1911/emicro/observability/opentelemetry"

// ServerOtelBuilder 用于构建服务器端跟踪拦截器
type ServerOtelBuilder struct {
	Tracer trace.Tracer
	Port   int
}

// Build 该函数用于拦截服务器端的Unary请求，并添加跟踪功能
func (s *ServerOtelBuilder) Build() grpc.UnaryServerInterceptor {
	if s.Tracer == nil {
		s.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	addr := observability.GetOutboundIP()
	if s.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, s.Port)
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		ctx = s.extract(ctx) // 提取context中的元数据（如果有）
		spanCtx, span := s.Tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(attribute.String("address", addr))
		defer func() {
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			span.End()
		}()
		resp, err = handler(spanCtx, err)
		return
	}
}

// extract 从给定的context中提取元数据，并使用OpenTelemetry的文本映射传播器将它们传递给新的context
func (s *ServerOtelBuilder) extract(ctx context.Context) context.Context {
	// 从上下文中获取元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	// 使用OpenTelemetry的文本映射传播器从元数据中提取跟踪信息，并返回新的context
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(md))
}
