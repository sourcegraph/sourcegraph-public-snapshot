package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type executionLogEntryResolver struct {
	entry workerutil.ExecutionLogEntry
}

var _ gql.ExecutionLogEntryResolver = &executionLogEntryResolver{}

func (r *executionLogEntryResolver) Key() string       { return r.entry.Key }
func (r *executionLogEntryResolver) Command() []string { return r.entry.Command }
func (r *executionLogEntryResolver) ExitCode() int32   { return int32(r.entry.ExitCode) }

func (r *executionLogEntryResolver) StartTime() gql.DateTime {
	return gql.DateTime{Time: r.entry.StartTime}
}

func (r *executionLogEntryResolver) DurationMilliseconds() int32 {
	return int32(r.entry.DurationMs)
}

func (r *executionLogEntryResolver) Out(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only site admins can view executor log contents.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err != backend.ErrMustBeSiteAdmin {
			return "", err
		}

		return "", nil
	}

	return r.entry.Out, nil
}
