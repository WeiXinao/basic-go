package service

import (
	"context"
	"errors"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService interface {
	Login(ctx context.Context, u domain.User) (domain.User, error)
	SignUp(ctx context.Context, u domain.User) error
	Edit(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx *gin.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}
func (svc *userService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	// 先找用户
	foundUser, err := svc.repo.FindByEmail(ctx, u)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(u.Password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return foundUser, nil
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Edit(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateProfile(ctx, u)
}

func (svc *userService) Profile(ctx context.Context, id int64) (domain.User, error) {
	// 在系统内部，基本上都是用 ID 的。
	// 有些人的系统比较复杂，有一个 GUID（global unique ID）

	return svc.repo.FindById(ctx, id)
}

func (svc *userService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	// 这时候，这个地方要怎么办？
	// 这个叫做快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	// 要判断，有没有这个用户
	if err != repository.ErrUserNotFound {
		// nil 会进来这里
		// 不为 ErrUserNotFound 的也会进来这里
		return u, err
	}

	// 在系统资源不足，触发降级后，不执行慢路径了
	//if ctx.Value("降级") == true {
	//	return domain.User{}, errors.New("系统降级了")
	//}

	// 这个叫做慢路径
	// 你明确知道，没有这个用户
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 因为这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func PathDownGrade(ctx context.Context, quick, slow func()) {
	quick()
	if ctx.Value("降级") == true {
		return
	}
	slow()
}
