package startup

import (
	"context"
	"database/sql"
	"github.com/WeiXinao/basic-go/webook/comment/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var db *gorm.DB

func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:123456@tcp(192.168.5.4)/webook_comment"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err != nil {
				break
			}
			log.Println("等待连接 MySQL", err)
		}
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		db = db.Debug()
		err = dao.InitTables(db)
		if err != nil {
			panic(err)
		}
	}
	return db
}
