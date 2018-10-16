package mutex_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (MutexDriver) {
	return &defaultMutexDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

