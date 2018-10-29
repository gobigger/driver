package event__default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (EventDriver) {
	return &defaultEventDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

