package event__default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (EventDriver) {
	return &defaultEventDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

