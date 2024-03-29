package loadbalance

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// Filter 过滤，是否保留
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilterBuilder struct{}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		target, _ := addr.Attributes.Value("group").(string)
		input, _ := info.Ctx.Value("group").(string)
		return target == input
	}
}
