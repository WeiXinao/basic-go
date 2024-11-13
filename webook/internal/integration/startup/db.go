package startup

import (
	"github.com/WeiXinao/basic-go/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 测试的话，不用控制并发。等遇到了并发问题再说
func InitDB() *gorm.DB {
	dsn := "root:123456@tcp(192.168.5.4:3307)/webook"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
