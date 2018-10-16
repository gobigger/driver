package session_memcache

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SessionDriver) {
	return &memcacheSessionDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

