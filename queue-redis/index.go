package queue_redis

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (QueueDriver) {
	return &redisQueueDriver{}
}


func init() {
	Bigger.Driver("redis", Driver())
}

