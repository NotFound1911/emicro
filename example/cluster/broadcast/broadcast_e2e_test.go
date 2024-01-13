package broadcast

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/NotFound1911/emicro/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestUseBroadCast(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	var servers []*UserServiceServer

	for i := 0; i < 3; i++ {
		//  创建一个服务器实例，使用"user-service"作为服务名称，并使用注册中心
		server, err := emicro.NewServer("user-service", emicro.ServerWithRegistry(r))
		require.NoError(t, err)
		us := &UserServiceServer{
			idx: i,
		}
		servers = append(servers, us)
		// 将新创建的UserServiceServer实例注册到server上
		gen.RegisterUserServiceServer(server, us)
		// 启动 8081,8082, 8083 三个端口
		port := fmt.Sprintf(":808%d", i+1)
		eg.Go(func() error {
			return server.Start(port)
		})
		defer func() {
			_ = server.Close()
		}()
	}
	time.Sleep(time.Second * 3)

	client, err := emicro.NewClient(emicro.ClientInsecure(),
		emicro.ClientWithRegistry(r, time.Second*3))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ctx = UseBroadCast(ctx) // 使用广播
	bd := NewClusterBuilder(r, "user-service", grpc.WithInsecure())
	cc, err := client.Dial(ctx, "user-service", grpc.WithUnaryInterceptor(bd.BuildUnaryInterceptor()))
	require.NoError(t, err)
	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log("resp:", resp)

	for _, s := range servers {
		require.Equal(t, 1, s.cnt)
	}
}

type UserServiceServer struct {
	idx int
	cnt int
	gen.UnimplementedUserServiceServer
}

func (u *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	u.cnt++
	fmt.Println("req:", req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: fmt.Sprintf("test %d", u.idx),
		},
	}, nil
}
