package fastest

import (
	"context"
	"github.com/NotFound1911/emicro/registry"
	"google.golang.org/grpc"
	"reflect"
	"sync"
)

type ClusterBuilder struct {
	registry    registry.Registry
	service     string
	dialOptions []grpc.DialOption
}

func NewClusterBuilder(r registry.Registry, service string, dialOptions ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry:    r,
		service:     service,
		dialOptions: dialOptions,
	}
}

func (c *ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	// method: users.UserService/GetByID
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {
		ok, ch := isBroadCast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		defer func() {
			close(ch)
		}()

		instances, err := c.registry.ListServices(ctx, c.service)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		typ := reflect.TypeOf(reply).Elem()
		wg.Add(len(instances))
		for _, ins := range instances {
			addr := ins.Addr
			go func() {
				insCc, er := grpc.Dial(addr, c.dialOptions...)
				if er != nil {
					ch <- Resp{Err: er}
					wg.Done()
					return
				}
				newReply := reflect.New(typ).Interface()
				err = invoker(ctx, method, req, newReply, insCc, opts...)
				// 如果没有接收 会被阻塞
				select {
				case ch <- Resp{Err: er, Reply: newReply}: // 最快响应
				default:

				}
				wg.Done()
			}()
		}
		wg.Wait()
		return err
	}
}

type broadcastKey struct {
}

func UseBroadCast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}
func isBroadCast(ctx context.Context) (bool, chan Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Err   error
	Reply any
}
