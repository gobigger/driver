package queue_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (QueueDriver) {
	return &defaultQueueDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

