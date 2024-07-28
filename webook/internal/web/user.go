package web

import (
	"fmt"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// UserHandler 我准备在它上面定义跟用户有关的路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)

	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	{
		ug.POST("/signup", u.SignUp)
		//ug.POST("/login", u.Login)
		ug.POST("/login", u.LoginJWT)
		ug.POST("/edit", u.Edit)
		ug.GET("/profile", u.Profile)
	}
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}

	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 调用一下 svc 的方法
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里用 JWT 设置登录态
	// 生成一个 JWT token
	// 下一节课，如果我们要在 JWT token 里面带我个人的数据，该怎么带？
	// 比如，我要带 userId
	token := jwt.New(jwt.SigningMethodHS512)
	tokenStr, err := token.SignedString([]byte("jDDkHf8BqSaREsYqM5ntKTCXQYUGgCpW"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	fmt.Println(tokenStr)
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 sess 里面的值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		MaxAge:   60,
	})
	err = sess.Save()
	if err != nil {
		panic(err)
	}

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 sess 里面的值
	sess.Options(sessions.Options{
		//Secure:   true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	err := sess.Save()
	if err != nil {
		panic(err)
	}

	ctx.String(http.StatusOK, "退出登录成功")
}
func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var editReq EditReq
	if err := ctx.Bind(&editReq); err != nil {
		return
	}

	if len(editReq.Nickname) <= 0 {
		ctx.String(http.StatusOK, "昵称不能为空")
		return
	}
	if len(editReq.Nickname) > 20 {
		ctx.String(http.StatusOK, "昵称太长，不能超过20个字符")
		return
	}

	brith, err := time.Parse(time.DateOnly, editReq.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "非法日期")
		return
	}

	if len(editReq.AboutMe) > 100 {
		ctx.String(http.StatusOK, "简介太长")
		return
	}

	sess := sessions.Default(ctx)
	id := sess.Get("userId")
	if id == nil {
		ctx.String(http.StatusUnauthorized, "请先登录")
		return
	}

	err = u.svc.Edit(ctx, domain.User{
		Id:       id.(int64),
		Nickname: editReq.Nickname,
		Birthday: brith,
		AboutMe:  editReq.AboutMe,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "编辑成功")
	return
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是你的 Profile")
}
