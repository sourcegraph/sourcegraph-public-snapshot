package watchdog

import (
	"time"

	"github.com/derision-test/glock"
)

type WatchDog struct {
	ticker   glock.Ticker
	callback func()
	done     chan struct{}
}

func New(interval time.Duration, callback func()) *WatchDog {
	ticker := glock.NewRealTicker(interval)
	return &WatchDog{
		ticker:   ticker,
		callback: callback,
		done:     make(chan struct{}, 1),
	}
}

func (w *WatchDog) Stop() {
	close(w.done)
}

func (w *WatchDog) Start() {
	for {
		select {
		case <-w.ticker.Chan():
			go w.callback()
		case <-w.done:
			w.ticker.Stop()
			return
		}
	}
}
