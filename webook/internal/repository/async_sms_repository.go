package repository

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"github.com/WeiXinao/xkit/sqlx"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSmsRepository interface {
	// Add 添加一个异步 SMS 记录
	Add(ctx context.Context, s domain.AsyncSms) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSmsRepository struct {
	dao dao.AsyncSmsDAO
}

func (a *asyncSmsRepository) Add(ctx context.Context, s domain.AsyncSms) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TpId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return domain.AsyncSms{
		Id:       as.Id,
		TpId:     as.Config.Val.TplId,
		Args:     as.Config.Val.Args,
		Numbers:  as.Config.Val.Numbers,
		RetryMax: as.RetryMax,
	}, nil
}

func (a *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}

func NewAsyncSMSRepository(dao dao.AsyncSmsDAO) AsyncSmsRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}
