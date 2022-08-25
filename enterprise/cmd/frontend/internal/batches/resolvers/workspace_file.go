package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const workspaceFileIDKind = "WorkspaceFile"

func marshalWorkspaceFileRandID(id string) graphql.ID {
	return relay.MarshalID(workspaceFileIDKind, id)
}

var _ graphqlbackend.BatchWorkspaceFileResolver = &workspaceFileResolver{}

type workspaceFileResolver struct {
	batchSpecRandID string
	file            *btypes.BatchSpecMount
}

func (r *workspaceFileResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalWorkspaceFileRandID(r.file.RandID)
}

func (r *workspaceFileResolver) ModifiedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.file.ModifiedAt}
}

func (r *workspaceFileResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.file.CreatedAt}
}

func (r *workspaceFileResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.file.UpdatedAt}
}

func (r *workspaceFileResolver) Name() string {
	return r.file.FileName
}

func (r *workspaceFileResolver) Path() string {
	return r.file.Path
}

func (r *workspaceFileResolver) IsDirectory() bool {
	// Always false
	return false
}

func (r *workspaceFileResolver) Content(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (r *workspaceFileResolver) ByteSize(ctx context.Context) (int32, error) {
	return int32(r.file.Size), nil
}

func (r *workspaceFileResolver) Binary(ctx context.Context) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *workspaceFileResolver) RichHTML(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (r *workspaceFileResolver) URL(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (r *workspaceFileResolver) CanonicalURL() string {
	return ""
}

func (r *workspaceFileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return nil, errors.New("not implemented")
}

func (r *workspaceFileResolver) Highlight(ctx context.Context, args *graphqlbackend.HighlightArgs) (*graphqlbackend.HighlightedFileResolver, error) {
	return nil, errors.New("not implemented")
}

func (r *workspaceFileResolver) ToGitBlob() (*graphqlbackend.GitTreeEntryResolver, bool) {
	return nil, false
}

func (r *workspaceFileResolver) ToVirtualFile() (*graphqlbackend.VirtualFileResolver, bool) {
	return nil, false
}

func (r *workspaceFileResolver) ToWorkspaceFile() (graphqlbackend.BatchWorkspaceFileResolver, bool) {
	return r, true
}
