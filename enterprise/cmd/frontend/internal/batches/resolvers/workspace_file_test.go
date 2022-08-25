package resolvers

import (
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func TestBatchSpecMountResolver(t *testing.T) {
	date := time.Date(2022, 1, 2, 3, 5, 6, 0, time.UTC)

	resolver := batchSpecMountResolver{
		batchSpecRandID: "123abc",
		mount: &btypes.BatchSpecMount{
			RandID:     "987xyz",
			FileName:   "hello.txt",
			Path:       "foo/bar",
			Size:       12,
			ModifiedAt: date,
			CreatedAt:  date,
			UpdatedAt:  date,
		},
	}

	tests := []struct {
		name      string
		getActual func() interface{}
		expected  interface{}
	}{
		{
			name: "ID",
			getActual: func() interface{} {
				return resolver.ID()
			},
			expected: graphql.ID("QmF0Y2hTcGVjTW91bnQ6Ijk4N3h5eiI="),
		},
		{
			name: "Name",
			getActual: func() interface{} {
				return resolver.Name()
			},
			expected: "hello.txt",
		},
		{
			name: "Path",
			getActual: func() interface{} {
				return resolver.Path()
			},
			expected: "foo/bar",
		},
		{
			name: "ByteSize",
			getActual: func() interface{} {
				return resolver.Size()
			},
			expected: int32(12),
		},
		{
			name: "ModifiedAt",
			getActual: func() interface{} {
				return resolver.ModifiedAt()
			},
			expected: graphqlbackend.DateTime{Time: date},
		},
		{
			name: "CreatedAt",
			getActual: func() interface{} {
				return resolver.CreatedAt()
			},
			expected: graphqlbackend.DateTime{Time: date},
		},
		{
			name: "UpdatedAt",
			getActual: func() interface{} {
				return resolver.UpdatedAt()
			},
			expected: graphqlbackend.DateTime{Time: date},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.getActual()
			assert.Equal(t, test.expected, actual)
		})
	}
}
