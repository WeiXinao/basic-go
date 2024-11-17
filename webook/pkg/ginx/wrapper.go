package ginx

import (
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
)

// 这个东西，放到你们的 ginx 插件库里面去
// 技术含量不是很高，但是绝对有技巧

// L 使用包变量
var L logger.LoggerV1

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

func WrapClaims[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		c := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		// 下半段的业务逻辑从哪里来？
		// 我的因为逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, c)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			//	开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyAndClaims[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		c := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		// 下半段的业务逻辑从哪里来？
		// 我的因为逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req, c)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			//	开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		// 下半段的业务逻辑从哪里来？
		// 我的因为逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			//	开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}
		// 下半段的业务逻辑从哪里来？
		// 我的因为逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			//	开始处理 error，其实就是记录一下日志
			l.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
