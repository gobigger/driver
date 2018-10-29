package cache_memcache

import (
	. "github.com/gobigger/bigger"
)

func Driver() (CacheDriver) {
	return &memcacheCacheDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

