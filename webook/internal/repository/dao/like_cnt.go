package dao

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"github.com/WeiXinao/xkit/slice"
	"gorm.io/gorm"
)

type LikeCntDao interface {
	GetLikeCntList(ctx context.Context) ([]LikeCnt, []article.Article, error)
}

type GORMLikeCntDao struct {
	db *gorm.DB
}

func (g *GORMLikeCntDao) GetLikeCntList(ctx context.Context) ([]LikeCnt, []article.Article, error) {
	lcs := make([]LikeCnt, 0, 100)
	var arts []article.Article
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows, err := tx.Exec(`SELECT count(*) AS cnt, biz, biz_id 
FROM interactives 
GROUP BY biz, biz_id 
ORDER BY cnt DESC 
LIMIT 100`).Rows()
		if err != nil {
			return err
		}
		var lc LikeCnt
		for rows.Next() {
			tx.ScanRows(rows, lc)
			lcs = append(lcs, lc)
		}

		ids := slice.Map[LikeCnt, int64](lcs, func(idx int, src LikeCnt) int64 {
			return src.BizId
		})
		return tx.Model(&article.Article{}).Where("id IN ?", ids).Find(&arts).Error
	})
	return lcs, arts, err
}

type LikeCnt struct {
	Cnt   int64
	Biz   string
	BizId int64
}
