package store

import "context"

// Channel defines the interface for a simple interprocess
// communication channel, implemented by PostgreSQL's LISTEN/NOTIFY
// (http://www.postgresql.org/docs/current/static/sql-notify.html).
type Channel interface {
	// Listen returns a channel that receives notifications on the
	// given channel. See
	// https://godoc.org/github.com/lib/pq#hdr-Notifications for
	// implementation details.
	//
	// Callers must call unlisten when they are done using ch, to free
	// resources.
	Listen(ctx context.Context, channel string) (ch <-chan ChannelNotification, unlisten func(), err error)

	// Notify sends a notification on the named channel. See
	// http://www.postgresql.org/docs/current/static/sql-notify.html.
	Notify(ctx context.Context, channel, payload string) error
}

// A ChannelNotification is a notification sent on a Channel.
type ChannelNotification struct {
	Channel string // the channel on which this notification was sent
	Payload string // the payload of the notification (possibly empty)
}
