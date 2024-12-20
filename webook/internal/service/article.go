package service

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	artEvent "github.com/WeiXinao/basic-go/webook/internal/events/article"
	articleEvent "github.com/WeiXinao/basic-go/webook/internal/events/article"
	"github.com/WeiXinao/basic-go/webook/internal/repository/article"
	logger "github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"time"
)

var ErrInteractiveNotFound = article.ErrInteractiveNotFound

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	GetPubById(ctx *gin.Context, id int64, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     article.ArticleRepository
	producer articleEvent.Producer

	// V1 依靠两个不同的 repository 来解决这种跨表，或者跨库的问题
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1

	ch chan readInfo
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

type readInfo struct {
	uid int64
	aid int64
}

func NewArticleService(repo article.ArticleRepository,
	producer articleEvent.Producer,
	l logger.LoggerV1) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository,
	reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func NewArticleServiceV2(repo article.ArticleRepository,
	producer articleEvent.Producer,
	l logger.LoggerV1) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		for {
			uids := make([]int64, 0, 10)
			aids := make([]int64, 0, 10)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			for i := 0; i < 10; i++ {
				select {
				case info, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					uids = append(uids, info.uid)
					aids = append(aids, info.aid)
				case <-ctx.Done():
					break
				}
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			producer.ProducerReadEventV1(ctx, artEvent.ReadEventV1{
				Uids: uids,
				Aids: aids,
			})
			cancel()
		}
	}()
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
	}
}

func (a *articleService) GetPubById(ctx *gin.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err != nil {
			//	生产者也可以通过改批量来提高性能
			er := a.producer.ProducerReadEvent(articleEvent.ReadEvent{
				Aid: id,
				Uid: id,
			})

			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(er))
			}
		}
	}()

	go func() {
		//	改批量的做法
		a.ch <- readInfo{
			aid: id,
			uid: uid,
		}
	}()

	return res, err
}

func (a *articleService) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	//art.Status = domain.ArticleStatusPrivate 然后把整个 art 往下传
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	//// 制作库？
	//id, err := a.repo.Create(ctx, art)
	//// 线上库呢？
	//a.repo.SyncToLiveDB(ctx, art)
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 制作和线上库的 ID 是想等的
	art.Id = id
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("部分失败，保存到线上库失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，保存到线上库失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
		// 接入你的告警系统，手工处理下
		// 走异步，我直接保存到文件本地
		// 走 Canal
		// 打 MQ
	}
	return id, err
	// 我先判断一下，线上库里面有没有这篇文章（就是看之前有没有发表过）
	// 有并发问题
	//ok := a.reader.Exists(art.Id)
	//if ok {
	//	//	之前就发表过
	//	a.reader.Update()
	//} else {
	//	return a.reader.Create(ctx, a)
	//}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Close() {
	close(a.ch)
}
