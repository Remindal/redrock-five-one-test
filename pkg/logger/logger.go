package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	info *log.Logger
	err  *log.Logger
}

var defaultLogger *Logger

func Init() {
	defaultLogger = &Logger{
		info: log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		err:  log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}
}

func Info(v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.info.Output(2, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.info.Output(2, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.err.Output(2, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.err.Output(2, fmt.Sprintf(format, v...))
}

func InfoWithTraceID(traceID string, v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.info.Output(2, fmt.Sprintf("[trace:%s] %s", traceID, fmt.Sprint(v...)))
}

func ErrorWithTraceID(traceID string, v ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.err.Output(2, fmt.Sprintf("[trace:%s] %s", traceID, fmt.Sprint(v...)))
}
