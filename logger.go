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
	// 日志级别优先级：DEBUG > INFO > ERROR
	levelPriority := map[string]int{
		ERROR: 0,
		INFO:  1,
		DEBUG: 2,
	}

	currentPriority := levelPriority[logLevel]
	msgPriority := levelPriority[level]

	// 只输出优先级大于等于当前级别的日志
	if msgPriority <= currentPriority {
		logger.Output(2, fmt.Sprintf("[%s] %s \n", level, message))
	}
}
