package mutex_memcache

import (
	. "github.com/gobigger/bigger"
)

func Driver() (MutexDriver) {
	return &memcacheMutexDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

