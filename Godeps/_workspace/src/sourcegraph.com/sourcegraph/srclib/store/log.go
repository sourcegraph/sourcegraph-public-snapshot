package store

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func vlogWriter() io.Writer {
	if v, _ := strconv.ParseBool(os.Getenv("V")); v {
		return os.Stderr
	}
	return ioutil.Discard
}

func cyan(s string) string {
	return "\x1b[36m" + s + "\x1b[0m"
}

var vlog = log.New(vlogWriter(), cyan("▶ "), log.Lmicroseconds|log.Lshortfile)
