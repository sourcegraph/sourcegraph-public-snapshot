package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *userResolver) Session(ctx context.Context) (*sessionResolver, error) {
	// 🚨 SECURITY: Only the user can view their session information, because it is retrieved from
	// the context of this request (and not persisted in a way that is queryable).
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() || actor.UID != r.user.ID {
		return nil, errors.New("unable to view session for a user other than the currently authenticated user")
	}

	var sr sessionResolver
	if actor.FromSessionCookie {
		sr.canSignOut = r.user.ExternalID == nil ||
			(r.user.ExternalID != nil && (conf.AuthProvider().Openidconnect != nil || (conf.EnhancedSAMLEnabled() && conf.AuthProvider().Saml != nil)))
	}
	return &sr, nil
}

type sessionResolver struct {
	canSignOut bool
}

func (r *sessionResolver) CanSignOut() bool { return r.canSignOut }
