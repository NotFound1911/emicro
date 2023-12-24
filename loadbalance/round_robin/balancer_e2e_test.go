package round_robin

import (
	"context"
	"fmt"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"net"
	"testing"
	"time"
)

func TestBalance_e2e_Pick(t *testing.T) {
	go func() {
		us := &Server{}
		server := grpc.NewServer()
		gen.RegisterUserServiceServer(server, us)
		l, err := net.Listen("tcp", ":8081")
		require.NoError(t, err)
		err = server.Serve(l)
		t.Log(err)
	}()

	time.Sleep(time.Second * 3)
	balancer.Register(base.NewBalancerBuilder("TEST_ROUND_ROBIN", &Builder{}, base.Config{HealthCheck: true}))
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log("resp:", resp)
}

type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
