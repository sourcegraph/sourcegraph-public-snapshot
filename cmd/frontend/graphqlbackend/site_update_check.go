package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *siteResolver) UpdateCheck(ctx context.Context) (*updateCheckResolver, error) {
	// 🚨 SECURITY: Only site admins can check for updates.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		// TODO(dax): This should return err once the site flags query is fixed for users
		return &updateCheckResolver{
			last: &database.Status{
				Error: err.Error(),
			},
		}, nil
	}

	status, isPending, err := r.db.UpdateChecks().GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &updateCheckResolver{
		last:    &status,
		pending: isPending,
	}, nil
}

type updateCheckResolver struct {
	last    *database.Status
	pending bool
}

func (r *updateCheckResolver) Pending() bool { return r.pending }

func (r *updateCheckResolver) CheckedAt() *DateTime {
	if r.last == nil {
		return nil
	}
	return &DateTime{Time: r.last.FinishedAt}
}

func (r *updateCheckResolver) ErrorMessage() *string {
	if r.last == nil || r.last.Error == "" {
		return nil
	}
	return &r.last.Error
}

func (r *updateCheckResolver) UpdateVersionAvailable() *string {
	if r.last == nil || !r.last.HasUpdate() {
		return nil
	}
	return &r.last.UpdateVersion
}
