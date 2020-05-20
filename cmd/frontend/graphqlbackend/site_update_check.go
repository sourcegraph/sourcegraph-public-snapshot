package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/updatecheck"
)

func (r *siteResolver) UpdateCheck(ctx context.Context) (*updateCheckResolver, error) {
	// 🚨 SECURITY: Only site admins can check for updates but users may see notifications
	return &updateCheckResolver{
		last:        updatecheck.Last(),
		pending:     updatecheck.IsPending(),
		IsSiteAdmin: backend.CheckCurrentUserIsSiteAdmin(ctx) == nil,
	}, nil
}

type updateCheckResolver struct {
	last        *updatecheck.Status
	pending     bool
	IsSiteAdmin bool
	alert       Alert
}

func (r *updateCheckResolver) Pending() bool { return r.pending }

func (r *updateCheckResolver) CheckedAt() *DateTime {
	if r.last == nil {
		return nil
	}
	return &DateTime{Time: r.last.Date}
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

// Alert only triggers when the instance is either offline or severely out of date
func (r *updateCheckResolver) Alert() *Alert {
	if r.last == nil || r.last.HasUpdate() {
		return nil
	}
	alert := OutOfDateAlert(r.last.MonthsOutOfDate, r.IsSiteAdmin)

	if alert.MessageValue == "" {
		return nil
	}
	return &alert
}
