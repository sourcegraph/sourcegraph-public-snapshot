package resolvers

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

const batchSpecMountIDKind = "BatchSpecMount"

func marshalBatchSpecMountRandID(id string) graphql.ID {
	return relay.MarshalID(batchSpecMountIDKind, id)
}

func unmarshalBatchSpecMountID(id graphql.ID) (batchSpecMountRandID string, err error) {
	err = relay.UnmarshalSpec(id, &batchSpecMountRandID)
	return
}

var _ graphqlbackend.BatchSpecMountResolver = &batchSpecMountResolver{}

type batchSpecMountResolver struct {
	batchSpecRandID string
	mount           *btypes.BatchSpecMount
}

func (r *batchSpecMountResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalBatchSpecMountRandID(r.mount.RandID)
}

func (r *batchSpecMountResolver) FileName() string {
	return r.mount.FileName
}

func (r *batchSpecMountResolver) Path() string {
	return r.mount.Path
}

func (r *batchSpecMountResolver) Size() int32 {
	// GraphQL does not support int64
	return int32(r.mount.Size)
}

func (r *batchSpecMountResolver) Modified() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.Modified}
}

func (r *batchSpecMountResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.CreatedAt}
}

func (r *batchSpecMountResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.mount.UpdatedAt}
}

func (r *batchSpecMountResolver) URL() string {
	return fmt.Sprintf(
		".api/batches/mount/%s/%s",
		string(marshalBatchSpecRandID(r.batchSpecRandID)),
		string(marshalBatchSpecMountRandID(r.mount.RandID)),
	)
}
