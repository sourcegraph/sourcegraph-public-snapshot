package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

type trialRequestForHubSpot struct {
	Email  *string `url:"email"`
	SiteID string  `url:"site_id"`
}

// RequestTrial makes a submission to the request trial form.
func (r *schemaResolver) RequestTrial(ctx context.Context, args *struct {
	Email string
}) (*EmptyResponse, error) {
	email := args.Email

	// If user is authenticated, use their uid and overwrite the optional email field.
	if actor := actor.FromContext(ctx); actor.IsAuthenticated() {
		e, _, err := db.UserEmails.GetPrimaryEmail(ctx, actor.UID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
		if e != "" {
			email = e
		}
	}

	// Submit form to HubSpot
	if err := hubspotutil.Client().SubmitForm(hubspotutil.TrialFormID, &trialRequestForHubSpot{
		Email:  &email,
		SiteID: siteid.Get(),
	}); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
