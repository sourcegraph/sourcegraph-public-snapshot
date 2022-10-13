package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBatchSpecWorkspaceFileResolver(t *testing.T) {
	date := time.Date(2022, 1, 2, 3, 5, 6, 0, time.UTC)

	resolver := batchSpecWorkspaceFileResolver{
		batchSpecRandID: "123abc",
		file: &btypes.BatchSpecWorkspaceFile{
			RandID:     "987xyz",
			FileName:   "hello.txt",
			Path:       "foo/bar",
			Size:       12,
			Content:    []byte("hello world!"),
			ModifiedAt: date,
			CreatedAt:  date,
			UpdatedAt:  date,
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
				return resolver.Content(context.Background())
			},
			expected:    "",
			expectedErr: errors.New("not implemented"),
		},
		{
			name: "Binary",
			getActual: func() (interface{}, error) {
				return resolver.Binary(context.Background())
			},
			expected:    false,
			expectedErr: errors.New("not implemented"),
		},
		{
			name: "RichHTML",
			getActual: func() (interface{}, error) {
				return resolver.RichHTML(context.Background())
			},
			expected:    "",
			expectedErr: errors.New("not implemented"),
		},
		{
			name: "URL",
			getActual: func() (interface{}, error) {
				return resolver.URL(context.Background())
			},
			expected:    "",
			expectedErr: errors.New("not implemented"),
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
				return resolver.Highlight(context.Background(), nil)
			},
			expected:    (*graphqlbackend.HighlightedFileResolver)(nil),
			expectedErr: errors.New("not implemented"),
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
}
