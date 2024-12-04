package startup

import (
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
