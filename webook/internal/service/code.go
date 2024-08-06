package service

import (
	"context"
	"fmt"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/WeiXinao/basic-go/webook/internal/service/sms"
	"math/rand"
	"time"
)

const CodeTplId = "1877556"

func init() {
	rand.Seed(time.Now().UnixMilli())
}

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func (svc *CodeService) Send(ctx context.Context,
	// 区别业务场景
	biz string,
	phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, CodeTplId, []string{code}, phone)
	//if err != nil {
	// 这个地方怎么办？
	// 这意味着，Redis 有这个验证码，但是不好意思
	// 我能不能删掉这个验证码？不能删除
	// 你这个 err 可能是超时的 err, 你都不知道，发出了没
	// 在这里重试
	// 要重试的话，初始化的时候，传入一个自己就会重试的 smsSvc
	//}

	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string,
	phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	// 六位数，num 在[0, 1000000)
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}

//func (svc *CodeService) VerifyV1(ctx context.Context, biz string,
//	phone string, inputCode string) error {
//	return nil
//}
