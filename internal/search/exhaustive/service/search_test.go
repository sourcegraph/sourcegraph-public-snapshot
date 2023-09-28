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
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
)

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
