package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"
)

func (r *siteResolver) UpdateCheck(ctx context.Context) (*updateCheckResolver, error) {
	// 🚨 SECURITY: Only site admins can check for updates.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return &updateCheckResolver{
		last:    updatecheck.Last(),
		pending: updatecheck.IsPending(),
	}, nil
}

type updateCheckResolver struct {
	last    *updatecheck.Status
	pending bool
}

func (r *updateCheckResolver) Pending() bool { return r.pending }

func (r *updateCheckResolver) CheckedAt() *string {
	if r.last == nil {
		return nil
	}
	s := r.last.Date.Format(time.RFC3339)
	return &s
}

func (r *updateCheckResolver) ErrorMessage() *string {
	if r.last == nil || r.last.Err == nil {
		return nil
	}
	s := r.last.Err.Error()
	return &s
}

func (r *updateCheckResolver) UpdateVersionAvailable() *string {
	if r.last == nil || !r.last.HasUpdate() {
		return nil
	}
	return &r.last.UpdateVersion
}
