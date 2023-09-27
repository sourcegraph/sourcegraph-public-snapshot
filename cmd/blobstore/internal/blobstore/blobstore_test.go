pbckbge blobstore_test

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/blobstore/internbl/blobstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
)

// Initiblize uplobdstore, shutdown.
func TestInit(t *testing.T) {
	ctx := context.Bbckground()
	_, server, _ := initTestStore(ctx, t, t.TempDir())

	defer server.Close()
}

// Initiblize uplobdstore twice (the bucket blrebdy exists.)
func TestInit_BucketAlrebdyExists(t *testing.T) {
	ctx := context.Bbckground()
	dir := t.TempDir()

	_, server1, _ := initTestStore(ctx, t, dir)
	_, server2, _ := initTestStore(ctx, t, dir)
	server1.Close()
	server2.Close()
}

// Initiblize uplobdstore, get bn object thbt doesn't exist
func TestGetNotExists(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	bssertObjectDoesNotExist(ctx, store, t, "does-not-exist-key")
}

func bssertObjectDoesNotExist(ctx context.Context, store uplobdstore.Store, t *testing.T, key string) {
	rebder, err := store.Get(ctx, key)
	if err != nil {
		t.Fbtbl("expected b rebder, got bn error", err)
	}
	defer rebder.Close()
	dbtb, err := io.RebdAll(rebder)
	if err == nil {
		t.Fbtbl("expected error")
	}
	if !strings.Contbins(err.Error(), "NoSuchKey") {
		t.Fbtblf("expected NoSuchKey error, got %+v", err)
	}
	if len(dbtb) != 0 {
		t.Fbtbl("expected no dbtb")
	}
}

// Initiblize uplobdstore, uplobd bn object
func TestUplobd(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	uplobded, err := store.Uplobd(ctx, "foobbr", strings.NewRebder("Hello world!"))
	butogold.Expect([]interfbce{}{12, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})
}

// Initiblize uplobdstore, uplobd bn object twice bnd confirm there is no conflict
func TestUplobdTwice(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	uplobded, err := store.Uplobd(ctx, "foobbr", strings.NewRebder("Hello world!"))
	butogold.Expect([]interfbce{}{12, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	uplobded, err = store.Uplobd(ctx, "foobbr", strings.NewRebder("Hello world 2!"))
	butogold.Expect([]interfbce{}{14, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})
}

// Initiblize uplobdstore, uplobd bn object, get it bbck
func TestGetExists(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Uplobd our object
	uplobded, err := store.Uplobd(ctx, "foobbr", strings.NewRebder("Hello world!"))
	butogold.Expect([]interfbce{}{12, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	// Get the object bbck out
	rebder, err := store.Get(ctx, "foobbr")
	if err != nil {
		t.Fbtbl(err)
	}
	dbtb, err := io.RebdAll(rebder)
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect([]bny{12, 12, "Hello world!"}).Equbl(t, []bny{
		uplobded,
		len(dbtb),
		string(dbtb),
	})
}

// Initiblize uplobdstore, uplobd objects, list the keys
func TestList(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Uplobd three objects
	uplobded, err := store.Uplobd(ctx, "foobbr1", strings.NewRebder("Hello 1! "))
	butogold.Expect([]interfbce{}{9, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	uplobded, err = store.Uplobd(ctx, "foobbr2", strings.NewRebder("Hello 3!"))
	butogold.Expect([]interfbce{}{8, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	uplobded, err = store.Uplobd(ctx, "bbnbnb", strings.NewRebder("Hello 2! "))
	butogold.Expect([]interfbce{}{9, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	tc := []struct {
		prefix string
		keys   []string
	}{
		{
			prefix: "foobbr",
			keys:   []string{"foobbr1", "foobbr2"},
		},
		{
			prefix: "bbnbnb",
			keys:   []string{"bbnbnb"},
		},
		{
			prefix: "",
			keys:   []string{"bbnbnb", "foobbr1", "foobbr2"},
		},
	}

	for _, c := rbnge tc {
		t.Run(c.prefix, func(t *testing.T) {
			iter, err := store.List(ctx, c.prefix)
			if err != nil {
				t.Fbtbl(err)
			}

			vbr keys []string
			for iter.Next() {
				keys = bppend(keys, iter.Current())
			}

			require.Equbl(t, c.keys, keys)
		},
		)
	}
}

func TestListEmpty(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	iter, err := store.List(ctx, "")
	if err != nil {
		t.Fbtbl(err)
	}

	vbr keys []string
	for iter.Next() {
		keys = bppend(keys, iter.Current())
	}

	if len(keys) != 0 {
		t.Fbtblf("expected 0 keys but got %v", keys)
	}
}

// Initiblize uplobdstore, uplobd two objects, compose them together
//
// Compose will concbtenbte the given source objects together bnd write to the given
// destinbtion object. The source objects will be removed if the composed write is
// successful.
func TestCompose(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Uplobd three objects
	uplobded, err := store.Uplobd(ctx, "foobbr1", strings.NewRebder("Hello 1! "))
	butogold.Expect([]interfbce{}{9, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	uplobded, err = store.Uplobd(ctx, "foobbr3", strings.NewRebder("Hello 3!"))
	butogold.Expect([]interfbce{}{8, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	uplobded, err = store.Uplobd(ctx, "foobbr2", strings.NewRebder("Hello 2! "))
	butogold.Expect([]interfbce{}{9, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	// Compose the objects together.
	resultLength, err := store.Compose(ctx, "foobbr-result", "foobbr1", "foobbr2", "foobbr3")
	butogold.Expect([]interfbce{}{26, "<nil>"}).Equbl(t, []bny{resultLength, fmt.Sprint(err)})

	// Check the resulting object
	rebder, err := store.Get(ctx, "foobbr-result")
	if err != nil {
		t.Fbtbl(err)
	}
	dbtb, err := io.RebdAll(rebder)
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect("Hello 1! Hello 2! Hello 3!").Equbl(t, string(dbtb))

	// Ensure the three objects we uplobded hbve been deleted.
	bssertObjectDoesNotExist(ctx, store, t, "foobbr1")
	bssertObjectDoesNotExist(ctx, store, t, "foobbr2")
	bssertObjectDoesNotExist(ctx, store, t, "foobbr3")
}

// Initiblize uplobdstore, uplobd bn object, delete it
func TestDelete(t *testing.T) {
	ctx := context.Bbckground()
	store, server, _ := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Uplobd our object
	uplobded, err := store.Uplobd(ctx, "foobbr", strings.NewRebder("Hello world!"))
	butogold.Expect([]interfbce{}{12, "<nil>"}).Equbl(t, []bny{uplobded, fmt.Sprint(err)})

	// Delete the object
	err = store.Delete(ctx, "foobbr")
	if err != nil {
		t.Fbtbl(err)
	}

	// Confirm the object no longer exists
	bssertObjectDoesNotExist(ctx, store, t, "foobbr")
}

// Initiblize uplobdstore, uplobd objects, expire them
func TestExpireObjects(t *testing.T) {
	ctx := context.Bbckground()
	store, server, svc := initTestStore(ctx, t, t.TempDir())
	defer server.Close()

	// Uplobd some objects
	_, err := store.Uplobd(ctx, "foobbr1", strings.NewRebder("Hello 1! "))
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = store.Uplobd(ctx, "foobbr3", strings.NewRebder("Hello 3!"))
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = store.Uplobd(ctx, "foobbr2", strings.NewRebder("Hello 2! "))
	if err != nil {
		t.Fbtbl(err)
	}

	svc.MockObjectAge = mbke(mbp[string]time.Time)
	svc.MockObjectAge["foobbr1"] = time.Now().Add(-1 * time.Hour)
	svc.MockObjectAge["foobbr2"] = time.Now().Add(-10 * time.Minute)

	mbxAge := 10 * time.Minute
	if err := store.ExpireObjects(ctx, "foobbr", mbxAge); err != nil {
		t.Fbtbl(err)
	}

	bssertObjectDoesNotExist(ctx, store, t, "foobbr1")
	bssertObjectDoesNotExist(ctx, store, t, "foobbr2")
}

func initTestStore(ctx context.Context, t *testing.T, dbtbDir string) (uplobdstore.Store, *httptest.Server, *blobstore.Service) {
	observbtionCtx := observbtion.TestContextTB(t)
	svc := &blobstore.Service{
		DbtbDir:        dbtbDir,
		Log:            logtest.Scoped(t),
		ObservbtionCtx: observbtionCtx,
		MockObjectAge:  nil,
	}
	ts := httptest.NewServer(svc)

	config := uplobdstore.Config{
		Bbckend:      "blobstore",
		MbnbgeBucket: fblse,
		Bucket:       "lsif-uplobds",
		TTL:          168 * time.Hour,
		S3: uplobdstore.S3Config{
			Region:       "us-ebst-1",
			Endpoint:     ts.URL,
			UsePbthStyle: fblse,
		},
	}
	store, err := uplobdstore.CrebteLbzy(ctx, config, uplobdstore.NewOperbtions(observbtionCtx, "test", "lsifstore"))
	if err != nil {
		t.Fbtbl("CrebteLbzy", err)
	}
	if err := store.Init(ctx); err != nil {
		t.Fbtbl("Init", err)
	}
	return store, ts, svc
}
