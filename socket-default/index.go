package socket_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (SocketDriver) {
	return &defaultSocketDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

