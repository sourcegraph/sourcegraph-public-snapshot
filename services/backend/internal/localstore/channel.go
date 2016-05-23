package localstore

import (
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

// channel is a DB-backed implementation of the Channel store.
type channel struct{}

var _ store.Channel = (*channel)(nil)

func (channel) Listen(ctx context.Context, channel string) (ch <-chan store.ChannelNotification, unlisten func(), err error) {
	l := pq.NewListener(appDataSource(ctx), 200*time.Millisecond, time.Second, nil)

	unlisten = func() {
		if err := l.Close(); err != nil {
			log15.Warn("Error in unlisten() call to close pgsql listener.", "channel", channel, "err", err)
		}
	}

	if err := l.Listen(channel); err != nil {
		unlisten()
		return nil, nil, err
	}

	chRW := make(chan store.ChannelNotification, 1000)
	go func() {
		for n := range l.Notify {
			chRW <- store.ChannelNotification{Channel: n.Channel, Payload: n.Extra}
		}
		close(chRW)
	}()
	ch = chRW

	return ch, unlisten, nil
}

func (channel) Notify(ctx context.Context, channel, payload string) error {
	_, err := appDBH(ctx).Exec("SELECT pg_notify($1, $2);", channel, payload)
	return err
}

// appListener returns the data source string used to connect to the
// app database.
//
// HACK: Only open one listener per *dbutil2.Handle.
func appDataSource(ctx context.Context) string {
	return ctx.Value(appDBHKey).(*dbutil2.Handle).DataSource
}
