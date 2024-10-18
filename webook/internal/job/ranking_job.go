package job

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	l       logger.LoggerV1
	timeout time.Duration

	client *rlock.Client
	key    string

	lockLock *sync.Mutex
	lock     *rlock.Lock
}

func NewRankingJob(
	svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration,
) *RankingJob {
	return &RankingJob{
		key:      "job:ranking",
		l:        l,
		client:   client,
		svc:      svc,
		lockLock: &sync.Mutex{},
		timeout:  timeout,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.lockLock.Lock()
	lock := r.lock
	if lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				//	重试的超时
			}, time.Second)
		if err != nil {
			r.lockLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
		}
		r.lock = lock
		r.lockLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				//	续约失败了
				//	你也没办法中断当下正在调度的热榜计算（如果有）
				r.l.Error("续约失败", logger.Error(er))
			}
			r.lockLock.Lock()
			r.lock = nil
			r.lockLock.Unlock()
		}()
	}
	// 这边你就是拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.lockLock.Lock()
	lock := r.lock
	r.lockLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
