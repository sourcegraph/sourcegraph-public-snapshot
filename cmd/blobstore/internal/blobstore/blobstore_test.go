package blobstore_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/blobstore/internal/blobstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

// Tests that initializing an uploadstore with blobstore as the backend works (performing no operations.)
func TestInit(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	_ = store
}

// Tests that getting an object that does not exist works.
func TestGetNotExists(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	reader, err := store.Get(ctx, "does-not-exist-key")
	if err != nil {
		t.Fatal("expected a reader, got an error", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "NotFound") {
		t.Fatalf("expected NotFound error, got %+v", err)
	}
	if len(data) != 0 {
		t.Fatal("expected no data")
	}
}

// Tests uploading an object works.
func TestUpload(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	// TODO(blobstore): call store.Upload(ctx context.Context, key string, r io.Reader) (int64, error)
	_ = store
}

// Tests uploading an object and getting it back works.
func TestGetExists(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	// TODO(blobstore): call store.Get(ctx context.Context, key string) (io.ReadCloser, error)
	_ = store
}

// Tests uploading two objects and then composing them together works.
func TestCompose(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	// TODO(blobstore): call store.Compose(ctx context.Context, destination string, sources ...string) (int64, error)
	//
	// Compose will concatenate the given source objects together and write to the given
	// destination object. The source objects will be removed if the composed write is
	// successful.
	_ = store
}

// Tests deleting an object works.
func TestDelete(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	// TODO(blobstore): call store.Delete(ctx context.Context, key string) error
	_ = store
}

// Tests expiring objects works.
func TestExpireObjects(t *testing.T) {
	ctx := context.Background()
	store, server := initTestStore(ctx, t)
	defer server.Close()

	// TODO(blobstore): call store.ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error
	_ = store
}

func initTestStore(ctx context.Context, t *testing.T) (uploadstore.Store, *httptest.Server) {
	observationCtx := observation.TestContextTB(t)
	ts := httptest.NewServer(&blobstore.Service{
		DataDir:        t.TempDir(),
		Log:            logtest.Scoped(t),
		ObservationCtx: observationCtx,
	})

	config := uploadstore.Config{
		Backend:      "blobstore",
		ManageBucket: false,
		Bucket:       "lsif-uploads",
		TTL:          168 * time.Hour,
		S3: uploadstore.S3Config{
			Region:       "us-east-1",
			Endpoint:     ts.URL,
			UsePathStyle: false,
		},
	}
	store, err := uploadstore.CreateLazy(ctx, config, uploadstore.NewOperations(observationCtx, "test", "lsifstore"))
	if err != nil {
		t.Fatal("CreateLazy", err)
	}
	if err := store.Init(ctx); err != nil {
		t.Fatal("Init", err)
	}
	return store, ts
}
