package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"strconv"
	"time"
)

type MemcachedCodeCache struct {
	client *memcache.Client
}

func NewMemcachedCodeCache(client *memcache.Client) CodeCache {
	return &MemcachedCodeCache{
		client: client,
	}
}

func (m *MemcachedCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	err := m.client.Add(&memcache.Item{
		Key:        m.key(biz, phone),
		Value:      []byte(code),
		Expiration: 600,
		Flags:      uint32(time.Now().Unix()),
	})
	if err == nil {
		err = m.client.Set(&memcache.Item{
			Key:        fmt.Sprintf("%s:cnt", m.key(biz, phone)),
			Value:      []byte("3"),
			Expiration: 600,
		})
		if err != nil {
			return errors.New("系统错误")
		} else {
			return nil
		}
	}
	if err != memcache.ErrNotStored {
		return errors.New("系统错误")
	}
	item, err := m.client.Get(m.key(biz, phone))
	if err != nil {
		return errors.New("系统错误")
	}
	if !m.canSend(item) {
		return ErrCodeSendTooMany
	}
	err = m.client.CompareAndSwap(&memcache.Item{
		Key:        m.key(biz, phone),
		Value:      []byte(code),
		Expiration: 600,
		Flags:      uint32(time.Now().Unix()),
		CasID:      item.CasID,
	})
	if err != nil {
		return errors.New("系统错误")
	}
	err = m.client.Set(&memcache.Item{
		Key:        fmt.Sprintf("%s:cnt", m.key(biz, phone)),
		Value:      []byte("3"),
		Expiration: 600,
	})
	if err != nil {
		return errors.New("系统错误")
	} else {
		return nil
	}
}

func (m *MemcachedCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	cntItem, err := m.client.Get(fmt.Sprintf("%s:cnt", m.key(biz, phone)))
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	if err != nil {
		return false, ErrUnknownForCode
	}
	cnt, err := strconv.Atoi(string(cntItem.Value))
	if err != nil {
		return false, ErrUnknownForCode
	}
	if cnt <= 0 {
		err := m.client.Delete(m.key(biz, phone))
		err = m.client.Delete(m.key(biz, phone) + ":cnt")
		if err != nil {
			return false, ErrUnknownForCode
		}
		return false, ErrCodeVerifyTooManyTimes
	}
	codeItem, err := m.client.Get(m.key(biz, phone))
	if err != nil {
		return false, ErrUnknownForCode
	}
	if string(codeItem.Value) == inputCode {
		err := m.client.Delete(m.key(biz, phone))
		err = m.client.Delete(m.key(biz, phone) + ":cnt")
		if err != nil {
			return false, ErrUnknownForCode
		}
		return true, nil
	}
	_, err = m.client.Decrement(m.key(biz, phone)+":cnt", 1)
	if err != nil {
		return false, ErrUnknownForCode
	} else {
		return false, nil
	}
}

func (m *MemcachedCodeCache) canSend(item *memcache.Item) bool {
	now := uint32(time.Now().Unix())
	return now-item.Flags >= 60
}

func (m *MemcachedCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
