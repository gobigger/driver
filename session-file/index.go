package session_file

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SessionDriver) {
	return &fileSessionDriver{}
}


func init() {
	Bigger.Driver("file", Driver())
}

