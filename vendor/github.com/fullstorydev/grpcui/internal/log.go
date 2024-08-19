package internal

import (
	"log"
	"time"
)

func LogErrorf(format string, args ...interface{}) {
	prefix := "ERROR: " + time.Now().Format("2006/01/02 15:04:05") + " "
	log.Printf(prefix+format, args...)
}

func LogInfof(format string, args ...interface{}) {
	prefix := "INFO: " + time.Now().Format("2006/01/02 15:04:05") + " "
	log.Printf(prefix+format, args...)
}
