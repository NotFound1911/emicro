package emicro

import (
	"context"
	"github.com/NotFound1911/emicro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	Name            string
	registry        registry.Registry
	registryTimeout time.Duration
	*grpc.Server
	listener net.Listener
	group    string
	weight   int
}

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		Name:            name,
		Server:          grpc.NewServer(),
		registryTimeout: time.Second * 10,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name:  s.Name,
			Addr:  listener.Addr().String(),
			Group: s.group,
		})
		if err != nil {
			return err
		}
	}
	err = s.Serve(listener)
	return err
}

func (s *Server) Close() error {
	// 需要先关registry
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}
func ServerWithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}
func ServerWithGroup(group string) ServerOption {
	return func(server *Server) {
		server.group = group
	}
}

func ServerWithWeight(weight int) ServerOption {
	return func(server *Server) {
		server.weight = weight
	}
}
