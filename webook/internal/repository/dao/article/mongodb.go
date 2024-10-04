package article

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongodbBDDAO struct {
	//client *mongo.Client
	// 代表 webook 的
	//database *mongo.Database
	// 代表的是制作库
	col *mongo.Collection
	// 代表的是线上库
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func (m *MongodbBDDAO) GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongodbBDDAO) GetById(ctx *gin.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongodbBDDAO) GetPubById(ctx *gin.Context, id int64) (PublishArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongodbBDDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	// 你没有自增主键
	// GLOBAL UNIFY ID (GUID，全局唯一ID)
	return id, err
}

func (m *MongodbBDDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 这边就是校验了 author_id 是不是正确的 ID
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (m *MongodbBDDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 没法子引入事务的概念
	// 首先第一步，保存制作库
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	// 操作线上库，upsert 语义
	// 更新语义
	now := time.Now().UnixMilli()
	art.Utime = now
	//update := bson.E{"$set", art}
	//upsert := bson.E{"$setOnInsert", bson.D{bson.E{"ctime", now}}}
	updateV1 := bson.M{
		// 更新，如果不存在，就是插入
		"$set": PublishArticle(art),
		// 在插入的时候，要插入 ctime
		"$setOnInsert": bson.M{"ctime": now},
	}
	filter := bson.M{"id": art.Id}
	_, err = m.liveCol.UpdateOne(ctx, filter,
		//bson.D{update, upsert},
		updateV1,
		options.Update().SetUpsert(true))
	return id, err
}

func (m *MongodbBDDAO) Upsert(ctx context.Context, art PublishArticle) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongodbBDDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func NewMongoDBArticleDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongodbBDDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

//func ToUpdate(vals map[string]any) bson.M {
//	return vals
//}
//
//func toFilter(vals map[string]any) bson.D {
//	var res bson.D
//	for k, v := range vals {
//		res = append(res, bson.E{k, v})
//	}
//	return res
//}
//
//func Set(vals map[string]any) bson.M {
//	return bson.M{"$set": bson.M(vals)}
//}
//
//func Upsert(vals map[string]any) bson.M {
//	return bson.M{"$set": bson.M(vals), "$setOnInsert":}
//}
