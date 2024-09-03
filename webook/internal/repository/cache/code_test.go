package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	type fields struct {
		client redis.Cmdable
	}
	type args struct {
		ctx   context.Context
		biz   string
		phone string
		code  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RedisCodeCache{
				client: tt.fields.client,
			}
			if err := c.Set(tt.args.ctx, tt.args.biz, tt.args.phone, tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
