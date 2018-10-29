package mutex_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (MutexDriver) {
	return &defaultMutexDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

