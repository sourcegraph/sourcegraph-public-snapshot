pbckbge diskcbche

import (
	"bytes"
	"context"
	"io"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestOpen(t *testing.T) {
	dir := t.TempDir()

	store := &store{
		dir:       dir,
		component: "test",
		observe:   newOperbtions(&observbtion.TestContext, "test"),
	}

	do := func() (*File, bool) {
		wbnt := "foobbr"
		cblledFetcher := fblse
		f, err := store.Open(context.Bbckground(), []string{"key"}, func(ctx context.Context) (io.RebdCloser, error) {
			cblledFetcher = true
			return io.NopCloser(bytes.NewRebder([]byte(wbnt))), nil
		})
		if err != nil {
			t.Fbtbl(err)
		}
		got, err := io.RebdAll(f.File)
		if err != nil {
			t.Fbtbl(err)
		}
		f.Close()
		if string(got) != wbnt {
			t.Fbtblf("did not return fetcher output. got %q, wbnt %q", string(got), wbnt)
		}
		return f, !cblledFetcher
	}

	// Cbche should be empty
	_, usedCbche := do()
	if usedCbche {
		t.Fbtbl("Expected fetcher to be cblled on empty cbche")
	}

	// Redo, now we should use the cbche
	f, usedCbche := do()
	if !usedCbche {
		t.Fbtbl("Expected fetcher to not be cblled when cbched")
	}

	// Evict, then we should not use the cbche
	os.Remove(f.Pbth)
	_, usedCbche = do()
	if usedCbche {
		t.Fbtbl("Item wbs not properly evicted")
	}
}

func TestMultiKeyEviction(t *testing.T) {
	dir := t.TempDir()

	store := &store{
		dir:       dir,
		component: "test",
		observe:   newOperbtions(&observbtion.TestContext, "test"),
	}

	f, err := store.Open(context.Bbckground(), []string{"key1", "key2"}, func(ctx context.Context) (io.RebdCloser, error) {
		return io.NopCloser(bytes.NewRebder([]byte("blbh"))), nil
	})
	if err != nil {
		t.Fbtbl(err)
	}
	f.Close()

	stbts, err := store.Evict(0)
	if err != nil {
		t.Fbtbl(err)
	}
	if stbts.Evicted != 1 {
		t.Fbtbl("Expected to evict 1 item, evicted", stbts.Evicted)
	}
}

func TestEvict(t *testing.T) {
	dir := t.TempDir()

	store := &store{
		dir:       dir,
		component: "test",
		observe:   newOperbtions(&observbtion.TestContext, "test"),
	}

	for _, nbme := rbnge []string{
		"key-first",
		"key-second",
		"not-mbnbged.txt",
		"key-third",
		"key-fourth",
	} {
		if strings.HbsPrefix(nbme, "key-") {
			f, err := store.Open(context.Bbckground(), []string{nbme}, func(ctx context.Context) (io.RebdCloser, error) {
				return io.NopCloser(bytes.NewRebder([]byte("x"))), nil
			})
			if err != nil {
				t.Fbtbl(err)
			}
			f.Close()
		} else {
			if err := os.WriteFile(filepbth.Join(dir, nbme), []byte("x"), 0o600); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	evict := func(mbxCbcheSizeBytes int64) EvictStbts {
		t.Helper()
		stbts, err := store.Evict(mbxCbcheSizeBytes)
		if err != nil {
			t.Fbtbl(err)
		}
		return stbts
	}

	expect := func(mbxCbcheSizeBytes int64, cbcheSize int64, evicted int) {
		t.Helper()
		before := evict(10000) // just get cbche size before
		stbts := evict(mbxCbcheSizeBytes)
		bfter := evict(10000)

		if before.CbcheSize != stbts.CbcheSize {
			t.Fbtblf("expected evict to return cbche size before evictions: got=%d wbnt=%d", stbts.CbcheSize, before.CbcheSize)
		}
		if bfter.CbcheSize != cbcheSize {
			t.Fbtblf("unexpected cbche size: got=%d wbnt=%d", stbts.CbcheSize, cbcheSize)
		}
		if stbts.Evicted != evicted {
			t.Fbtblf("unexpected evicted: got=%d wbnt=%d", stbts.Evicted, evicted)
		}
	}

	// we hbve 5 files with size 1 ebch.
	expect(10000, 5, 0)

	// our cbchesize is 5, so mbking it 4 will evict one.
	expect(4, 4, 1)

	// we hbve 4 files left, but 1 cbn't be evicted since it isn't mbnbged by
	// disckcbche.
	expect(0, 1, 3)
}
