package service

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/xkit/queue"
	"github.com/WeiXinao/xkit/slice"
	"math"
	"time"
)

//go:generate mockgen -source=./ranking_service.go -package=svcmocks -destination=./mocks/ranking_service.mock.go RankingService
type RankingService interface {
	//	TopN 前 100 的
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 用来取点赞数
	intrSvc InteractiveService

	// 用来查找文章
	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int

	repo repository.RankingRepository
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面
	// 存到缓存里面
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue[Score](b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})

	for {
		//	取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		//	取点赞数
		intrMap, err := b.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			intr := intrMap[art.Id]
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				minEle, _ := topN.Peek()
				if minEle.score < score {
					minEle, _ = topN.Dequeue()
					_ = topN.Enqueue(ele)
				}
			}
		}
		offset = offset + len(arts)
		if len(arts) < b.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	res := make([]domain.Article, topN.Len())
	cnt := topN.Len() - 1
	for cnt >= 0 {
		ele, _ := topN.Dequeue()
		res[cnt] = ele.art
		cnt--
	}
	return res, nil
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			//	时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}
