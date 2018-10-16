package cache_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (CacheDriver) {
	return &defaultCacheDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

