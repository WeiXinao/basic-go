package dao

import (
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&article.PublishArticle{},
		&Job{})
}
