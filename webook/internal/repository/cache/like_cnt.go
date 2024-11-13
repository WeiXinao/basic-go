package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/xkit/slice"
	"github.com/redis/go-redis/v9"
)

//go:embed lua/add_articles.lua
var addArticlesScript string

type LikeCntCache interface {
	AddLikeCntZSet(ctx context.Context, likeCnts []domain.LikeCnt) error
	SetArticles(ctx context.Context, arts []domain.Article) error
}

type LikeCntRedisCache struct {
	client redis.Cmdable
}

func (l *LikeCntRedisCache) SetArticles(ctx context.Context, arts []domain.Article) error {
	artStrs := slice.Map[domain.Article, string](arts, func(idx int, src domain.Article) string {
		marshal, _ := json.Marshal(src)
		return string(marshal)
	})
	keys := slice.Map(arts, func(idx int, src domain.Article) string {
		return l.key(src.Id)
	})
	res, err := l.client.Eval(ctx, addArticlesScript, keys, artStrs).Int()
	if err != nil {
		return err
	}
	if res == 1 {
		return errors.New("缓存排行榜文章失败")
	}
	return nil
}

func (l *LikeCntRedisCache) AddLikeCntZSet(ctx context.Context, likeCnts []domain.LikeCnt) error {
	likeCntZs := slice.Map[domain.LikeCnt, redis.Z](likeCnts, func(idx int, src domain.LikeCnt) redis.Z {
		src.Cnt = 0
		marshal, _ := json.Marshal(src)
		return redis.Z{
			Score:  float64(src.Cnt),
			Member: marshal,
		}
	})
	return l.client.ZAdd(ctx, "top_100", likeCntZs...).Err()
}
func (a *LikeCntRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}
