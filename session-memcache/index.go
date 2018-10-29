package session_memcache

import (
	. "github.com/gobigger/bigger"
)

func Driver() (SessionDriver) {
	return &memcacheSessionDriver{}
}


func init() {
	Bigger.Driver("memcache", Driver())
}

