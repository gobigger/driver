package session_memory

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SessionDriver) {
	return &memorySessionDriver{}
}


func init() {
	Bigger.Driver("memory", Driver())
}

