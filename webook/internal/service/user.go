package service

import (
	"context"
	"errors"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}
func (svc *UserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
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

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Edit(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateProfile(ctx, u)
}
