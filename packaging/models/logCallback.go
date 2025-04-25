package models

import (
	"log"
	"time"
)

const (
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorReset   = "\033[0m"
)

const (
	VERBOSE = iota + 1
	DEBUG
	INFO
	WARN
	ERROR
)

var logLevel = DEBUG

type LogCallback interface {
	LogInfo(data ...any)
	LogDebug(data ...any)
	LogVerbose(data ...any)
	//Printf(str string, data ...any)
}

type LogImpl struct {
}

func (LogImpl) LogVerbose(data ...any) {
	if logLevel <= VERBOSE {
		log.Println(append([]interface{}{"[VERBOSE]", time.DateTime}, data...)...)
	}
}

func (LogImpl) LogInfo(data ...any) {
	if logLevel <= INFO {
		log.Println(append([]interface{}{colorBlue + "[INFO]" + colorReset, time.DateTime}, data...)...)
	}
}

func (LogImpl) LogDebug(data ...any) {
	if logLevel <= DEBUG {
		log.Println(append([]interface{}{colorGreen + "[DEBUG]" + colorReset, time.DateTime}, data...)...)
	}
}

func LogInterface() *LogImpl {
	return &LogImpl{}
}
