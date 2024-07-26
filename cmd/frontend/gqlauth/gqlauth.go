package gqlauth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GraphQLRole string

const (
	UserGraphQLRole      GraphQLRole = "USER"
	SiteAdminGraphQLRole GraphQLRole = "SITE_ADMIN"
)

type HasRoleDirective struct {
	Role GraphQLRole
}

func (h *HasRoleDirective) ImplementsDirective() string {
	return "hasRole"
}

type contextKey int

const GraphQLRoleKey contextKey = iota

func (h *HasRoleDirective) Validate(ctx context.Context, _ interface{}) error {
	if h.Role != SiteAdminGraphQLRole {
		return nil
	}
	if role, ok := ctx.Value(GraphQLRoleKey).(GraphQLRole); ok && role == SiteAdminGraphQLRole {
		return nil
	}
	return errors.New("Operation requires site-admin permissions")
}
