package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp/chimputil"
)

func serveBetaSubscription(w http.ResponseWriter, r *http.Request) error {
	var sub sourcegraph.BetaRegistration
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		return err
	}

	actor := auth.ActorFromContext(r.Context())
	if actor.Email == "" {
		// User has no primary email, so use the one specified.
		if sub.Email == "" {
			return fmt.Errorf("user has no primary email address, and BetaRegistration.Email is not set")
		}
	} else {
		sub.Email = actor.Email
	}

	chimp, err := chimputil.Client()
	if err != nil {
		return err
	}
	_, err = chimp.PutListsMembers(chimputil.SourcegraphBetaListID, mailchimp.SubscriberHash(sub.Email), &mailchimp.PutListsMembersOptions{
		StatusIfNew:  "subscribed",
		EmailAddress: sub.Email,
		MergeFields: map[string]interface{}{
			"FNAME":    sub.FirstName,
			"LNAME":    sub.LastName,
			"LANGUAGE": mailchimp.Array(sub.Languages),
			"EDITOR":   mailchimp.Array(sub.Editors),
			"MESSAGE":  sub.Message,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, &sourcegraph.BetaResponse{
		EmailAddress: sub.Email,
	})
}
