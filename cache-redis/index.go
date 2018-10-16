package cache_redis

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (CacheDriver) {
	return &redisCacheDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

