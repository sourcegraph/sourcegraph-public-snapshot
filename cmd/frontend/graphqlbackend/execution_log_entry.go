package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type ExecutionLogEntryResolver interface {
	Key() string
	Command() []string
	StartTime() gqlutil.DateTime
	ExitCode() *int32
	Out(ctx context.Context) (string, error)
	DurationMilliseconds() *int32
}

func NewExecutionLogEntryResolver(db database.DB, entry workerutil.ExecutionLogEntry) *executionLogEntryResolver {
	return &executionLogEntryResolver{
		db:    db,
		entry: entry,
	}
}

type executionLogEntryResolver struct {
	db    database.DB
	entry workerutil.ExecutionLogEntry
}

var _ ExecutionLogEntryResolver = &executionLogEntryResolver{}

func (r *executionLogEntryResolver) Key() string       { return r.entry.Key }
func (r *executionLogEntryResolver) Command() []string { return r.entry.Command }

func (r *executionLogEntryResolver) ExitCode() *int32 {
	if r.entry.ExitCode == nil {
		return nil
	}
	val := int32(*r.entry.ExitCode)
	return &val
}

func (r *executionLogEntryResolver) StartTime() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.entry.StartTime}
}

func (r *executionLogEntryResolver) DurationMilliseconds() *int32 {
	if r.entry.DurationMs == nil {
		return nil
	}
	val := int32(*r.entry.DurationMs)
	return &val
}

func (r *executionLogEntryResolver) Out(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only site admins can view executor log contents.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if err != auth.ErrMustBeSiteAdmin {
			return "", err
		}

		return "", nil
	}

	return r.entry.Out, nil
}
