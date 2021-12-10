package monitor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/lib/pq"
)

type Monitor struct {
	inc chan string
	dec chan string
	sub map[uint64]subscriber
	m   sync.RWMutex
}

type subscriber struct {
	channel string
	ch      chan string
}

func New() *Monitor {
	return &Monitor{
		inc: make(chan string),
		dec: make(chan string),
		sub: map[uint64]subscriber{},
	}
}

func (m *Monitor) Run(ctx context.Context, dsn string) error {
	counts := map[string]int{}
	listener := pq.NewListener(dsn, time.Second, time.Minute, eventCallback)

	for {
		select {
		case notification := <-listener.Notify:
			m.dispatch(notification.Channel, notification.Extra)

		case channel := <-m.inc:
			if counts[channel]++; counts[channel] == 1 {
				if err := listener.Listen(channel); err != nil {
					return err
				}
			}

		case channel := <-m.dec:
			if counts[channel]--; counts[channel] == 0 {
				if err := listener.Unlisten(channel); err != nil {
					return err
				}
			}

		case <-time.After(time.Second * 60):
			if err := listener.Ping(); err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Monitor) dispatch(channel, payload string) {
	for _, id := range m.subscribersIDsFor(channel) {
		m.dispatchToSubscriber(id, payload)
	}
}

func (m *Monitor) dispatchToSubscriber(id uint64, payload string) {
	m.m.Lock()
	defer m.m.Unlock()

	sub, ok := m.sub[id]
	if !ok {
		return
	}

	select {
	case sub.ch <- payload:
	default:
		log15.Warn("Slow subscriber", "id", id, "channel", sub.channel)
	}
}

func (m *Monitor) subscribersIDsFor(channel string) []uint64 {
	m.m.Lock()
	defer m.m.Unlock()

	ids := make([]uint64, 0, len(m.sub))
	for id, sub := range m.sub {
		if sub.channel == channel {
			ids = append(ids, id)
		}
	}

	return ids
}

func (m *Monitor) Listen(channel string) (id uint64, _ <-chan string) {
	id = newID()
	ch := make(chan string, 128)
	subscriber := subscriber{channel: channel, ch: ch}

	m.m.Lock()
	m.sub[id] = subscriber
	m.m.Unlock()

	m.inc <- channel
	return id, ch
}

func (m *Monitor) Unlisten(id uint64) error {
	m.m.Lock()
	sub, ok := m.sub[id]
	if !ok {
		m.m.Unlock()
		return fmt.Errorf("unknown subscriber")
	}
	delete(m.sub, id)
	m.m.Unlock()

	close(sub.ch)
	m.dec <- sub.channel
	return nil
}

func eventCallback(ev pq.ListenerEventType, err error) {
	if err != nil {
		log15.Error("listener failed", "error", err)
	}
}

var lastID uint64

func newID() uint64 {
	return atomic.AddUint64(&lastID, 1)
}
