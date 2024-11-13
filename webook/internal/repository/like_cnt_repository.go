package repository

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"github.com/WeiXinao/xkit/slice"
	"time"
)

type LikeCntRepository interface {
	LikeCntTop100(ctx context.Context) error
}

type CachedLikeCntRepository struct {
	dao   dao.LikeCntDao
	cache cache.LikeCntCache
}

func (c *CachedLikeCntRepository) LikeCntTop100(ctx context.Context) error {
	lcs, articles, err := c.dao.GetLikeCntList(ctx)
	if err != nil {
		return err
	}
	err = c.cache.AddLikeCntZSet(ctx, slice.Map[dao.LikeCnt, domain.LikeCnt](lcs,
		func(idx int, src dao.LikeCnt) domain.LikeCnt {
			return domain.LikeCnt{
				Cnt:   src.Cnt,
				Biz:   src.Biz,
				BizId: src.BizId,
			}
		}))
	if err != nil {
		return err
	}
	return c.cache.SetArticles(ctx, slice.Map(articles, func(idx int, src article.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Content: src.Content,
			Author:  domain.Author{Id: src.AuthorId},
			Status:  domain.ArticleStatus(src.Status),
			Ctime:   time.UnixMilli(src.Ctime),
			Utime:   time.UnixMilli(src.Utime),
		}
	}))
}
