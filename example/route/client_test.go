package route

import (
	"context"
	"github.com/NotFound1911/emicro"
	"github.com/NotFound1911/emicro/internal/proto/gen"
	"github.com/NotFound1911/emicro/loadbalance"
	"github.com/NotFound1911/emicro/loadbalance/round_robin"
	"github.com/NotFound1911/emicro/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	client, err := emicro.NewClient(emicro.ClientInsecure(),
		emicro.ClientWithRegistry(r, time.Second*3),
		emicro.ClientWithPickerBuilder("GROUP_ROUND_ROBIN", &round_robin.Builder{
			Filter: loadbalance.GroupFilterBuilder{}.Build(),
		}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ctx = context.WithValue(ctx, "group", "B")

	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)
	uc := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{
			Id: 13,
		})
		require.NoError(t, err)
		t.Log(resp)
	}
}
