package tag

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

const (
	// If the owner of an external service has this tag, the service is allowed to sync private code
	AllowUserExternalServicePrivate = "AllowUserExternalServicePrivate"
	// If the owner of an external service has this tag, the service is allowed to sync public code only
	AllowUserExternalServicePublic = "AllowUserExternalServicePublic"
)

// CheckUserHasTag reports whether the context actor has the given tag.
// If not, it returns false and a nil error.
func CheckUserHasTag(ctx context.Context, id int32, tag string) (bool, error) {
	user, err := db.Users.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	for _, t := range user.Tags {
		if t == tag {
			return true, nil
		}
	}
	return false, nil
}

// CheckActorHasTag reports whether the context actor has the given tag.
// If not, it returns false and a nil error.
func CheckActorHasTag(ctx context.Context, tag string) (bool, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, nil
	}

	return CheckUserHasTag(ctx, a.UID, tag)
}
