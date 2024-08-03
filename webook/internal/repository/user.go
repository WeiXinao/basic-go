package repository

import (
	"context"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
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
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		return u, err
	}
	// 没这个数据
	//if err == cache.ErrKeyNotExist {
	//	// 去数据库里面加载
	//}

	// 这里要怎么办？err = io.EOF
	// 要不要去数据库加载？
	// 看起来我不应该加载？
	// 看起来我好像也要加载？

	// 选加载 --- 做好兜底，万一 Redis 真的崩了，你要保护住你的数据库
	// 我数据库限流

	// 选不加载 --- 用户体验差一点

	userModel, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	birthStr := time.UnixMilli(userModel.Birthday).Format(time.DateOnly)
	birth, err := time.Parse(time.DateTime, birthStr)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       userModel.Id,
		Email:    userModel.Email,
		Nickname: userModel.Nickname,
		Birthday: birth,
	}
	err = r.cache.Set(ctx, u)
	if err != nil {
		// 这里怎么办？
		// 打日志，做监控
	}

	return u, err
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
