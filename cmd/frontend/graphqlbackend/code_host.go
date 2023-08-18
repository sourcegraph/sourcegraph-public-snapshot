package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type codeHostResolver struct {
	ch *types.CodeHost
	db database.DB
}

func (r *codeHostResolver) ID() graphql.ID {
	return MarshalCodeHostID(r.ch.ID)
}

func (r *codeHostResolver) Kind() string {
	return r.ch.Kind
}

func (r *codeHostResolver) URL() string {
	return r.ch.URL
}

func (r *codeHostResolver) APIRateLimitQuota() *int32 {
	return r.ch.APIRateLimitQuota
}

func (r *codeHostResolver) APIRateLimitIntervalSeconds() *int32 {
	return r.ch.APIRateLimitIntervalSeconds
}

func (r *codeHostResolver) GitRateLimitQuota() *int32 {
	return r.ch.GitRateLimitQuota
}

func (r *codeHostResolver) GitRateLimitIntervalSeconds() *int32 {
	return r.ch.GitRateLimitIntervalSeconds
}

type CodeHostExternalServicesArgs struct {
	First int32
	After *string
}

func (r *codeHostResolver) ExternalServices(args *CodeHostExternalServicesArgs) (*externalServiceConnectionResolver, error) {
	// 🚨 SECURITY: This may only be returned to site-admins, but code host is
	// only accessible to admins anyways.

	var afterID int64
	if args.After != nil {
		var err error
		afterID, err = UnmarshalExternalServiceID(graphql.ID(*args.After))
		if err != nil {
			return nil, err
		}
	}

	opt := database.ExternalServicesListOptions{
		// Only return services of this code host.
		CodeHostID:  r.ch.ID,
		AfterID:     afterID,
		LimitOffset: &database.LimitOffset{Limit: int(args.First)},
	}
	return &externalServiceConnectionResolver{db: r.db, opt: opt}, nil
}
