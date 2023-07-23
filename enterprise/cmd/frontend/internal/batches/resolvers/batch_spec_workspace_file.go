package resolvers

import (
	"context"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const workspaceFileIDKind = "BatchSpecWorkspaceFile"

func marshalWorkspaceFileRandID(id string) graphql.ID {
	return relay.MarshalID(workspaceFileIDKind, id)
}

func unmarshalWorkspaceFileRandID(id graphql.ID) (batchWorkspaceFileRandID string, err error) {
	err = relay.UnmarshalSpec(id, &batchWorkspaceFileRandID)
	return
}

var _ graphqlbackend.BatchWorkspaceFileResolver = &batchSpecWorkspaceFileResolver{}

type batchSpecWorkspaceFileResolver struct {
	batchSpecRandID string
	file            *btypes.BatchSpecWorkspaceFile

	/*
	 * Added this to the struct, so it's easy to mock in tests.
	 * We expect `createVirtualFile` to return an interface so it's mockable.
	 */
	createVirtualFile func(content []byte, path string) graphqlbackend.FileResolver
}

func newBatchSpecWorkspaceFileResolver(batchSpecRandID string, file *btypes.BatchSpecWorkspaceFile) *batchSpecWorkspaceFileResolver {
	return &batchSpecWorkspaceFileResolver{
		batchSpecRandID:   batchSpecRandID,
		file:              file,
		createVirtualFile: createVirtualFile,
	}
}

func createVirtualFile(content []byte, path string) graphqlbackend.FileResolver {
	fileInfo := graphqlbackend.CreateFileInfo(path, false)
	return graphqlbackend.NewVirtualFileResolver(fileInfo, func(ctx context.Context) (string, error) {
		return string(content), nil
	}, graphqlbackend.VirtualFileResolverOptions{
		// TODO: Add URL to file in webapp.
		URL: "",
	})
}

func (r *batchSpecWorkspaceFileResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalWorkspaceFileRandID(r.file.RandID)
}

func (r *batchSpecWorkspaceFileResolver) ModifiedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.file.ModifiedAt}
}

func (r *batchSpecWorkspaceFileResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.file.CreatedAt}
}

func (r *batchSpecWorkspaceFileResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.file.UpdatedAt}
}

func (r *batchSpecWorkspaceFileResolver) Name() string {
	return r.file.FileName
}

func (r *batchSpecWorkspaceFileResolver) Path() string {
	return r.file.Path
}

func (r *batchSpecWorkspaceFileResolver) IsDirectory() bool {
	// A workspace file cannot be a directory.
	return false
}

func (r *batchSpecWorkspaceFileResolver) Content(ctx context.Context, args *graphqlbackend.GitTreeContentPageArgs) (string, error) {
	return "", errors.New("not implemented")
}

func (r *batchSpecWorkspaceFileResolver) ByteSize(ctx context.Context) (int32, error) {
	return int32(r.file.Size), nil
}

func (r *batchSpecWorkspaceFileResolver) TotalLines(ctx context.Context) (int32, error) {
	// If it is a binary, return 0
	binary, err := r.Binary(ctx)
	if err != nil || binary {
		return 0, err
	}
	return int32(len(strings.Split(string(r.file.Content), "\n"))), nil
}

func (r *batchSpecWorkspaceFileResolver) Binary(ctx context.Context) (bool, error) {
	vfr := r.createVirtualFile(r.file.Content, r.file.Path)
	return vfr.Binary(ctx)
}

func (r *batchSpecWorkspaceFileResolver) RichHTML(ctx context.Context, args *graphqlbackend.GitTreeContentPageArgs) (string, error) {
	return "", errors.New("not implemented")
}

func (r *batchSpecWorkspaceFileResolver) URL(ctx context.Context) (string, error) {
	return fmt.Sprintf("/files/batch-changes/%s/%s", r.batchSpecRandID, r.file.RandID), nil
}

func (r *batchSpecWorkspaceFileResolver) CanonicalURL() string {
	return ""
}

func (r *batchSpecWorkspaceFileResolver) ChangelistURL(_ context.Context) (*string, error) {
	return nil, nil
}

func (r *batchSpecWorkspaceFileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return nil, errors.New("not implemented")
}

func (r *batchSpecWorkspaceFileResolver) Highlight(ctx context.Context, args *graphqlbackend.HighlightArgs) (*graphqlbackend.HighlightedFileResolver, error) {
	vfr := r.createVirtualFile(r.file.Content, r.file.Path)
	return vfr.Highlight(ctx, args)
}

func (r *batchSpecWorkspaceFileResolver) ToGitBlob() (*graphqlbackend.GitTreeEntryResolver, bool) {
	return nil, false
}

func (r *batchSpecWorkspaceFileResolver) ToVirtualFile() (*graphqlbackend.VirtualFileResolver, bool) {
	return nil, false
}

func (r *batchSpecWorkspaceFileResolver) ToBatchSpecWorkspaceFile() (graphqlbackend.BatchWorkspaceFileResolver, bool) {
	return r, true
}
