package trace

import (
	"encoding/json"
	"time"

	"github.com/inconshreveable/log15"
)

type InsightsTracer struct {
	store EventStore
	queue chan event
}

func NewInsightsTracer(store EventStore) *InsightsTracer {
	t := &InsightsTracer{store: store, queue: make(chan event, 100)}

	go func() {
		for {
			select {
			case x := <-t.queue:
				// do a thing
				log15.Info("fromChannel", "val", x)
				encoded, err := json.Marshal(x.tags)
				if err != nil {
					log15.Error(err.Error())
				}
				log15.Info("writing to store", "name", x.name, "time", x.stamp, "tags", string(encoded))
			}
		}

	}()

	return t
}

type event struct {
	name  string
	stamp time.Time
	tags  map[string]interface{}
}

type Field struct {
	Key   string
	Value interface{}
}

func (t *InsightsTracer) Log(name string, tags ...Field) {
	r := make(map[string]interface{})
	for _, tag := range tags {
		r[tag.Key] = tag.Value
	}
	t.queue <- event{
		name:  name,
		stamp: time.Now(),
		tags:  r,
	}
}

type EventStore interface {
	DoThing()
}

type NoopEventStore struct {
}

func (n *NoopEventStore) DoThing() {
}
