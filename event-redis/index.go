package event_redis

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (EventDriver) {
	return &redisEventDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

