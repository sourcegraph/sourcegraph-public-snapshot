package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
)

func TestBackendFake(t *testing.T) {
	assert := require.New(t)

	ctx := context.Background()
	searcher, err := service.NewSearcherFake().NewSearch(ctx, "1@rev1 1@rev2 2@rev3")
	assert.NoError(err)

	// Test RepositoryRevSpecs
	refSpecs, err := searcher.RepositoryRevSpecs(ctx)
	assert.NoError(err)
	assert.Equal("RepositoryRevSpec{1@spec} RepositoryRevSpec{2@spec}", joinStringer(refSpecs))

	// Test ResolveRepositoryRevSpec
	{
		// RepositoryRevSpec{1@spec}
		repoRevs, err := searcher.ResolveRepositoryRevSpec(ctx, refSpecs[0])
		assert.NoError(err)
		assert.Equal("RepositoryRevision{1@rev1} RepositoryRevision{1@rev2}", joinStringer(repoRevs))

		// RepositoryRevSpec{2@spec}
		repoRevs, err = searcher.ResolveRepositoryRevSpec(ctx, refSpecs[1])
		assert.NoError(err)
		assert.Equal("RepositoryRevision{2@rev3}", joinStringer(repoRevs))
	}

	// Test Search
	var csv csvBuffer
	for _, refSpec := range refSpecs {
		repoRevs, err := searcher.ResolveRepositoryRevSpec(ctx, refSpec)
		assert.NoError(err)
		for _, repoRev := range repoRevs {
			err := searcher.Search(ctx, repoRev, &csv)
			assert.NoError(err)
		}
	}
	assert.Equal(`repo,revspec,revision
1,spec,rev1
1,spec,rev2
2,spec,rev3
`, csv.buf.String())
}

func joinStringer[T fmt.Stringer](xs []T) string {
	var parts []string
	for _, x := range xs {
		parts = append(parts, x.String())
	}
	return strings.Join(parts, " ")
}

type csvBuffer struct {
	buf    bytes.Buffer
	header []string
}

func (c *csvBuffer) WriteHeader(header ...string) error {
	if c.header == nil {
		c.header = header
		return c.WriteRow(header...)
	}
	if !slices.Equal(c.header, header) {
		return errors.New("different header passed to WriteHeader")
	}
	return nil
}

func (c *csvBuffer) WriteRow(row ...string) error {
	if len(row) != len(c.header) {
		return errors.New("row size does not match header size in WriteRow")
	}
	_, err := c.buf.WriteString(strings.Join(row, ",") + "\n")
	return err
}
