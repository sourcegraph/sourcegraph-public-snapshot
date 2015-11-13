// +build !windows,!nacl,!plan9

package sgx

import (
	"log/syslog"
	"os"
)

var (
	papertrailHost = os.Getenv("SG_SYSLOG_HOST")
	usePapertrail  = papertrailHost != ""
)

func newPapertrailLogger(tag string) (LogWriter, error) {
	return syslog.Dial("udp", papertrailHost, syslog.LOG_INFO, tag)
}
