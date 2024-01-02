package random

import (
	"github.com/NotFound1911/emicro/loadbalance"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

type WeightBalancer struct {
	connections []*weightConn
	filter      loadbalance.Filter
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var totalWeight uint32
	candidates := make([]*weightConn, 0, len(w.connections))
	for _, c := range w.connections {
		if w.filter != nil && !w.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
		totalWeight = totalWeight + c.weight
	}
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(totalWeight) + 1)
	var idx int
	for i, c := range candidates {
		tgt = tgt - int(c.weight)
		if tgt <= 0 {
			idx = i
			break
		}
	}
	return balancer.PickResult{
		SubConn: w.connections[idx].c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type WeightBalancerBuilder struct {
	Filter loadbalance.Filter
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	var totalWeight uint32
	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		totalWeight += weight
		cs = append(cs, &weightConn{
			c:      sub,
			weight: weight,
			addr:   subInfo.Address,
		})
	}
	return &WeightBalancer{
		connections: cs,
		filter:      w.Filter,
	}
}

type weightConn struct {
	c      balancer.SubConn
	weight uint32
	addr   resolver.Address
}
