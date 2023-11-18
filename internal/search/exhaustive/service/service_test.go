package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func Test_copyBlobs(t *testing.T) {
	keysIter := iterator.From([]string{"a", "b", "c"})

	blobs := map[string]io.Reader{
		"a": bytes.NewReader([]byte("h/h/h\na/a/a\n")),
		"b": bytes.NewReader([]byte("h/h/h\nb/b/b\n")),
		"c": bytes.NewReader([]byte("h/h/h\nc/c/c\n")),
	}

	blobstore := mocks.NewMockStore()
	blobstore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(blobs[key]), nil
	})

	w := &bytes.Buffer{}

	n, err := writeSearchJobCSV(context.Background(), keysIter, blobstore, w)
	require.NoError(t, err)
	require.Equal(t, int64(24), n)

	want := "h/h/h\na/a/a\nb/b/b\nc/c/c\n"
	require.Equal(t, want, w.String())
}
