package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// cc 是一个连接池的池子，就是 cc 里面放了很多个连接池，一个 IP + 端口 一个连接池
	cc, err := grpc.Dial(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 456,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}
