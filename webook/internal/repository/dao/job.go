package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context, refreshInterval time.Duration) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGormJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

func (dao *GORMJobDAO) Preempt(ctx context.Context, refreshInterval time.Duration) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		err := db.Model(&Job{}).Where("status = ? AND next_time < ?",
			jobStatusWaiting, now).First(&j).Error
		//if err == gorm.ErrRecordNotFound {
		//	err = db.Model(&Job{}).Where("status = ? AND utime < ?",
		//		jobStatusRunning, time.Now().Add(-3*refreshInterval)).First(&j).Error
		//	if err != nil {
		//		return j, err
		//	}
		//}
		if err != nil {
			return j, err
		}
		res := db.Model(&Job{}).Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			//	没抢到
			continue
		}
		return j, err
	}
}

func (dao *GORMJobDAO) Release(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"utime": now,
	}).Error
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Cfg        string
	Status     int

	Version int

	NextTime int64 `gorm:"index"`

	Utime int64
	Ctime int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	//	jobStatusPaused 已经不需要调度了
	jobStatusPaused
)
