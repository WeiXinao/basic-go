package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
	"time"
)

// LoginMiddlewareBuilder 扩展性
type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

// IgnorePaths 中间方法，用于构建部分
func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 终结方法，返回你最终希望的数据
func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 go 的方式编码解码
	//gob.Register(time.Now())
	return func(ctx *gin.Context) {
		//不需要登录校验的
		if slices.Contains[[]string](l.paths, ctx.Request.URL.Path) {
			return
		}
		//不需要登录校验的
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		now := time.Now().UnixMilli()
		// 说明还没有刷新过，刚登陆，还没有刷新
		if updateTime == nil {
			sess.Set("userId", id)
			sess.Set("update_time", now)
			sess.Options(sessions.Options{
				MaxAge: 60,
			})
			sess.Save()
			return
		}

		// update_time 是有的
		updateTimeVal, ok := updateTime.(int64)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if now-updateTimeVal > 10*1000 {
			sess.Set("userId", id)
			sess.Set("update_time", now)
			sess.Options(sessions.Options{
				MaxAge: 60,
			})
			sess.Save()
		}
	}
}
