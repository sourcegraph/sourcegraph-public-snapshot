package glock

import "time"

type realTicker struct {
	ticker *time.Ticker
}

var _ Ticker = &realTicker{}

func (t *realTicker) Chan() <-chan time.Time {
	return t.ticker.C
}

func (t *realTicker) Stop() {
	t.ticker.Stop()
}
