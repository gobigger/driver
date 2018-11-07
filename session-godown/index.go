package session_godown

import (
	. "github.com/gobigger/bigger"
)

func Driver() (SessionDriver) {
	return &godownSessionDriver{}
}


func init() {
	Bigger.Driver("godown", Driver())
}

