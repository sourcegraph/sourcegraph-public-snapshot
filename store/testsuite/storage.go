package testsuite

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var (
	storageKeyName = randomKey()

	storageBucket = &sourcegraph.StorageBucket{
		AppName: "go-test",
		Name:    "go-test-bucket" + fmt.Sprint(time.Now().UnixNano()),
		Repo:    "github.com/foo/bar",
	}

	storageKey = &sourcegraph.StorageKey{
		Bucket: storageBucket, // TODO(slimsag): Bucket should not be nullable
		Key:    storageKeyName,
	}

	storageValue = fullByteRange()
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

// Storage_Get tests that Storage.Get works.
func Storage_Get(ctx context.Context, t *testing.T, s store.Storage) {
	// Test that a NotFound error is returned.
	value, err := s.Get(ctx, storageKey)
	if !os.IsNotExist(err) {
		t.Fatalf("Expected os.IsNotExist(err), got: %+v\n", err)
	}

	// Put the first object in.
	_, err = s.Put(ctx, &sourcegraph.StoragePutOp{
		Key:   *storageKey,
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get the object.
	value, err = s.Get(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, storageValue) {
		t.Fatalf("got %q expected %q\n", value, storageValue)
	}
}

// Storage_Put tests that Storage.Put works.
func Storage_Put(ctx context.Context, t *testing.T, s store.Storage) {
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
	if !reflect.DeepEqual(value, newValue) {
		t.Fatalf("got %q expected %q\n", value, newValue)
	}
}

// Storage_PutNoOverwrite tests that Storage.PutNoOverwrite works.
func Storage_PutNoOverwrite(ctx context.Context, t *testing.T, s store.Storage) {
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
	if !os.IsExist(err) {
		t.Fatalf("Expected os.IsExist(err), got: %+v\n", err)
	}
}

// Storage_Delete tests that Storage.Delete works.
func Storage_Delete(ctx context.Context, t *testing.T, s store.Storage) {
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

// Storage_TestGarbageNames tests that garbage, non-alphanumeric, bucket and app
// names are rejected by at least one method in the storage service.
func Storage_TestGarbageNames(ctx context.Context, t *testing.T, s store.Storage) {
	_, err := s.Put(ctx, &sourcegraph.StoragePutOp{
		Key: sourcegraph.StorageKey{
			Bucket: &sourcegraph.StorageBucket{
				Name:    " the bucket is a bit em&ty",
				AppName: "my-app",
			},
			Key: storageKeyName,
		},
		Value: storageValue,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err == nil {
		t.Fatal("expected error for non-compliant bucket name")
	}

}
