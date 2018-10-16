package logger_file

import (
	. "github.com/yatlabs/bigger"
)

func Driver() (LoggerDriver) {
	return &fileLoggerDriver{}
}


func init() {
	Bigger.Driver("file", Driver())
}

