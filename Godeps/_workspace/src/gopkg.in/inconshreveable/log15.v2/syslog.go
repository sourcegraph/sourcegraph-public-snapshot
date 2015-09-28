// +build !windows,!plan9

package log15

import (
	"log/syslog"
	"strings"
)

// SyslogHandler opens a connection to the system syslog daemon by calling
// syslog.New and writes all records to it.
func SyslogHandler(tag string, fmtr Format) (Handler, error) {
	wr, err := syslog.New(syslog.LOG_INFO, tag)
	return sharedSyslog(fmtr, wr, err)
}

// SyslogHandler opens a connection to a log daemon over the network and writes
// all log records to it.
func SyslogNetHandler(net, addr string, tag string, fmtr Format) (Handler, error) {
	wr, err := syslog.Dial(net, addr, syslog.LOG_INFO, tag)
	return sharedSyslog(fmtr, wr, err)
}

func sharedSyslog(fmtr Format, sysWr *syslog.Writer, err error) (Handler, error) {
	if err != nil {
		return nil, err
	}
	h := FuncHandler(func(r *Record) error {
		var syslogFn = sysWr.Info
		switch r.Lvl {
		case LvlCrit:
			syslogFn = sysWr.Crit
		case LvlError:
			syslogFn = sysWr.Err
		case LvlWarn:
			syslogFn = sysWr.Warning
		case LvlInfo:
			syslogFn = sysWr.Info
		case LvlDebug:
			syslogFn = sysWr.Debug
		}

		s := strings.TrimSpace(string(fmtr.Format(r)))
		return syslogFn(s)
	})
	return LazyHandler(&closingHandler{sysWr, h}), nil
}

func (m muster) SyslogHandler(tag string, fmtr Format) Handler {
	return must(SyslogHandler(tag, fmtr))
}

func (m muster) SyslogNetHandler(net, addr string, tag string, fmtr Format) Handler {
	return must(SyslogNetHandler(net, addr, tag, fmtr))
}
