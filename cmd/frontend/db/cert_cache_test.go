package db

import (
	"reflect"
	"testing"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"golang.org/x/crypto/acme/autocert"
)

func TestCertCache(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	// all the possible bytes to ensure we encode correctly.
	data := []byte("header")
	for c := 0; c <= 256; c++ {
		data = append(data, byte(c))
	}

	_, err := CertCache.Get(ctx, "k")
	if err != autocert.ErrCacheMiss {
		t.Fatal("expected cache miss:", err)
	}

	err = CertCache.Delete(ctx, "k")
	if err != nil {
		t.Fatal("expected deleting a cache miss to not fail:", err)
	}

	err = CertCache.Put(ctx, "k", data)
	if err != nil {
		t.Fatal("put failed:", err)
	}
	got, err := CertCache.Get(ctx, "k")
	if err != nil {
		t.Fatal("get failed:", err)
	}
	if !reflect.DeepEqual(got, data) {
		t.Fatal("did not get back the data we put")
	}

	// overwrite
	data = []byte("hello world")
	err = CertCache.Put(ctx, "k", data)
	if err != nil {
		t.Fatal("overwrite put failed:", err)
	}
	got, err = CertCache.Get(ctx, "k")
	if err != nil {
		t.Fatal("overwrite get failed:", err)
	}
	if !reflect.DeepEqual(got, data) {
		t.Fatal("overwrite did not get back the data we put")
	}

	err = CertCache.Delete(ctx, "k")
	if err != nil {
		t.Fatal("delete failed:", err)
	}

	_, err = CertCache.Get(ctx, "k")
	if err != autocert.ErrCacheMiss {
		t.Fatal("expected cache miss after delete:", err)
	}
}
