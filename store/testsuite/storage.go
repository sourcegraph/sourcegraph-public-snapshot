package testsuite

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

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

// Storage_Get tests that Storage.Get works.
func Storage_Get(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket, // TODO(slimsag): Bucket should not be nullable
		Key:    storageKeyName,
	}

	// Test that a NotFound error is returned.
	value, err := s.Get(ctx, storageKey)
	if grpc.Code(err) != codes.NotFound {
		t.Fatalf("Expected codes.NotFound, got: %+v\n", err)
	}

	// Put the first object in.
	_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Now test that NotFound is returned for a valid bucket but an invalid key.
	value, err = s.Get(ctx, &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName + "-secondary",
	})
	if grpc.Code(err) != codes.NotFound {
		t.Fatalf("(2) Expected codes.NotFound, got: %+v\n", err)
	}

	// Get the object.
	value, err = s.Get(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.Value, storageValue) {
		t.Fatalf("got %q expected %q\n", value, storageValue)
	}
}

// Storage_Put tests that Storage.Put works.
func Storage_Put(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName,
	}

	// Put the first object in.
	_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Overwrite the value.
	newValue := []byte("new value")
	_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: newValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get the object.
	value, err := s.Get(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.Value, newValue) {
		t.Fatalf("got %q expected %q\n", value, newValue)
	}
}

// Storage_PutNoOverwrite tests that Storage.PutNoOverwrite works.
func Storage_PutNoOverwrite(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName,
	}

	// Put the first object in.
	_, err := s.PutNoOverwrite(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test that overwrite returns a AlreadyExists error.
	_, err = s.PutNoOverwrite(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if grpc.Code(err) != codes.AlreadyExists {
		t.Fatalf("Expected codes.AlreadyExists, got: %+v\n", err)
	}
}

// TestStorage_PutNoOverwriteConcurrent tests that Storage.PutNoOverwrite works.
func TestStorage_PutNoOverwriteConcurrent(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()

	for attempt := 0; attempt < 4; attempt++ {
		// Spawn off a bunch of goroutines to try PutNoOverwrite; only one should
		// succeed.
		var (
			success    uint32
			wg         sync.WaitGroup
			storageKey = sourcegraph.StorageKey{
				Bucket: storageBucket,
				Key:    randomKey(),
			}
		)
		for g := 0; g < 10; g++ {
			wg.Add(1)
			go func() {
				_, err := s.PutNoOverwrite(ctx, &sourcegraph.StoragePutOp{
					Key:   storageKey,
					Value: storageValue,
				})
				if err == nil {
					atomic.AddUint32(&success, 1)
				} else if grpc.Code(err) != codes.AlreadyExists {
					t.Log("got error:", err)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		if success != 1 {
			t.Log("expected 1 success, got", success)
		} else {
			t.Log("got 1 success")
		}
	}
}

// Storage_Delete tests that Storage.Delete works.
func Storage_Delete(ctx context.Context, t *testing.T, s store.Storage) {
	storageBucket := randomBucket()
	storageKey := &sourcegraph.StorageKey{
		Bucket: storageBucket,
		Key:    storageKeyName,
	}

	// Ensure delete on non-existant bucket is no-op.
	_, err := s.Delete(ctx, &sourcegraph.StorageKey{
		Bucket: storageBucket,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure delete on non-existant key is no-op.
	_, err = s.Delete(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}

	// Put the first object in.
	_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete the object.
	_, err = s.Delete(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}

	// Check that it no longer exists.
	exists, err := s.Exists(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if exists.Exists {
		t.Fatal("expected deleted key to no longer exist")
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
