package logger_default

import (
	. "github.com/gobigger/bigger"
)

func Driver() (LoggerDriver) {
	return &defaultLoggerDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

