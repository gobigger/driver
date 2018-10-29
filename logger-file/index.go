package logger_file

import (
	. "github.com/gobigger/bigger"
)

func Driver() (LoggerDriver) {
	return &fileLoggerDriver{}
}


func init() {
	Bigger.Driver("file", Driver())
}

