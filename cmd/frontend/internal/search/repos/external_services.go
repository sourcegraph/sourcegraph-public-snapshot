package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

func CurrentUserAllowedExternalServices(ctx context.Context) conf.ExternalServiceMode {
	mode := conf.ExternalServiceUserMode()
	if mode != conf.ExternalServiceModeDisabled {
		return mode
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return conf.ExternalServiceModeDisabled
	}

	// The user may have a tag that opts them in
	ok, _ := db.Users.HasTag(ctx, a.UID, db.TagAllowUserExternalServicePrivate)
	if ok {
		return conf.ExternalServiceModeAll
	}

	ok, _ = db.Users.HasTag(ctx, a.UID, db.TagAllowUserExternalServicePublic)
	if ok {
		return conf.ExternalServiceModePublic
	}

	return conf.ExternalServiceModeDisabled
}
