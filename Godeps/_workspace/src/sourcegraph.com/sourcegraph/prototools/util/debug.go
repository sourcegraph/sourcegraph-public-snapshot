// +build debug

package util

import (
	"fmt"
	"os"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.Create(os.Args[0] + ".log")
	if err != nil {
		panic(err)
	}
}

func Debugf(f string, args ...interface{}) {
	fmt.Fprintf(logFile, f, args...)
}
