package event_redis

import (
	. "github.com/gobigger/bigger"
)

func Driver() (EventDriver) {
	return &redisEventDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

