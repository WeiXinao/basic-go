package repository

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) UpdateProfile(ctx context.Context, u domain.User) error {
	return r.dao.UpdateProfile(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
	})
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 里面找
	// 再从 dao 里面找
	// 最后写回 cache
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	birthStr := time.UnixMilli(u.Birthday).Format(time.DateOnly)
	birth, err := time.Parse(time.DateTime, birthStr)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Nickname: u.Nickname,
		Birthday: birth,
	}, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, u domain.User) (domain.User, error) {
	foundUser, err := r.dao.FindByEmail(ctx, dao.User{
		Email: u.Email,
	})
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       foundUser.Id,
		Email:    foundUser.Email,
		Password: foundUser.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}
