package resolvers

import (
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newRealDashboardID(arg int64) dashboardID {
	return newDashboardID("custom", arg)
}

func newDashboardID(idType string, arg int64) dashboardID {
	return dashboardID{
		IdType: idType,
		Arg:    arg,
	}
}

const dashboardKind = "dashboard"

// dashboardID represents a GraphQL ID for insight dashboards. Each of these IDs have a sub-type (case-insensitive) to identify
// subcategories of dashboards. The argument is the ID associated with the sub-category of dashboard, if relevant.
type dashboardID struct {
	IdType string
	Arg    int64
}

func (id dashboardID) isVirtualized() bool {
	return id.isUser() || id.isOrg()
}

func (id dashboardID) isUser() bool {
	return strings.EqualFold(id.IdType, "user")
}

func (id dashboardID) isOrg() bool {
	return strings.EqualFold(id.IdType, "organization")
}

func (id dashboardID) isReal() bool {
	return strings.EqualFold(id.IdType, "custom")
}

func unmarshalDashboardID(id graphql.ID) (dashboardID, error) {
	var dbid dashboardID
	err := relay.UnmarshalSpec(id, &dbid)
	if err != nil {
		return dashboardID{}, errors.Wrap(err, "unmarshalDashboardID")
	}
	return dbid, nil
}

func (id dashboardID) marshal() graphql.ID {
	return relay.MarshalID(dashboardKind, id)
}
