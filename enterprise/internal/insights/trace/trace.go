package trace

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/inconshreveable/log15"
)

type BackgroundDebugEventLogger struct {
	store EventStore
	queue chan event
}

func (t *BackgroundDebugEventLogger) WithDefaultFields(field ...Field) DebugEventLogger {
	return &defaultFieldLogger{
		defaults: field,
		delegate: t,
	}
}

func NewBackgroundDebugEventLogger(store EventStore) *BackgroundDebugEventLogger {
	t := &BackgroundDebugEventLogger{store: store, queue: make(chan event, 100)}

	go func() {
		for {
			select {
			case x := <-t.queue:
				// do a thing
				err := store.Write(x)
				if err != nil {
					log15.Error(errors.Wrap(err, "EventStore.Write").Error())
				}
			}
		}

	}()

	return t
}

type event struct {
	name  string
	stamp time.Time
	tags  []byte
}

type Field struct {
	Key   string
	Value interface{}
}

type DebugEventLogger interface {
	Log(name string, tags ...Field)
	WithDefaultFields(...Field) DebugEventLogger
}

type defaultFieldLogger struct {
	defaults []Field
	delegate DebugEventLogger
}

func (d *defaultFieldLogger) Log(name string, tags ...Field) {
	d.delegate.Log(name, append(tags, d.defaults...)...)
}

func (d *defaultFieldLogger) WithDefaultFields(field ...Field) DebugEventLogger {
	newFields := make([]Field, 0, len(field)+len(d.defaults))
	newFields = append(newFields, field...)
	newFields = append(d.defaults)

	return &defaultFieldLogger{
		defaults: newFields,
		delegate: d.delegate,
	}
}

func (t *BackgroundDebugEventLogger) Log(name string, tags ...Field) {
	r := make(map[string]interface{})
	for _, tag := range tags {
		r[tag.Key] = tag.Value
	}
	encoded, err := json.Marshal(r)
	if err != nil {
		log15.Error(err.Error())
		return
	}
	t.queue <- event{
		name:  name,
		stamp: time.Now(),
		tags:  encoded,
	}
}

type EventStore interface {
	Write(event) error
}

type NoopEventStore struct {
}

func (n *NoopEventStore) Write(e event) error {
	log15.Info("no-op event store", "name", e.name, "time", e.stamp, "tags", string(e.tags))
	return nil
}

type DBEventStore struct {
	*basestore.Store
	Now func() time.Time
}

func NewDBEventStore(db dbutil.DB) *DBEventStore {
	return &DBEventStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func (t *DBEventStore) Write(e event) error {
	return t.Exec(context.Background(), sqlf.Sprintf(writeEventSql, e.name, e.stamp, e.tags))
}

const writeEventSql = `
insert into debug_event_logs (name, time, tags) values (%s, %s, %s);
`

type DebugEvent []Field

var (
	seriesIdKey    string = "series_id"
	viewIdKey      string = "view_id"
	dashboardIdKey string = "dashboard_id"
)

func (d DebugEvent) WithSeriesId(seriesId string) {

}
