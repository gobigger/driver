package cache_file

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (CacheDriver) {
	return &fileCacheDriver{}
}


func init() {
	Bigger.Driver("file", Driver())
}

