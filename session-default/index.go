package session_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SessionDriver) {
	return &defaultSessionDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

