package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockFileResolver struct {
	mock.Mock
}

func (m *mockFileResolver) Path() string      { return "" }
func (m *mockFileResolver) Name() string      { return "" }
func (r *mockFileResolver) IsDirectory() bool { return false }
func (m *mockFileResolver) Binary(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
func (m *mockFileResolver) ByteSize(ctx context.Context) (int32, error) {
	return 0, errors.New("not implemented")
}
func (m *mockFileResolver) TotalLines(ctx context.Context) (int32, error) {
	return 0, errors.New("not implemented")
}
func (m *mockFileResolver) Content(ctx context.Context, args *graphqlbackend.GitTreeContentPageArgs) (string, error) {
	return "", errors.New("not implemented")
}
func (m *mockFileResolver) RichHTML(ctx context.Context, args *graphqlbackend.GitTreeContentPageArgs) (string, error) {
	return "", errors.New("not implemented")
}
func (m *mockFileResolver) URL(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockFileResolver) CanonicalURL() string { return "" }

func (r *mockFileResolver) ChangelistURL(_ context.Context) (*string, error) {
	return nil, nil
}

func (m *mockFileResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return nil, errors.New("not implemented")
}
func (m *mockFileResolver) Highlight(ctx context.Context, highlightArgs *graphqlbackend.HighlightArgs) (*graphqlbackend.HighlightedFileResolver, error) {
	args := m.Called(ctx, highlightArgs)
	if args.Get(0) != nil {
		return args.Get(0).(*graphqlbackend.HighlightedFileResolver), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockFileResolver) ToGitBlob() (*graphqlbackend.GitTreeEntryResolver, bool) {
	return nil, false
}
func (m *mockFileResolver) ToVirtualFile() (*graphqlbackend.VirtualFileResolver, bool) {
	return nil, false
}
func (m *mockFileResolver) ToBatchSpecWorkspaceFile() (graphqlbackend.BatchWorkspaceFileResolver, bool) {
	return nil, false
}

func (m *mockFileResolver) Languages(context.Context) ([]string, error) {
	return nil, errors.New("not implemented")
}

func TestBatchSpecWorkspaceFileResolver(t *testing.T) {
	date := time.Date(2022, 1, 2, 3, 5, 6, 0, time.UTC)
	batchSpecRandID := "123abc"
	file := &btypes.BatchSpecWorkspaceFile{
		RandID:     "987xyz",
		FileName:   "hello.txt",
		Path:       "foo/bar",
		Size:       12,
		Content:    []byte("hello world!"),
		ModifiedAt: date,
		CreatedAt:  date,
		UpdatedAt:  date,
	}

	t.Run("non binary file", func(t *testing.T) {
		var ctx = context.Background()
		var highlightResolver = &graphqlbackend.HighlightedFileResolver{}
		var highlightArgs = &graphqlbackend.HighlightArgs{}

		resolver := &batchSpecWorkspaceFileResolver{
			batchSpecRandID: batchSpecRandID,
			file:            file,
			createVirtualFile: func(content []byte, path string) graphqlbackend.FileResolver {
				fileResolver := new(mockFileResolver)

				fileResolver.On("Binary", ctx).Return(false, nil)
				fileResolver.On("Highlight", ctx, highlightArgs).Return(highlightResolver, nil)
				return fileResolver
			},
		}

		tests := []struct {
			name        string
			getActual   func() (interface{}, error)
			expected    interface{}
			expectedErr error
		}{
			{
				name: "ID",
				getActual: func() (interface{}, error) {
					return resolver.ID(), nil
				},
				expected: graphql.ID("QmF0Y2hTcGVjV29ya3NwYWNlRmlsZToiOTg3eHl6Ig=="),
			},
			{
				name: "Name",
				getActual: func() (interface{}, error) {
					return resolver.Name(), nil
				},
				expected: file.FileName,
			},
			{
				name: "Path",
				getActual: func() (interface{}, error) {
					return resolver.Path(), nil
				},
				expected: file.Path,
			},
			{
				name: "ByteSize",
				getActual: func() (interface{}, error) {
					return resolver.ByteSize(context.Background())
				},
				expected: int32(file.Size),
			},
			{
				name: "ModifiedAt",
				getActual: func() (interface{}, error) {
					return resolver.ModifiedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "CreatedAt",
				getActual: func() (interface{}, error) {
					return resolver.CreatedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "UpdatedAt",
				getActual: func() (interface{}, error) {
					return resolver.UpdatedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "IsDirectory",
				getActual: func() (interface{}, error) {
					return resolver.IsDirectory(), nil
				},
				expected: false,
			},
			{
				name: "Content",
				getActual: func() (interface{}, error) {
					return resolver.Content(context.Background(), &graphqlbackend.GitTreeContentPageArgs{})
				},
				expected:    "",
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "Binary",
				getActual: func() (interface{}, error) {
					return resolver.Binary(ctx)
				},
				expected: false,
			},
			{
				name: "RichHTML",
				getActual: func() (interface{}, error) {
					return resolver.RichHTML(context.Background(), &graphqlbackend.GitTreeContentPageArgs{})
				},
				expected:    "",
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "URL",
				getActual: func() (interface{}, error) {
					return resolver.URL(context.Background())
				},
				expected: fmt.Sprintf("/files/batch-changes/%s/%s", batchSpecRandID, file.RandID),
			},
			{
				name: "CanonicalURL",
				getActual: func() (interface{}, error) {
					return resolver.CanonicalURL(), nil
				},
				expected: "",
			},
			{
				name: "ExternalURLs",
				getActual: func() (interface{}, error) {
					return resolver.ExternalURLs(context.Background())
				},
				expected:    []*externallink.Resolver(nil),
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "Highlight",
				getActual: func() (interface{}, error) {
					return resolver.Highlight(ctx, highlightArgs)
				},
				expected: highlightResolver,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				actual, err := test.getActual()
				if test.expectedErr != nil {
					assert.ErrorContains(t, err, test.expectedErr.Error())
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expected, actual)
			})
		}
	})

	t.Run("binary file", func(t *testing.T) {
		var ctx = context.Background()
		var highlightArgs = &graphqlbackend.HighlightArgs{}
		var highlightErr = errors.New("cannot highlight binary file")

		resolver := &batchSpecWorkspaceFileResolver{
			batchSpecRandID: batchSpecRandID,
			file:            file,
			createVirtualFile: func(content []byte, path string) graphqlbackend.FileResolver {
				fileResolver := new(mockFileResolver)

				fileResolver.On("Binary", ctx).Return(true, nil)
				fileResolver.On("Highlight", ctx, highlightArgs).Return(nil, highlightErr)
				return fileResolver
			},
		}

		tests := []struct {
			name        string
			getActual   func() (interface{}, error)
			expected    interface{}
			expectedErr error
		}{
			{
				name: "ID",
				getActual: func() (interface{}, error) {
					return resolver.ID(), nil
				},
				expected: graphql.ID("QmF0Y2hTcGVjV29ya3NwYWNlRmlsZToiOTg3eHl6Ig=="),
			},
			{
				name: "Name",
				getActual: func() (interface{}, error) {
					return resolver.Name(), nil
				},
				expected: "hello.txt",
			},
			{
				name: "Path",
				getActual: func() (interface{}, error) {
					return resolver.Path(), nil
				},
				expected: "foo/bar",
			},
			{
				name: "ByteSize",
				getActual: func() (interface{}, error) {
					return resolver.ByteSize(context.Background())
				},
				expected: int32(12),
			},
			{
				name: "ModifiedAt",
				getActual: func() (interface{}, error) {
					return resolver.ModifiedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "CreatedAt",
				getActual: func() (interface{}, error) {
					return resolver.CreatedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "UpdatedAt",
				getActual: func() (interface{}, error) {
					return resolver.UpdatedAt(), nil
				},
				expected: gqlutil.DateTime{Time: date},
			},
			{
				name: "IsDirectory",
				getActual: func() (interface{}, error) {
					return resolver.IsDirectory(), nil
				},
				expected: false,
			},
			{
				name: "Content",
				getActual: func() (interface{}, error) {
					return resolver.Content(context.Background(), &graphqlbackend.GitTreeContentPageArgs{})
				},
				expected:    "",
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "Binary",
				getActual: func() (interface{}, error) {
					return resolver.Binary(ctx)
				},
				expected: true,
			},
			{
				name: "RichHTML",
				getActual: func() (interface{}, error) {
					return resolver.RichHTML(context.Background(), &graphqlbackend.GitTreeContentPageArgs{})
				},
				expected:    "",
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "URL",
				getActual: func() (interface{}, error) {
					return resolver.URL(context.Background())
				},
				expected: fmt.Sprintf("/files/batch-changes/%s/%s", batchSpecRandID, file.RandID),
			},
			{
				name: "CanonicalURL",
				getActual: func() (interface{}, error) {
					return resolver.CanonicalURL(), nil
				},
				expected: "",
			},
			{
				name: "ExternalURLs",
				getActual: func() (interface{}, error) {
					return resolver.ExternalURLs(context.Background())
				},
				expected:    []*externallink.Resolver(nil),
				expectedErr: errors.New("not implemented"),
			},
			{
				name: "Highlight",
				getActual: func() (interface{}, error) {
					return resolver.Highlight(ctx, highlightArgs)
				},
				expected:    (*graphqlbackend.HighlightedFileResolver)(nil),
				expectedErr: highlightErr,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				actual, err := test.getActual()
				if test.expectedErr != nil {
					assert.ErrorContains(t, err, test.expectedErr.Error())
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expected, actual)
			})
		}
	})
}
