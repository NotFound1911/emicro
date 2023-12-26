package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	index       int32              // 当前连接的索引
	connections []balancer.SubConn // 存储所有子连接
	length      int32
}

// Picker 用于gRpc实现负载均衡算法的接口
var _ balancer.Picker = &Balancer{}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 { // 无可用连接
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1) // 轮询加1
	c := b.connections[idx%b.length]    // 获取连接
	return balancer.PickResult{
		SubConn: c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
}

// PickerBuilder创建 balancer.Picker
var _ base.PickerBuilder = &Builder{}

func (b Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs { // 遍历已准备好的子连接列表
		connections = append(connections, c)
	}
	return &Balancer{ // 返回一个新创建的Balancer实例
		connections: connections,               // 所有已准备好的子连接
		index:       -1,                        // 开始索引
		length:      int32(len(info.ReadySCs)), // 子连接的数量
	}
}
