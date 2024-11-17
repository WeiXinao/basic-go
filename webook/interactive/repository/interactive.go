package repository

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/interactive/domain"
	"github.com/WeiXinao/basic-go/webook/interactive/repository/cache"
	"github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/WeiXinao/xkit/slice"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CacheInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

func (c *CacheInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CacheInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	go func() {
		for i := 0; i < len(bizs); i++ {
			er := c.cache.IncrReadCntIfPresent(ctx, bizs[i], bizIds[i])
			if er != nil {
				//	记录日志
			}
		}
	}()
	return nil
}

func (c *CacheInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}

func (c *CacheInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, id)
}

func (c *CacheInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}

	res := c.toDomain(ie)
	err = c.cache.Set(ctx, biz, id, res)
	if err != nil {
		c.l.Error("会写缓存失败",
			logger.String("biz", biz),
			logger.Int64("bizId", id),
			logger.Error(err))
	}
	return res, nil
}

func (c *CacheInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CacheInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CacheInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (c *CacheInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//	你要更新缓存了
	//	部分失败问题 —— 数据不一致
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CacheInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO,
	cache cache.InteractiveCache) InteractiveRepository {
	return &CacheInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
