package http_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (HttpDriver) {
	return &defaultHttpDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

