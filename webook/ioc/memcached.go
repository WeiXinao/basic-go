package ioc

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/viper"
)

func InitMemcached() *memcache.Client {
	addr := viper.GetString("memcached.addr")
	return memcache.New(addr)
}
