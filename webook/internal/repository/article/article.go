package article

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	dao "github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

// repository 还是要用来操作缓存和 DAO
// 事务概念还是应该在 DAO 这一层

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	GetPubById(ctx *gin.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	// 如果你直接访问 UserDAO, 你就饶开了 repository,
	// repository 一般有一些缓存机制
	userRepo repository.UserRepository

	// v1 操作两个 DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果你要在 repository 层面上操作事务
	// 那么就只能利用 db 开启事务后，创建基于事务的 DAO
	// 或者，去掉 DAO 这一层，在 repository 的实现中，直接操作 db
	db *gorm.DB
}

func (c *CachedArticleRepository) GetPubById(ctx *gin.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err != nil {
		return res, nil
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	//	现在要去查询对应的 User 信息，拿到作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = c.cache.SetPub(ctx, res)
		if err != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err != nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	go func() {
		err = c.cache.Set(ctx, res)
		if err != nil {
			//	记录日志
		}
	}()
	return res, err
}

func (c *CachedArticleRepository) GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, nil
		} else {
			//	要考虑记录日志
			//	缓存未命中，你是可以忽略的
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit <= 100 {
			//	会写失败，不一定是大问题，但是有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我要监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			//	也要记录日志
		}
	}
	return id, err
}

// SyncV2 尝试在 repository 层面上解决事务问题
// 确保保存到制作库和线上库同时成功，或者同时失败
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启了一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 利用 tx 来构建 DAO
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	// 先保存到制作库，再保存到线上库
	if id > 0 {
		err = author.UpdateById(ctx, artn)
	} else {
		id, err = author.Insert(ctx, artn)
	}
	if err != nil {
		// 执行有问题，要回滚
		//tx.Rollback()
		return id, err
	}
	defer tx.Rollback()
	// 考虑上课库了，同步数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT OR UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = reader.UpsertV2(ctx, dao.PublishArticle(artn))
	// 执行成功，直接提交
	tx.Commit()
	return id, err
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	// 先保存到制作库，再保存到线上库
	if id > 0 {
		err = c.authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = c.authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		return id, err
	}
	// 考虑上课库了，同步数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT OR UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			//	也要记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		err = c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			//	也要记录日志
		}
	}
	return id, err
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func NewCachedArticleRepository(dao dao.ArticleDAO,
	userRepo repository.UserRepository,
	cache cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
		}
	}
}
