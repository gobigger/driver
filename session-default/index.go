package session_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (SessionDriver) {
	return &defaultSessionDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

