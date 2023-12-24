package grpc

import (
	"context"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

// protoc --go_out=. --go-grpc_out=. user.proto
func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
}
