package mutex_redis

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (MutexDriver) {
	return &redisMutexDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

