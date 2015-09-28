package base

import (
	"github.com/golang/glog"
)

// Implements a very simple timer for profiling regions of code.  Intended use is with defer, for example
//  	defer NewOneShotTimer("code region").Fire()
//  	defer NewOneShotTimer("code region").FireToLog(glog.V(2))
//
// You may also specify a minimum threshold before the region will be logged:
//  	defer NewOneShotTimerWithThreshold("code region", 10 * base.MICROS_PER_MS).FireToLog(glog.V(0))
type OneShotTimer struct {
	startTime     Micros
	fireThreshold Micros
	label         string
}

func NewOneShotTimer(label string) *OneShotTimer {
	return NewOneShotTimerWithThreshold(label, 0)
}

func NewOneShotTimerWithThreshold(label string, thresholdMicros Micros) *OneShotTimer {
	return &OneShotTimer{
		label:         label,
		startTime:     NowMicros(),
		fireThreshold: thresholdMicros,
	}
}

func (p *OneShotTimer) Fire() {
	p.FireToLog(glog.V(0))
}

func (p *OneShotTimer) FireToLog(log glog.Verbose) {
	if log {
		elapsed := NowMicros() - p.startTime
		if elapsed > p.fireThreshold {
			log.Infof("%v took %v", p.label, elapsed.ToDuration())
		}
	}
}
