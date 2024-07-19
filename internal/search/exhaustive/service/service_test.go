package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/object/mocks"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

// Test_copyBlobs tests that we concatenate objects from the object store
// properly.
func Test_copyBlobs(t *testing.T) {
	keysIter := iterator.From([]string{"a", "b", "c"})

	blobs := map[string]io.Reader{
		"a": bytes.NewReader([]byte("{\"Key\":\"a\"}\n{\"Key\":\"b\"}\n")),
		"b": bytes.NewReader([]byte("{\"Key\":\"c\"}\n")),
		"c": bytes.NewReader([]byte("{\"Key\":\"d\"}\n{\"Key\":\"e\"}\n{\"Key\":\"f\"}\n")),
	}

	blobstore := mocks.NewMockStorage()
	blobstore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(blobs[key]), nil
	})

	w := &bytes.Buffer{}

	n, err := writeSearchJobJSON(context.Background(), keysIter, blobstore, w)
	require.NoError(t, err)
	require.Equal(t, int64(72), n)

	want := "{\"Key\":\"a\"}\n{\"Key\":\"b\"}\n{\"Key\":\"c\"}\n{\"Key\":\"d\"}\n{\"Key\":\"e\"}\n{\"Key\":\"f\"}\n"
	require.Equal(t, want, w.String())
}
