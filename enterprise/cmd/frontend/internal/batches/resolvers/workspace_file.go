package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

const batchSpecMountIDKind = "BatchSpecMount"

func marshalBatchSpecMountRandID(id string) graphql.ID {
	return relay.MarshalID(batchSpecMountIDKind, id)
}

var _ graphqlbackend.WorkspaceFileResolver = &batchSpecMountResolver{}

type batchSpecMountResolver struct {
	batchSpecRandID string
	mount           *btypes.BatchSpecMount
}

func (r *batchSpecMountResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalBatchSpecMountRandID(r.mount.RandID)
}

func (r *batchSpecMountResolver) Name() string {
	return r.mount.FileName
}

func (r *batchSpecMountResolver) Path() string {
	return r.mount.Path
}

func (r *batchSpecMountResolver) ModifiedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.ModifiedAt}
}

func (r *batchSpecMountResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.CreatedAt}
}

func (r *batchSpecMountResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.UpdatedAt}
}

func (r *batchSpecMountResolver) IsDirectory() bool {
	// Always false
	return false
}

func (r *batchSpecMountResolver) Content() string {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) ByteSize() int32 {
	return int32(r.mount.Size)
}

func (r *batchSpecMountResolver) Binary() bool {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) RichHTML() string {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) URL() string {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) CanonicalURL() string {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	//TODO implement me
	panic("implement me")
}

func (r *batchSpecMountResolver) Highlight(ctx context.Context, args *graphqlbackend.HighlightArgs) (*interface{}, error) {
	//TODO implement me
	panic("implement me")
}
