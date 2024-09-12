package mongo

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongo(t *testing.T) {
	// 控制初始化超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		// 每个命令（查询）执行之前
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
		// 执行成功
		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {

		},
		// 执行失败
		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {

		},
	}
	opts := options.Client().ApplyURI("mongodb://localhost:27017").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	mdb := client.Database("webook")
	col := mdb.Collection("articles")
	res, err := col.InsertOne(ctx, Article{
		Id:      123,
		Title:   "我的标题",
		Content: "我的内容",
	})
	// 这个是文档 ID，也就是 mongodb 中的 _id 字段
	fmt.Printf("id %s", res.InsertedID)
}

type Article struct {
	Id       int64
	Title    string
	Content  string
	AuthorId int64
	Status   uint8
	Ctime    int64
	Utime    int64
}
