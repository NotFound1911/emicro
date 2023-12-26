package registry

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/NotFound1911/emicro/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	us := UserServiceServer{}
	server, err := emicro.NewServer("user-service", emicro.ServerWithRegistry(r))
	// 服务注册
	gen.RegisterUserServiceServer(server, us)

	err = server.Start(":8081")
	t.Log(err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (u UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println("req:", req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "test",
		},
	}, nil
}
