// +build pgsqltest

package pgsql

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
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

// TestStorage_Get tests that Storage.Get works.
func TestStorage_Get(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &storage{}
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

// TestStorage_Put tests that Storage.Put works.
func TestStorage_Put(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &storage{}
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

func TestStorage_PutNoOverwrite(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_PutNoOverwrite(ctx, t, &storage{})
}

func TestStorage_PutNoOverwriteConcurrent(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.TestStorage_PutNoOverwriteConcurrent(ctx, t, &storage{})
}

func TestStorage_Delete(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Delete(ctx, t, &storage{})
}

func TestStorage_Exists(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Exists(ctx, t, &storage{})
}

func TestStorage_List(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_List(ctx, t, &storage{})
}

func TestStorage_InvalidNames(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_InvalidNames(ctx, t, &storage{})
}

func TestStorage_ValidNames(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_ValidNames(ctx, t, &storage{})
}
