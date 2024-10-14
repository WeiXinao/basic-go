package article

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
	GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx *gin.Context, id int64) (Article, error)
	GetPubById(ctx *gin.Context, id int64) (PublishArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishArticle, error)
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishArticle, error) {
	var res []PublishArticle
	const ArticleStatusPublished = 2
	err := dao.db.WithContext(ctx).Where("utime < 2 AND status = ?",
		start.UnixMilli(), ArticleStatusPublished).Offset(offset).Limit(limit).First(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) GetPubById(ctx *gin.Context, id int64) (PublishArticle, error) {
	var res PublishArticle
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) GetById(ctx *gin.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	var (
		id = art.Id
	)
	// tx => transaction, trx, txn
	// 在事务内部，这里采用了闭包形态
	// GORM 帮我们管理了事务的生命周期
	// Begin，Rollback 和 Commit 都不需要我们操心
	err := dao.db.Transaction(func(tx *gorm.DB) error {
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

		// 要操作线上库了
		return txDAO.Upsert(ctx, PublishArticle(art))
	})
	return id, err
}

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	// 这里是插入
	// OnConflict 的意思是数据冲突了
	err := dao.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT AAAA ON CONFLICT(BBB) DO NOTHING
		// INSERT AAAA ON CONFLICT(BBB) DO UPDATES CCC WHERE DDD

		// 那些列冲突
		// Columns: []clause.Column{clause.Column{Name: "id"}},
		// 意思是数据冲突了，啥也不干
		// DoNothing:
		// 数据冲突了，并且符合 WHERE 条件的就会执行更新
		// Where

		// MySQL 只需要关心这里
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&art).Error
	// MySQL 最终的语句 INSERT xxx ON DUPLICATE KEY UPDATE xxx

	// 一条 SQL 语句，都不需要开启事务
	//  autocommit: 意思是自动提交

	return err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键进行更新
	// 可读性很差
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id=? AND author_id = ?", art.Id, art.AuthorId).
		// 当你使用这种每次指定被更新的列的写法
		// 可读性强，但是每一次更新更多的列的时候，你都要修改
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	// 你要不要检查真的更新没？
	//res.RowsAffected // 更新行数
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// 补充一点日志
		return fmt.Errorf("更新失败, 可能是创作者非法 id %d author_id %d",
			art.Id, art.AuthorId)
	}
	return res.Error
}

// 事务传播机制是指如果当前有事务，就在事务内部执行 Insert
// 如果没有事务：
// 1. 开启事务，执行 Insert
// 2. 直接执行
// 3. 报错

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// Article 这是制作库的
// 准备在 articles 表中准备一百万条数据
// 准备一个 author_id = 123 的，插入 200 条数据
// 执行 SELECT * FROM articles WHERE author_id = 123 ORDER BY ctime DESC
// 比较两种索引的性能
type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 长度 1024
	Title   string `gorm:"type=varchar(1024)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 如何设计索引
	// WHERE
	// 在帖子这里，什么样的查询场景？
	// 对于创作者来说，是不是看草稿箱，看到所有自己的文章？
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM articles WHERE id = 1;
	// 在查询接口，我们深入讨论这个问题
	// - 最佳选择，就是要在 author_id 和 ctime 上创建联合索引
	// - 在 author_id 上创建索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`

	// 学学 Explain 命令

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`

	// 有些人考虑到，经常用状态来查询
	// WHERE status = xxx AND
	// 在 status 上和别的列混在一起，创建一个联合索引
	// 要看别的列究竟是什么列。
	Status uint8 `bson:"status,omitempty"`
	Ctime  int64 `bson:"ctime,omitempty"`
	Utime  int64 `bson:"utime,omitempty"`
}
