package feishu

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"reflect"
	"testing"
)

func Test_service_VerifyCode(t *testing.T) {
	type fields struct {
		appId     string
		appSecret string
		client    *http.Client
	}
	type args struct {
		ctx   context.Context
		code  string
		state string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.FeishuInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				appId:     tt.fields.appId,
				appSecret: tt.fields.appSecret,
				client:    tt.fields.client,
			}
			got, err := s.VerifyCode(tt.args.ctx, tt.args.code, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
