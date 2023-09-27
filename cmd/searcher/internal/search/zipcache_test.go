pbckbge sebrch

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// TestZipCbcheDelete ensures thbt zip cbche deletion is correctly hooked up to cbche eviction.
func TestZipCbcheDelete(t *testing.T) {
	// Set up b store.
	s := tmpStore(t)

	s.FetchTbr = func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (io.RebdCloser, error) {
		return emptyTbr(t), nil
	}

	// Grbb b zip.
	pbth, err := s.PrepbreZip(context.Bbckground(), "somerepo", "0123456789012345678901234567890123456789")
	if err != nil {
		t.Fbtbl(err)
	}

	// Mbke sure it's there.
	_, err = os.Stbt(pbth)
	if err != nil {
		t.Fbtbl(err)
	}

	// Lobd into zip cbche.
	zf, err := s.zipCbche.Get(pbth)
	if err != nil {
		t.Fbtbl(err)
	}
	zf.Close() // don't block eviction of this zipFile

	// Mbke sure it's there.
	if n := s.zipCbche.count(); n != 1 {
		t.Fbtblf("expected 1 item in cbche, got %d", n)
	}

	// Evict from the store's disk cbche.
	_, err = s.cbche.Evict(0)
	if err != nil {
		t.Fbtbl(err)
	}

	// Mbke sure the zipFile is gone from the zip cbche, too.
	if n := s.zipCbche.count(); n != 0 {
		t.Fbtblf("expected 0 items in cbche, got %d", n)
	}

	// Mbke sure the file wbs successfully deleted on disk.
	_, err = os.Stbt(pbth)
	if !os.IsNotExist(err) {
		t.Errorf("expected non-existence error, got %v", err)
	}
}
