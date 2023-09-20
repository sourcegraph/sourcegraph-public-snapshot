package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
)

func TestBackendFake(t *testing.T) {
	testNewSearcher(t, context.Background(), NewSearcherFake(), newSearcherTestCase{
		Query:        "1@rev1 1@rev2 2@rev3",
		WantRefSpecs: "RepositoryRevSpec{1@spec} RepositoryRevSpec{2@spec}",
		WantRepoRevs: "RepositoryRevision{1@rev1} RepositoryRevision{1@rev2} RepositoryRevision{2@rev3}",
		WantCSV: `repo,revspec,revision
1,spec,rev1
1,spec,rev2
2,spec,rev3
`,
	})
}

type newSearcherTestCase struct {
	Query        string
	WantRefSpecs string
	WantRepoRevs string
	WantCSV      string
}

func testNewSearcher(t *testing.T, ctx context.Context, newSearcher NewSearcher, tc newSearcherTestCase) {
	assert := require.New(t)

	userID := int32(1)
	ctx = actor.WithActor(ctx, actor.FromMockUser(userID))

	searcher, err := newSearcher.NewSearch(ctx, userID, tc.Query)
	assert.NoError(err)

	// Test RepositoryRevSpecs
	refSpecs, err := searcher.RepositoryRevSpecs(ctx)
	assert.NoError(err)
	assert.Equal(tc.WantRefSpecs, joinStringer(refSpecs))

	// Test ResolveRepositoryRevSpec
	var repoRevs []types.RepositoryRevision
	for _, refSpec := range refSpecs {
		repoRevsPart, err := searcher.ResolveRepositoryRevSpec(ctx, refSpec)
		assert.NoError(err)
		repoRevs = append(repoRevs, repoRevsPart...)
	}
	assert.Equal(tc.WantRepoRevs, joinStringer(repoRevs))

	// Test Search
	var csv csvBuffer
	for _, repoRev := range repoRevs {
		err := searcher.Search(ctx, repoRev, &csv)
		assert.NoError(err)
	}
	assert.Equal(tc.WantCSV, csv.buf.String())
}

func TestWrongUser(t *testing.T) {
	assert := require.New(t)

	userID1 := int32(1)
	userID2 := int32(2)

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(userID1))

	newSearcher := FromSearchClient(client.NewStrictMockSearchClient())
	_, err := newSearcher.NewSearch(ctx, userID2, "foo")
	assert.Error(err)
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

func TestBlobstoreCSVWriter(t *testing.T) {
	// Each entry in bucket corresponds to one 1 uploaded csv file.
	var bucket [][]byte
	var keys []string

	mockStore := mocks.NewMockStore()
	mockStore.UploadFunc.SetDefaultHook(func(ctx context.Context, key string, r io.Reader) (int64, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return 0, err
		}

		bucket = append(bucket, b)
		keys = append(keys, key)

		return int64(len(b)), nil
	})

	csvWriter := NewBlobstoreCSVWriter(context.Background(), mockStore, "blob")
	csvWriter.maxBlobSizeBytes = 12

	err := csvWriter.WriteHeader("h", "h", "h") // 3 bytes (letters) + 2 bytes (commas) + 1 byte (newline) = 6 bytes
	require.NoError(t, err)
	err = csvWriter.WriteRow("a", "a", "a")
	require.NoError(t, err)
	// We expect a new file to be created here because we have reached the max blob size.
	err = csvWriter.WriteRow("b", "b", "b")
	require.NoError(t, err)

	err = csvWriter.Close()
	require.NoError(t, err)

	wantFiles := 2
	require.Equal(t, wantFiles, len(bucket))

	require.Equal(t, "blob", keys[0])
	require.Equal(t, "h,h,h\na,a,a\n", string(bucket[0]))

	require.Equal(t, "blob-2", keys[1])
	require.Equal(t, "h,h,h\nb,b,b\n", string(bucket[1]))
}
