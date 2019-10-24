package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (*lsifDumpResolver, error) {
	return lsifDumpByGQLID(ctx, args.ID)
}

type lsifDumpResolver struct {
	lsifDump *types.LSIFDump
}

func lsifDumpByGQLID(ctx context.Context, id graphql.ID) (*lsifDumpResolver, error) {
	dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	return lsifDumpByStringID(ctx, dumpID)
}

func lsifDumpByStringID(ctx context.Context, id string) (*lsifDumpResolver, error) {
	var lsifDump *types.LSIFDump
	if err := lsifRequest(ctx, fmt.Sprintf("dumps/%s", id), nil, &lsifDump); err != nil {
		return nil, err
	}

	return &lsifDumpResolver{lsifDump: lsifDump}, nil
}

func (r *lsifDumpResolver) ID() graphql.ID {
	return marshalLSIFDumpGQLID(fmt.Sprintf("%d", r.lsifDump.ID))
}

func (r *lsifDumpResolver) Repository() string { return r.lsifDump.Repository }
func (r *lsifDumpResolver) Commit() string     { return r.lsifDump.Commit }
func (r *lsifDumpResolver) Root() string       { return r.lsifDump.Root }

func marshalLSIFDumpGQLID(lsifDumpID string) graphql.ID {
	return relay.MarshalID("LSIFDump", lsifDumpID)
}

func unmarshalLSIFDumpGQLID(id graphql.ID) (lsifDumpID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifDumpID)
	return
}
