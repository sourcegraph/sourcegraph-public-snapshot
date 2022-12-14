package blobstore_test

import (
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/blobstore/internal/blobstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestBlobstore(t *testing.T) {
	ts := httptest.NewServer(&blobstore.Service{
		DataDir:        t.TempDir(),
		Log:            logtest.Scoped(t),
		ObservationCtx: observation.TestContextTB(t),
	})
	defer ts.Close()
}
