package cache_file

import (
	. "github.com/gobigger/bigger"
)

func Driver() (CacheDriver) {
	return &fileCacheDriver{}
}


func init() {
	Bigger.Driver("file", Driver())
}

