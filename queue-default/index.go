package queue_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (QueueDriver) {
	return &defaultQueueDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

