package grpc_resolver

import "google.golang.org/grpc/resolver"

var _ resolver.Builder = &Builder{}
var _ resolver.Resolver = &Resolver{}

// Builder 用于构建Resolver，gRpc维护一个scheme->Builder的映射
type Builder struct {
}

func (b Builder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &Resolver{
		cc: cc,
	}
	r.ResolveNow(resolver.ResolveNowOptions{}) // 立即解析获得地址
	return r, nil
}

func (b Builder) Scheme() string {
	return "registry"
}

type Resolver struct {
	cc resolver.ClientConn // 存储客户端连接 服务连接的抽象
}

// ResolveNow 方法用于立即解析gRPC的目标地址
func (r Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	// 固定ip port
	// localhost:8081
	err := r.cc.UpdateState(resolver.State{ // 更新客户端连接的状态
		Addresses: []resolver.Address{
			{
				Addr: "localhost:8081",
			},
		},
	})
	if err != nil {
		r.cc.ReportError(err) // 报告错误到ClientConn
	}
}

func (r Resolver) Close() {
}
