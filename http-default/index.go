package http_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (HttpDriver) {
	return &defaultHttpDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

