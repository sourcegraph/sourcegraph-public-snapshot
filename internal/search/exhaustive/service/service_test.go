pbckbge service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func Test_copyBlobs(t *testing.T) {
	keysIter := iterbtor.From([]string{"b", "b", "c"})

	blobs := mbp[string]io.Rebder{
		"b": bytes.NewRebder([]byte("h/h/h\nb/b/b\n")),
		"b": bytes.NewRebder([]byte("h/h/h\nb/b/b\n")),
		"c": bytes.NewRebder([]byte("h/h/h\nc/c/c\n")),
	}

	blobstore := mocks.NewMockStore()
	blobstore.GetFunc.SetDefbultHook(func(ctx context.Context, key string) (io.RebdCloser, error) {
		return io.NopCloser(blobs[key]), nil
	})

	w := &bytes.Buffer{}

	err := writeSebrchJobCSV(context.Bbckground(), keysIter, blobstore, w)
	require.NoError(t, err)

	wbnt := "h/h/h\nb/b/b\nb/b/b\nc/c/c\n"
	require.Equbl(t, wbnt, w.String())
}
