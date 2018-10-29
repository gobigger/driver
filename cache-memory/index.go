package cache_memory

import (
	. "github.com/gobigger/bigger"
)

func Driver() (CacheDriver) {
	return &memoryCacheDriver{}
}


func init() {
	Bigger.Driver("memory", Driver())
}

