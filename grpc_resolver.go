package emicro

import (
	"context"
	"github.com/NotFound1911/emicro/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

var _ resolver.Builder = &grpcBuilder{}
var _ resolver.Resolver = &grpcResolver{}

// grpcBuilder 用于构建Resolver，gRpc维护一个scheme->Builder的映射
type grpcBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*grpcBuilder, error) {
	return &grpcBuilder{r: r, timeout: timeout}, nil
}
func (b *grpcBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}
	r.resolve()
	go r.watch()
	return r, nil
}

func (b *grpcBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	target  resolver.Target
	r       registry.Registry
	cc      resolver.ClientConn // 存储客户端连接 服务连接的抽象
	timeout time.Duration
	close   chan struct{}
}

// ResolveNow 方法用于立即解析gRPC的目标地址
func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
	}
	for {
		select {
		case <-events:
			g.resolve()

		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{
			Addr:       si.Addr,
			ServerName: si.Name,
			Attributes: attributes.New("weight", si.Weight).
				WithValue("group", si.Group),
		})
	}
	err = g.cc.UpdateState(resolver.State{Addresses: address})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}
