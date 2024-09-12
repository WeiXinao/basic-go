package dao

import (
	"bytes"
	"context"
	"github.com/WeiXinao/xkit"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
	"time"
)

func TestS3(t *testing.T) {
	cosId, ok := os.LookupEnv("COS_APP_ID")
	if !ok {
		panic("没有找到环境变量 COS_APP_ID")
	}
	cosKey, ok := os.LookupEnv("COS_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 COS_APP_SECRET")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(cosId, cosKey, ""),
		Region:      xkit.ToPtr[string]("ap-nanjing"),
		Endpoint:    xkit.ToPtr[string]("https://cos.ap-nanjing.myqcloud.com"),
		// 强制使用 /bucket/key 的形态
		S3ForcePathStyle: xkit.ToPtr[bool](false),
	})
	assert.NoError(t, err)
	client := s3.New(sess)
	ctx, cancal := context.WithTimeout(context.Background(), time.Second*10)
	defer cancal()
	_, err = client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1311988500"),
		Key:         ekit.ToPtr[string]("12634"),
		Body:        bytes.NewReader([]byte("测试内容 abc")),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	assert.NoError(t, err)

	res, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: ekit.ToPtr[string]("webook-1311988500"),
		Key:    ekit.ToPtr[string]("12634"),
	})
	assert.NoError(t, err)
	data, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	t.Log(string(data))
}
