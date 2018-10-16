package logger_default

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (LoggerDriver) {
	return &defaultLoggerDriver{}
}


func init() {
	Bigger.Driver("default", Driver())
}

