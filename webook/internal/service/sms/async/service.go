package async

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/service/sms"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/WeiXinao/basic-go/webook/pkg/ratelimit"
	"time"
)

//var errLimited = fmt.Errorf("触发了限流")
//
//type SMSService struct {
//	svc         sms.Service
//	repo        repository.SMSAsyncReqRepository
//	limiterKey  string
//	limiter     ratelimit.Limiter
//	rate        *big.Float
//	respTimeSum *big.Float
//	avgRespTime *big.Float
//	Cnt         *big.Float
//}
//
//func (s *SMSService) StartAsync() func() {
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			default:
//				ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
//				reqs := s.repo.Find没法出去的请求()
//				for _, req := range reqs {
//					// 在这里发送，并且控制重试
//					s.svc.Send(ctx2, req.biz, req.args, req.numbers...)
//				}
//				cancel2()
//			}
//		}
//	}()
//	return func() {
//		cancel()
//	}
//}
//
//func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
//	limited, err := s.limiter.Limit(ctx, s.limiterKey)
//	if err != nil {
//		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
//	}
//	if limited {
//		// 触发限流
//		// 将请求添加到数据库
//		return errLimited
//	}
//	start := time.Now()
//	err = s.svc.Send(ctx, biz, args, numbers...)
//	if err != nil {
//		return err
//	}
//	dur := big.NewFloat(float64(time.Since(start)))
//
//	sum := big.NewFloat(0)
//	sum.Add(s.respTimeSum, dur)
//
//	cnt := big.NewFloat(0)
//	cnt.Add(s.Cnt, big.NewFloat(1))
//
//	avg := big.NewFloat(0)
//	avg.Quo(sum, cnt)
//
//	rateAvg := avg.Mul(s.avgRespTime, s.rate)
//	if s.Cnt.Cmp(big.NewFloat(0)) == 0 || avg.Cmp(rateAvg) == -1 {
//		s.respTimeSum.Set(sum)
//		s.avgRespTime.Set(avg)
//		s.Cnt.Set(cnt)
//	} else {
//		//	服务商已经崩溃
//		// 将请求添加到数据库
//	}
//}
//
//func NewSMSService(svc sms.Service, repo repository.SMSAsyncReqRepository,
//	limiterKey string, limiter ratelimit.Limiter, rate float64) sms.Service {
//	return &SMSService{
//		svc:         svc,
//		repo:        repo,
//		limiterKey:  limiterKey,
//		limiter:     limiter,
//		rate:        big.NewFloat(rate),
//		respTimeSum: big.NewFloat(0),
//		avgRespTime: big.NewFloat(0),
//		Cnt:         big.NewFloat(0),
//	}
//}

type Service struct {
	svc sms.Service
	// 转异步，存储发短信请求的 repository
	repo       repository.AsyncSmsRepository
	l          logger.LoggerV1
	limiterKey string
	limiter    ratelimit.Limiter
	rate       float64
	sendTimes  map[int64]int64
	avgTime    int64
}

func NewService(svc sms.Service,
	repo repository.AsyncSmsRepository,
	l logger.LoggerV1, limiterKey string,
	limiter ratelimit.Limiter, rate float64) *Service {
	res := &Service{
		svc:        svc,
		repo:       repo,
		l:          l,
		limiterKey: limiterKey,
		limiter:    limiter,
		rate:       rate,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 异步发消息
// 这里我们没有设计退出机制，因为没啥必要
// 因为程序停止的时候，他自然就停止了
// 原理：这里是最简单的抢占式调度
func (s *Service) StartAsyncCycle() {
	// 这个是为了测试而引入的， 防止你在运行测试的时候，会出现偶发性失败
	time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	抢占一个异步发送的消息，确保在非常多的实例
	//	 比如 k8s 部署了三个 pod，一个请求，只有一个实例能拿到
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		//	执行发送
		//	这个也可以做成配置的
		ctx, cancal := context.WithTimeout(context.Background(), time.Second)
		defer cancal()
		err = s.svc.Send(ctx, as.TpId, as.Args, as.Numbers...)
		if err != nil {
			//	啥也不要干
			s.l.Error("执行异步发送短信失败",
				logger.Error(err),
				logger.Int64("id", as.Id))
		}
		res := err == nil
		//	通知 repository 我这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Int64("id", as.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		//	睡一秒。这个你可以自己决定
		time.Sleep(time.Second)
	default:
		//	正常来说应该是数据库那边出现了问题，
		//	但是为了尽量运行，还是要继续的
		//	你可以稍微睡眠，也可以不睡眠
		//	睡眠的话可以帮你规避掉短时间的网络抖动问题
		s.l.Error("抢占异步发送短信任务失败",
			logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	if s.needAsync() {
		// 需要异步发送，直接转储到数据库
		err := s.repo.Add(ctx, domain.AsyncSms{
			TpId:    biz,
			Args:    args,
			Numbers: numbers,
			//	设置可以重试三次
			RetryMax: 3,
		})
		return err
	}
	start := time.Now().UnixMilli()
	err := s.svc.Send(ctx, biz, args, numbers...)
	end := time.Now().UnixMilli()
	s.sendTimes[end] = end - start
	return err
}

func (s *Service) needAsync() bool {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟所有请求的响应时间比上一秒增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	limited, err := s.limiter.Limit(ctx, s.limiterKey)
	if err != nil {
		s.l.Error("短信服务判断是否限流出现问题 ", logger.Error(err))
		return false
	}
	if limited {
		return limited
	}

	var sum int64 = 0
	for t, d := range s.sendTimes {
		if t > time.Now().UnixMilli()-time.Second.Milliseconds() {
			sum += d
		} else {
			delete(s.sendTimes, t)
		}
	}

	avg := sum / int64(len(s.sendTimes))
	if s.avgTime != 0 || float64(avg) > s.rate*float64(s.avgTime) {
		return true
	} else {
		s.avgTime = avg
		return false
	}
}
