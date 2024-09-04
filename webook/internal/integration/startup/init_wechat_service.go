package startup

import (
	"github.com/WeiXinao/basic-go/webook/internal/service/oauth2/wechat"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
)

// InitPhantomWechatService 没啥用的虚拟的 wechatService
func InitPhantomWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
