package session_memory

import (
	. "github.com/gobigger/bigger"
)

func Driver() (SessionDriver) {
	return &memorySessionDriver{}
}


func init() {
	Bigger.Driver("memory", Driver())
}

