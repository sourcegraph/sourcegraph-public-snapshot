package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type executionLogEntryResolver struct {
	entry            executor.ExecutionLogEntry
	siteAdminChecker SiteAdminChecker
}

func NewExecutionLogEntryResolver(siteAdminChecker SiteAdminChecker, entry executor.ExecutionLogEntry) resolverstubs.ExecutionLogEntryResolver {
	return &executionLogEntryResolver{
		entry:            entry,
		siteAdminChecker: siteAdminChecker,
	}
}

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
	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err != auth.ErrMustBeSiteAdmin {
			return "", err
		}

		return "", nil
	}

	return r.entry.Out, nil
}
