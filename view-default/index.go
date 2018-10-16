package view_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (ViewDriver) {
	return &defaultViewDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

