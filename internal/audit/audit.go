package audit

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
)

// Log creates an INFO log statement that will be a part of the audit log.
// The audit log records comply with the following design: an actor takes an action on an entity within a context.
// Refer to Record struct to see details about individual components.
func Log(ctx context.Context, logger log.Logger, record Record) {
	var fields []log.Field

	client := requestclient.FromContext(ctx)

	fields = append(fields, log.Object("audit",
		log.String("entity", record.Entity),
		log.Object("actor",
			log.String("actorUID", actorId(actor.FromContext(ctx))),
			log.String("ip", ip(client)),
			log.String("X-Forwarded-For", forwardedFor(client)))))
	fields = append(fields, record.Fields...)

	logger.Info(record.Action, fields...)
}

func actorId(act *actor.Actor) string {
	if act.UID > 0 {
		return act.UIDString()
	}
	if act.AnonymousUID != "" {
		return act.AnonymousUID
	}
	return "unknown"
}

func ip(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.IP
}

func forwardedFor(client *requestclient.Client) string {
	if client == nil {
		return "unknown"
	}
	return client.ForwardedFor
}

type Record struct {
	// Entity is the name of the audited entity
	Entity string
	// Action describes the state change relevant to the audit log
	Action string
	// Fields hold any additional context relevant to the Action
	Fields []log.Field
}
