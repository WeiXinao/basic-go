package startup

import (
	"github.com/WeiXinao/basic-go/webook/internal/service/oauth2/feishu"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
)

// InitPhantomWechatService 没啥用的虚拟的 feishuService
func InitPhantomWechatService(l logger.LoggerV1) feishu.Service {
	return feishu.NewService("", "", l)
}
