package scheduler

import (
	"context"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/pkg/ginx"
	"github.com/WeiXinao/basic-go/webook/pkg/gormx/connpool"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/WeiXinao/basic-go/webook/pkg/migrator"
	"github.com/WeiXinao/basic-go/webook/pkg/migrator/events"
	"github.com/WeiXinao/basic-go/webook/pkg/migrator/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
)

// Scheduler 用来统一管理整个迁移过程
// 它不是必须的，你可以理解为这是为了方便用户操作（和你理解）而引入的。
type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	//	如果你要允许多个全量校验同时运行
	fulls map[string]func()
}

// 这个也不是必须的，就是你可以考虑利用配置中心，监听配置中心的变化
// 把全量校验，增量校验做出分布式任务，利用分布式认为调度平台来调度
func NewScheduler[T migrator.Entity](
	l logger.LoggerV1,
	src *gorm.DB,
	dst *gorm.DB,
	// 这个是业务用的 DoubleWritePool
	pool *connpool.DoubleWritePool,
	producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		src:     src,
		dst:     dst,
		pool:    pool,
		l:       l,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {
			//	初始的时候，啥都不用做
		},
		cancelIncr: func() {
			//	初始的时候，啥都不用做
		},
		producer: producer,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	//	将这个暴露为 HTTP 接口
	//	你可以配上对应的 UI
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.DisFirst))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))
	server.POST("/incr/start", ginx.WrapBodyV1[StartIncrRequest](s.StartIncrementValidation))
}

// SrcOnly 只读写源表
func (s *Scheduler[T]) SrcOnly(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DisFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer), nil
	case connpool.PatternDstOnly, connpool.PatternDstFirst:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

func (s *Scheduler[T]) StopFullValidation(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	//	可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	//	取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		//	先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

type StartIncrRequest struct {
	Utime int64 `json:"utime"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrRequest) (ginx.Result, error) {
	//	可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	//	取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		//	先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()

	return ginx.Result{
		Msg: "OK",
	}, nil
}
