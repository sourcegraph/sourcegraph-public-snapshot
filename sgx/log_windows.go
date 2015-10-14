package sgx

import (
	"errors"
)

var (
	papertrailHost = ""
	usePapertrail  = false
)

func newPapertrailLogger(tag string) (LogWriter, error) {
	return nil, errors.New("syslog is not supported on current platform")
}
