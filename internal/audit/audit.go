package audit

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
)

// Log creates an INFO log statement that will be a part of the audit log.
// The audit log records comply with the following design: an actor takes an action on an entity within a context
// actor - obtained from the context
// action - caller supplied, free form (log message)
// entity - caller supplied, free form (should describe the component that's being audited)
// context - any additional logging fields (log.Field)
func Log(logger log.Logger, ctx context.Context, record *Record) {
	var fields []log.Field
	fields = append(fields, log.String("audit", "true"))

	act := actor.FromContext(ctx)
	client := requestclient.FromContext(ctx)

	fields = append(fields, log.Object("audit.actor",
		log.String("actorUID", act.UIDString()),
		log.String("ip", ip(client)),
		log.String("X-Forwarded-For", forwardedFor(client))))
	fields = append(fields, log.String("audit.entity", record.Entity))
	fields = append(fields, record.Fields...)

	logger.Info(record.Action, fields...)
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
	// Entity is the audited entity name
	Entity string
	// Action is he audit action taking place
	Action string
	// Fields hold any additional context relevant to the action
	Fields []log.Field
}
