package main

import (
	"fmt"
	"log"
)

const (
	INFO  = "INFO"
	DEBUG = "DEBUG"
	ERROR = "ERROR"
)

var (
	logger   *log.Logger
	logLevel = INFO
)

func setLogLevel(level string) {
	logLevel = level
}

func logMessage(level, message string) {
	if level == ERROR || (level == INFO && logLevel != DEBUG) || logLevel == DEBUG {
		logger.Output(2, fmt.Sprintf("[%s] %s \n", level, message))
	}
}
