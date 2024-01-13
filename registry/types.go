package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) (<-chan Event, error)

	io.Closer
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	Name string
	// Addr 定位信息
	Addr string
	// Weight 权重
	Weight uint32
	// Group 分组
	Group string
}

type Event struct {
	Type string
}
