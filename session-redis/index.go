package session_redis

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SessionDriver) {
	return &redisSessionDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

