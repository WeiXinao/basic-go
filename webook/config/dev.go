//go:build !k8s

// 没有 k8s 这个编译标签
package config

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(192.168.5.3:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "192.168.5.3:6379",
	},
}
