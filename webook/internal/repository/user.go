package repository

import (
	"context"
	"database/sql"
	"github.com/WeiXinao/basic-go/webook/internal/domain"
	"github.com/WeiXinao/basic-go/webook/internal/repository/cache"
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	UpdateProfile(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByEmail(ctx context.Context, u domain.User) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedUserRepository) UpdateProfile(ctx context.Context, u domain.User) error {
	return r.dao.UpdateProfile(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
	})
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

	u = r.entityToDomain(userModel)
	err = r.cache.Set(ctx, u)
	if err != nil {
		// 这里怎么办？
		// 打日志，做监控
	}

	return u, err
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, u domain.User) (domain.User, error) {
	foundUser, err := r.dao.FindByEmail(ctx, r.domainToEntity(u))
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(foundUser), nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	foundUser, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(foundUser), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 我确实有 email
			Valid: u.Email != "",
		},
		Password: u.Password,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Nickname: u.Nickname,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
