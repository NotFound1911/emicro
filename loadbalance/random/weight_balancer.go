package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type WeightBalancer struct {
	connections []*weightConn
	totalWeight uint32
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(w.totalWeight) + 1)
	var idx int
	for i, c := range w.connections {
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
		})
	}
	return &WeightBalancer{
		connections: cs,
		totalWeight: totalWeight,
	}
}

type weightConn struct {
	c      balancer.SubConn
	weight uint32
}
