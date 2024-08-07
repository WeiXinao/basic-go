package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) UpdateProfile(ctx context.Context, u User) error {
	return dao.db.WithContext(ctx).Model(&u).Select("nickname", "birthday", "about_me").Updates(u).Error
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return u, err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, u User) (User, error) {
	var foundUser User
	err := dao.db.WithContext(ctx).Where("email = ?", u.Email).First(&foundUser).Error
	//err := dao.db.WithContext(ctx).First(&foundUser, "email = ?", u.Email).Error
	return foundUser, err
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

// User 直接对应数据库表结构
// 有些人叫做 entity, 有些人叫做 model, 也有些人叫做 PO(persistent object)
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 全部用户唯一
	Email    string `gorm:"unique"`
	Password string

	// 往这里面加
	Nickname string
	Birthday int64
	AboutMe  string

	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}
