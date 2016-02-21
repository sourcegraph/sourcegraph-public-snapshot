package testsuite

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var (
	storageKeyName = randomKey()
	storageValue   = fullByteRange()
)

func fullByteRange() (v []byte) {
	for i := byte(0); i < 255; i++ {
		v = append(v, i)
	}
	return
}

func randomKey() string {
	return "my-awesome\x00\x00key" + fmt.Sprint(time.Now().UnixNano())
}

func randomBucket() *sourcegraph.StorageBucket {
	return &sourcegraph.StorageBucket{
		AppName: "go-test",
		Name:    "go-test-bucket" + fmt.Sprint(time.Now().UnixNano()),
		Repo:    "github.com/foo/bar",
	}
}

// Storage_Delete tests that Storage.Delete works.
func Storage_Delete(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()

	// Ensure delete on non-existant bucket is no-op.
	_, err := s.Delete(ctx, &sourcegraph.StorageKey{
		Bucket: storageBucket,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure delete on non-existant key is no-op.
	_, err = s.Delete(ctx, &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    randomKey(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Put three objects in.
	keys := []string{randomKey(), randomKey(), randomKey()}
	for _, key := range keys {
		_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
			Key: sourcegraph.StorageKey{
				Bucket: storageBucket,
				Key:    key,
			},
			Value: storageValue,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete the first object.
	first := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    keys[0],
	}
	_, err = s.Delete(ctx, first)
	if err != nil {
		t.Fatal(err)
	}

	// Check that it no longer exists.
	exists, err := s.Exists(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if exists.Exists {
		t.Fatal("expected deleted key to no longer exist")
	}

	// Check that two objects remain.
	list, err := s.List(ctx, &sourcegraph.StorageKey{Bucket: storageBucket})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Keys) != 2 {
		t.Fatal("expect 2 keys, found", len(list.Keys))
	}

	// Delete the entire bucket
	_, err = s.Delete(ctx, &sourcegraph.StorageKey{
		Bucket: storageBucket,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check that no objects remain.
	list, err = s.List(ctx, &sourcegraph.StorageKey{Bucket: storageBucket})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Keys) != 0 {
		t.Fatal("expect 0 keys, found", len(list.Keys))
	}
}

// Storage_Exists tests that Storage.Exists works.
func Storage_Exists(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName,
	}

	// Check that no error is returned for non-existant object.
	exists, err := s.Exists(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if exists.Exists {
		t.Fatal("expected Exists == false, got Exists == true")
	}

	// Put the first object in.
	_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check that it exists.
	exists, err = s.Exists(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if !exists.Exists {
		t.Fatal("expected Exists == true, got Exists == false")
	}
}

// Storage_List tests that Storage.List works.
func Storage_List(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName,
	}

	// Check that no error is returned for non-existant bucket.
	list, err := s.List(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Keys) != 0 {
		t.Fatalf("expected zero keys, got %q\n", list.Keys)
	}

	// Put the objects in.
	want := []string{
		"a",
		"b",
		"c",
		storageKeyName,
	}
	for _, k := range want {
		_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
			Key: sourcegraph.StorageKey{
				Bucket: storageBucket,
				Key:    k,
			},
			Value: storageValue,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Check list.
	list, err = s.List(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, list.Keys) {
		t.Fatalf("expected %q, got %q\n", want, list.Keys)
	}
}

// Storage_InvalidNames tests that invalid names are not allowed by the storage
// service.
func Storage_InvalidNames(ctx context.Context, t *testing.T, s store.Storage) {
	tests := []sourcegraph.StorageBucket{
		// Invalid bucket name tests.
		sourcegraph.StorageBucket{
			Name:    " startswithspace",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "endswithspace ",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "contains space",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar",
		},

		// Invalid app name tests.
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: " startswithspace",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "endswithspace ",
			Repo:    "src.sourcegraph.com/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "contains space",
			Repo:    "src.sourcegraph.com/foo/bar",
		},

		// Invalid repo URI tests.
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    " starts.with.space/foo/bar",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "ends.with.space/foo/bar ",
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "http://src.sourcegraph.com/foo/bar", // scheme not allowed
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar?ok=true", // query not allowed
		},
		sourcegraph.StorageBucket{
			Name:    "my-bucket",
			AppName: "my-app",
			Repo:    "src.sourcegraph.com/foo/bar#ok", // fragment not allowed
		},
	}

	for _, bucket := range tests {
		_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
			Key: sourcegraph.StorageKey{
				Bucket: &bucket,
				Key:    storageKeyName,
			},
			Value: storageValue,
		})
		if err == nil {
			t.Logf("Put Key.Bucket: %#q\n", bucket)
			t.Fatal("expected error for non-compliant bucket name")
		}
	}
}

// Storage_ValidNames tests that valid and complex names are accepted by the
// storage service.
func Storage_ValidNames(ctx context.Context, t *testing.T, s store.Storage) {
	tests := []sourcegraph.StorageKey{
		// Valid bucket name tests.
		sourcegraph.StorageKey{ // Just a normal situation.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "normal-bucket",
				AppName: "normal-app",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Bucket names may contain periods.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
			},
			Key: "normal-key",
		},

		// Valid app name tests.
		sourcegraph.StorageKey{
			Bucket: &sourcegraph.StorageBucket{
				Name:    "my-bucket",
				AppName: "core.app", // App names may contain periods.
				Repo:    "src.sourcegraph.com/foo/bar",
			},
			Key: "normal-key",
		},

		// Valid repo URI tests.
		sourcegraph.StorageKey{ // A normal repo URI.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "src.sourcegraph.com/foo/bar",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Another normal repo URI.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "github.com/foo/bar",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Repo "" is accepted as "global storage".
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "",
			},
			Key: "normal-key",
		},
		sourcegraph.StorageKey{ // Crazy null bytes etc. are escaped.
			Bucket: &sourcegraph.StorageBucket{
				Name:    "www.sourcegraph.com",
				AppName: "normal-app",
				Repo:    "src.\x00\x00example.com/\x00/bar",
			},
			Key: "normal-key",
		},
	}

	for _, sk := range tests {
		_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
			Key:   sk,
			Value: storageValue,
		})
		if err != nil {
			t.Fatalf("Put Key: %#q got error=%v\n", sk, err)
		}
	}
}
