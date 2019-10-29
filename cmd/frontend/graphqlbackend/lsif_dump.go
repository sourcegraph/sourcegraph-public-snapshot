package graphqlbackend

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (*lsifDumpResolver, error) {
	return lsifDumpByGQLID(ctx, args.ID)
}

type lsifDumpResolver struct {
	lsifDump *types.LSIFDump
}

func lsifDumpByGQLID(ctx context.Context, id graphql.ID) (*lsifDumpResolver, error) {
	repository, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	return lsifDumpByStringID(ctx, repository, dumpID)
}

func lsifDumpByStringID(ctx context.Context, repository, id string) (*lsifDumpResolver, error) {
	path := fmt.Sprintf("/dumps/%s/%s", url.QueryEscape(repository), id)

	var lsifDump *types.LSIFDump
	if err := lsif.TraceRequestAndUnmarshalPayload(ctx, path, nil, &lsifDump); err != nil {
		return nil, err
	}

	return &lsifDumpResolver{lsifDump: lsifDump}, nil
}

func (r *lsifDumpResolver) ID() graphql.ID {
	return marshalLSIFDumpGQLID(r.lsifDump.Repository, fmt.Sprintf("%d", r.lsifDump.ID))
}

func (r *lsifDumpResolver) Repository() string   { return r.lsifDump.Repository }
func (r *lsifDumpResolver) Commit() string       { return r.lsifDump.Commit }
func (r *lsifDumpResolver) Root() string         { return r.lsifDump.Root }
func (r *lsifDumpResolver) VisibleAtTip() bool   { return r.lsifDump.VisibleAtTip }
func (r *lsifDumpResolver) UploadedAt() DateTime { return DateTime{Time: r.lsifDump.UploadedAt} }

func marshalLSIFDumpGQLID(repository, lsifDumpID string) graphql.ID {
	// Encode both repository and ID, as we need both to make the backend request
	return relay.MarshalID("LSIFDump", fmt.Sprintf("%s:%s", base64.StdEncoding.EncodeToString([]byte(repository)), lsifDumpID))
}

func unmarshalLSIFDumpGQLID(id graphql.ID) (string, string, error) {
	var raw string
	if err := relay.UnmarshalSpec(id, &raw); err != nil {
		return "", "", err
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", "", errors.New("malformed LSIF dump id")
	}

	repository, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", err
	}

	return string(repository), parts[1], nil
}
