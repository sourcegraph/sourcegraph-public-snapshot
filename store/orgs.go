package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

type Orgs interface {
	Get(context.Context, sourcegraph.OrgSpec) (*sourcegraph.Org, error)
	List(context.Context, sourcegraph.UserSpec, *sourcegraph.ListOptions) ([]*sourcegraph.Org, error)
	ListMembers(context.Context, sourcegraph.OrgSpec, *sourcegraph.OrgListMembersOptions) ([]*sourcegraph.User, error)
}
