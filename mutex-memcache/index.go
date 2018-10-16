package mutex_memcache

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (MutexDriver) {
	return &memcacheMutexDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

