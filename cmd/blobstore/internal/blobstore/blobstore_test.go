package blobstore_test

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/blobstore/internal/blobstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

// Initialize uploadstore, shutdown.
func TestInit(t *testing.T) {
	ctx := context.Background()
	_, server, _ := initTestStore(ctx, t, t.TempDir())

	defer server.Close()
}

// Initialize uploadstore twice (the bucket already exists.)
func TestInit_BucketAlreadyExists(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	_, server1, _ := initTestStore(ctx, t, dir)
	_, server2, _ := initTestStore(ctx, t, dir)
	server1.Close()
	server2.Close()
}

// Initialize uploadstore, get an object that doesn't exist
func TestGetNotExists(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	assertObjectDoesNotExist(ctx, store, t, "does-not-exist-key")
}

func assertObjectDoesNotExist(ctx context.Context, store uploadstore.Store, t *testing.T, key string) {
	reader, err := store.Get(ctx, key)
	if err != nil {
		t.Fatal("expected a reader, got an error", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "NoSuchKey") {
		t.Fatalf("expected NoSuchKey error, got %+v", err)
	}
	if len(data) != 0 {
		t.Fatal("expected no data")
	}
}

// Initialize uploadstore, upload an object
func TestUpload(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	uploaded, err := store.Upload(ctx, "foobar", strings.NewReader("Hello world!"))
	autogold.Expect([]interface{}{12, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})
}

// Initialize uploadstore, upload an object twice and confirm there is no conflict
func TestUploadTwice(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	uploaded, err := store.Upload(ctx, "foobar", strings.NewReader("Hello world!"))
	autogold.Expect([]interface{}{12, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	uploaded, err = store.Upload(ctx, "foobar", strings.NewReader("Hello world 2!"))
	autogold.Expect([]interface{}{14, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})
}

// Initialize uploadstore, upload an object, get it back
func TestGetExists(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Upload our object
	uploaded, err := store.Upload(ctx, "foobar", strings.NewReader("Hello world!"))
	autogold.Expect([]interface{}{12, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	// Get the object back out
	reader, err := store.Get(ctx, "foobar")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Expect([]any{12, 12, "Hello world!"}).Equal(t, []any{
		uploaded,
		len(data),
		string(data),
	})
}

// Initialize uploadstore, upload objects, list the keys
func TestList(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Upload three objects
	uploaded, err := store.Upload(ctx, "foobar1", strings.NewReader("Hello 1! "))
	autogold.Expect([]interface{}{9, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	uploaded, err = store.Upload(ctx, "foobar2", strings.NewReader("Hello 3!"))
	autogold.Expect([]interface{}{8, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	uploaded, err = store.Upload(ctx, "banana", strings.NewReader("Hello 2! "))
	autogold.Expect([]interface{}{9, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	tc := []struct {
		prefix string
		keys   []string
	}{
		{
			prefix: "foobar",
			keys:   []string{"foobar1", "foobar2"},
		},
		{
			prefix: "banana",
			keys:   []string{"banana"},
		},
		{
			prefix: "",
			keys:   []string{"banana", "foobar1", "foobar2"},
		},
	}

	for _, c := range tc {
		t.Run(c.prefix, func(t *testing.T) {
			iter, err := store.List(ctx, c.prefix)
			if err != nil {
				t.Fatal(err)
			}

			var keys []string
			for iter.Next() {
				keys = append(keys, iter.Current())
			}

			require.Equal(t, c.keys, keys)
		},
		)
	}
}

func TestListEmpty(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	iter, err := store.List(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Current())
	}

	if len(keys) != 0 {
		t.Fatalf("expected 0 keys but got %v", keys)
	}
}

// Initialize uploadstore, upload two objects, compose them together
//
// Compose will concatenate the given source objects together and write to the given
// destination object. The source objects will be removed if the composed write is
// successful.
func TestCompose(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Upload three objects
	uploaded, err := store.Upload(ctx, "foobar1", strings.NewReader("Hello 1! "))
	autogold.Expect([]interface{}{9, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	uploaded, err = store.Upload(ctx, "foobar3", strings.NewReader("Hello 3!"))
	autogold.Expect([]interface{}{8, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	uploaded, err = store.Upload(ctx, "foobar2", strings.NewReader("Hello 2! "))
	autogold.Expect([]interface{}{9, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	// Compose the objects together.
	resultLength, err := store.Compose(ctx, "foobar-result", "foobar1", "foobar2", "foobar3")
	autogold.Expect([]interface{}{26, "<nil>"}).Equal(t, []any{resultLength, fmt.Sprint(err)})

	// Check the resulting object
	reader, err := store.Get(ctx, "foobar-result")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Expect("Hello 1! Hello 2! Hello 3!").Equal(t, string(data))

	// Ensure the three objects we uploaded have been deleted.
	assertObjectDoesNotExist(ctx, store, t, "foobar1")
	assertObjectDoesNotExist(ctx, store, t, "foobar2")
	assertObjectDoesNotExist(ctx, store, t, "foobar3")
}

// Initialize uploadstore, upload an object, delete it
func TestDelete(t *testing.T) {
	ctx := context.Background()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Upload our object
	uploaded, err := store.Upload(ctx, "foobar", strings.NewReader("Hello world!"))
	autogold.Expect([]interface{}{12, "<nil>"}).Equal(t, []any{uploaded, fmt.Sprint(err)})

	// Delete the object
	err = store.Delete(ctx, "foobar")
	if err != nil {
		t.Fatal(err)
	}

	// Confirm the object no longer exists
	assertObjectDoesNotExist(ctx, store, t, "foobar")
}

// Initialize uploadstore, upload objects, expire them
func TestExpireObjects(t *testing.T) {
	ctx := context.Background()
	store, server, svc := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Upload some objects
	_, err := store.Upload(ctx, "foobar1", strings.NewReader("Hello 1! "))
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Upload(ctx, "foobar3", strings.NewReader("Hello 3!"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Upload(ctx, "foobar2", strings.NewReader("Hello 2! "))
	if err != nil {
		t.Fatal(err)
	}

	svc.MockObjectAge = make(map[string]time.Time)
	svc.MockObjectAge["foobar1"] = time.Now().Add(-1 * time.Hour)
	svc.MockObjectAge["foobar2"] = time.Now().Add(-10 * time.Minute)

	maxAge := 10 * time.Minute
	if err := store.ExpireObjects(ctx, "foobar", maxAge); err != nil {
		t.Fatal(err)
	}

	assertObjectDoesNotExist(ctx, store, t, "foobar1")
	assertObjectDoesNotExist(ctx, store, t, "foobar2")
}

func initTestStore(ctx context.Context, t *testing.T, dataDir string) (uploadstore.Store, *httptest.Server, *blobstore.Service) {
	observationCtx := observation.TestContextTB(t)
	svc := &blobstore.Service{
		DataDir:        dataDir,
		Log:            logtest.Scoped(t),
		ObservationCtx: observationCtx,
		MockObjectAge:  nil,
	}
	ts := httptest.NewServer(svc)

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
	return store, ts, svc
}
