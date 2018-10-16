package cache_memcache

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (CacheDriver) {
	return &memcacheCacheDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

