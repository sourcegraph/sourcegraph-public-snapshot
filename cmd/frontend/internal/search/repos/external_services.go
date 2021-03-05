package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	ok, _ := database.GlobalUsers.HasTag(ctx, a.UID, database.TagAllowUserExternalServicePrivate)
	if ok {
		return conf.ExternalServiceModeAll
	}

	ok, _ = database.GlobalUsers.HasTag(ctx, a.UID, database.TagAllowUserExternalServicePublic)
	if ok {
		return conf.ExternalServiceModePublic
	}

	return conf.ExternalServiceModeDisabled
}
