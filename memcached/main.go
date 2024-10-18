package main

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"time"
)

func main() {
	mc := memcache.New("192.168.5.4:11211")

	now := time.Now().Unix()
	err := mc.Set(&memcache.Item{
		Key:        "key1",
		Value:      []byte("value"),
		Expiration: 600,
		Flags:      uint32(now),
	})
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)
	item, err := mc.Get("key1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(item.Value), item.Flags == uint32(now))
}
