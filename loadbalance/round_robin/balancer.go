package round_robin

import (
	"github.com/NotFound1911/emicro/loadbalance"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index       int32
	connections []subConn
	length      int32
	filter      loadbalance.Filter
}

var _ balancer.Picker = &Balancer{}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := candidates[int(idx)%len(candidates)]
	return balancer.PickResult{
		SubConn: c.c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	Filter loadbalance.Filter
}

var _ base.PickerBuilder = &Builder{}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			addr: ci.Address,
		})
	}
	return &Balancer{
		connections: connections,
		index:       -1,
		length:      int32(len(info.ReadySCs)),
		filter:      b.Filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	addr resolver.Address
}
