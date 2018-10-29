package cache_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (CacheDriver) {
	return &defaultCacheDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

