package grpc_resolver

import (
	"context"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// grpc.WithResolvers(&Builder{})指定了一个解析器，这里使用了前面定义的Builder结构体
	cc, err := grpc.Dial("registry://localhost:8081", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
}
