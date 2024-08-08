package ioc

import (
	"github.com/WeiXinao/basic-go/webook/internal/service/sms"
	"github.com/WeiXinao/basic-go/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 换内存，还是换别的
	return memory.NewService()
}
