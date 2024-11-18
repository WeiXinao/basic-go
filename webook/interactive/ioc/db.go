package ioc

import (
	"github.com/WeiXinao/basic-go/webook/interactive/repository/dao"
	"github.com/WeiXinao/basic-go/webook/pkg/gormx"
	"github.com/WeiXinao/basic-go/webook/pkg/gormx/connpool"
	"github.com/WeiXinao/basic-go/webook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB() SrcDB {
	return initDB("src")
}

func InitDstDB() DstDB {
	return initDB("dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.LoggerV1) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: p,
	}))
	if err != nil {
		panic(err)
	}
	return doubleWrite
}

func initDB(key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root:root@tcp(192.168.5.4:3307)/webook",
	}
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook_" + key,
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	cb := gormx.NewCallback(prometheus2.SummaryOpts{
		Namespace: "xiaoxin",
		Subsystem: "webook",
		Name:      "gorm_db_" + key,
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.001,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(),
		tracing.WithDBName("webook_"+key)))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

type DoSomething interface {
	DoABC() string
}

type DoSomethingFunc func() string

func (d DoSomethingFunc) DoABC() string {
	return d()
}
