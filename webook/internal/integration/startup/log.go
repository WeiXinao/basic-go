package startup

import (
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
