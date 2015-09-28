package base

import (
	"fmt"

	"github.com/golang/glog"
)

// Combines a glog.Error() call with creation of an error object
//
// I.e. the following can be written instead of expanding out the
// the fmt call, the log call, and then the return:
//
// return base.LogNewErrorf("Some error with value %v", myValue)
//
func LogNewErrorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	glog.Error(err.Error())
	return err
}

// This is a largely arbitrary: the goal is often enough for the error to
// be noticable, but infrequent enough to not spam the logs
const kLogErrorfThrottle = MICROS_PER_MINUTE

var gLastErrorfForString map[string]Micros

func init() {
	gLastErrorfForString = make(map[string]Micros)
}

// Throttles an glog.Errorf() to write out the error no more than once per
// kLogErrorfThrottle.  Quick, easy solution for avoiding log spam without
// compromising a legitimate error log call.
//
func LogErrorfThrottled(f string, args ...interface{}) {
	now := NowMicros()
	last := gLastErrorfForString[f]
	if now > last+kLogErrorfThrottle {
		gLastErrorfForString[f] = now
		glog.Errorf(f, args...)
	} else {
		glog.V(1).Infof("[throttled] "+f, args...)
	}
}
