package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type codeHostResolver struct {
	codeHost *types.CodeHost
	warning         string
}

const codeHostIDKind = "CodeHost"

func codeHostByID(ctx context.Context, id graphql.ID) (*codeHostResolver, error) {
	// 🚨 SECURITY: Only site admins are allowed to read external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	codeHostID, err := unmarshalCodeHostID(id)
	if err != nil {
		return nil, err
	}

	codeHost, err := db.CodeHosts.GetByID(ctx, codeHostID)
	if err != nil {
		return nil, err
	}

	return &codeHostResolver{codeHost: codeHost}, nil
}

func marshalCodeHostID(id int64) graphql.ID {
	return relay.MarshalID(codeHostIDKind, id)
}

func unmarshalCodeHostID(id graphql.ID) (codeHostID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != codeHostIDKind {
		err = fmt.Errorf("expected graphql ID to have kind %q; got %q", codeHostIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &codeHostID)
	return
}

func (r *codeHostResolver) ID() graphql.ID {
	return marshalCodeHostID(r.codeHost.ID)
}

func (r *codeHostResolver) Kind() string {
	return r.codeHost.Kind
}

func (r *codeHostResolver) DisplayName() string {
	return r.codeHost.DisplayName
}

func (r *codeHostResolver) Config() JSONCString {
	return JSONCString(r.codeHost.Config)
}

func (r *codeHostResolver) CreatedAt() DateTime {
	return DateTime{Time: r.codeHost.CreatedAt}
}

func (r *codeHostResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.codeHost.UpdatedAt}
}

func (r *codeHostResolver) Warning() *string {
	if r.warning == "" {
		return nil
	}
	return &r.warning
}
