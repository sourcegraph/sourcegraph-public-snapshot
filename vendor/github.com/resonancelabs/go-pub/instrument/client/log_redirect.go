package client

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/resonancelabs/go-pub/instrument"
)

// TODO move this to log_adapters/log?

// Interpose on golang's built-in logging package (which just barely supports
// such a thing).
//
// And by "support", I mean "we can interpose at the byte-stream level and
// decode individual log lines by string parsing". The code here is borrowed
// from golang's glog package.
func captureStandardGoLogging() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(croutonLogBridge{})
}

// logBridge provides the Write method that enables captureStandardGoLogging to
// connect Go's standard logs to crouton.
type croutonLogBridge struct{}

// Write parses the standard logging line and passes its components to the
// crouton logging machinery.
func (_ croutonLogBridge) Write(b []byte) (n int, err error) {
	// (The interesting bits below were copied from glog.go)
	var (
		file = "???"
		line = 1
		text string
	)
	// Split "d.go:23: message" into "d.go", "23", and "message".
	if parts := bytes.SplitN(b, []byte{':'}, 3); len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		text = fmt.Sprintf("bad log format: %s", b)
	} else {
		file = string(parts[0])
		text = string(parts[2][1:]) // skip leading space
		line, err = strconv.Atoi(string(parts[1]))
		if err != nil {
			text = fmt.Sprintf("bad line number: %s", b)
			line = 1
		}
	}
	instrument.Log(&instrument.LogRecord{
		FileName:   file,
		LineNumber: line,
		Level:      "I",
		Message:    text,
	})
	return len(b), nil
}

/*
func init() {
	// TODO: this *replaces* the output for the standard logger rather than
	// `tee` it -- argh! We'll need to make this optional at some point.
	captureStandardGoLogging()
}
*/
