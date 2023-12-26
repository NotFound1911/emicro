package registry

import (
	"context"
	"github.com/NotFound1911/emicro"
	"github.com/NotFound1911/emicro/internal/proto/gen"
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
	// 注册中心创建
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	// 客户端建立
	client, err := emicro.NewClient(emicro.ClientInsecure(), emicro.ClientWithRegistry(r, time.Second*3))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	require.NoError(t, err)
	// 与服务建立连接
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)
	// 使用前面建立的连接创建一个新的客户端实例
	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log("resp:", resp)
}
