package socket_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (SocketDriver) {
	return &defaultSocketDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

