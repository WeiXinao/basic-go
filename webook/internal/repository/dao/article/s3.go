package article

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	GORMArticleDAO
	oss *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, oss *s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{GORMArticleDAO: GORMArticleDAO{db: db}, oss: oss}
}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, author).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 0 {
			// 要么 ID 是错误的，要么作者不对
			// 后者情况下，你就要小心，可能有人搞你的系统
			// 没必要用 ID 搜索数据库来区分两种情况
			// 用 prometheus 打点，只要频繁出现，你就告警。然后手工介入排查
			return fmt.Errorf("可能有人在搞你，误操作非自己的文章, uid: %d, aid: %d", author, id)
		}
		return tx.Model(&PublishArticle{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
	if err != nil {
		return err
	}
	const statusPrivate = 3
	if status == statusPrivate {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1311988500"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	var (
		id = art.Id
	)
	// tx => transaction, trx, txn
	// 在事务内部，这里采用了闭包形态
	// GORM 帮我们管理了事务的生命周期
	// Begin，Rollback 和 Commit 都不需要我们操心
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}

		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now

		// 要操作线上库了
		return txDAO.Upsert(ctx, pubArt)
	})
	if err != nil {
		return 0, err
	}
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1311988500"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}
