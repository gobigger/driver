package view_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (ViewDriver) {
	return &defaultViewDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

