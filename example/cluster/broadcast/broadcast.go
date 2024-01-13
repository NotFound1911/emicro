package broadcast

import (
	"context"
	"github.com/NotFound1911/emicro/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	registry    registry.Registry // 注册中心实例
	service     string            // 服务名称
	dialOptions []grpc.DialOption // grpc连接配置选项
}

func NewClusterBuilder(r registry.Registry, service string, dialOptions ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry:    r,
		service:     service,
		dialOptions: dialOptions,
	}
}

// BuildUnaryInterceptor 创建gRpc的一元拦截器，用于处理 gRPC 的单向调用
func (c *ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	// method: users.UserService/GetByID
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		if !isBroadCast(ctx) { // 判断是否需要广播
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		instances, err := c.registry.ListServices(ctx, c.service) // 从注册中心获取服务实例列表
		if err != nil {
			return err
		}
		var eg errgroup.Group
		for _, ins := range instances {
			addr := ins.Addr // 获取服务实例的地址
			eg.Go(func() error {
				insCc, er := grpc.Dial(addr, c.dialOptions...) // 建立 gRPC 连接
				if er != nil {
					return er
				}
				return invoker(ctx, method, req, reply, insCc, opts...) // 在每个服务实例上执行请求
			})
		}
		return eg.Wait()
	}
}

type broadcastKey struct {
}

func UseBroadCast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}
func isBroadCast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}
