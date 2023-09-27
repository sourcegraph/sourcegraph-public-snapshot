pbckbge rubygems

import (
	"bytes"
	"context"
	"flbg"
	"os"
	"pbth/filepbth"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
)

// Run go test ./internbl/extsvc/rubygems -updbte to updbte snbpshots.
func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}

vbr updbteRecordings = flbg.Bool("updbte", fblse, "mbke npm API cblls, record bnd sbve dbtb")

func newTestHTTPClient(t *testing.T) (client *Client, stop func()) {
	t.Helper()
	recorderFbctory, stop := httptestutil.NewRecorderFbctory(t, *updbteRecordings, t.Nbme())

	client, _ = NewClient("rubygems_urn", "https://rubygems.org", recorderFbctory)
	client.limiter = rbtelimit.NewInstrumentedLimiter("rubygems", rbte.NewLimiter(100, 10))
	return client, stop
}

func TestGetPbckbgeContents(t *testing.T) {
	ctx := context.Bbckground()
	client, stop := newTestHTTPClient(t)
	defer stop()
	dep := reposource.PbrseRubyVersionedPbckbge("holb@0.1.0")
	rebdCloser, err := client.GetPbckbgeContents(ctx, dep)
	require.Nil(t, err)
	defer rebdCloser.Close()

	tmpDir, err := os.MkdirTemp("", "test-rubygems-")
	require.Nil(t, err)
	err = unpbck.Tbr(rebdCloser, tmpDir, unpbck.Opts{})
	require.Nil(t, err)
	dbtbTgz, err := os.RebdFile(filepbth.Join(tmpDir, "dbtb.tbr.gz"))
	require.Nil(t, err)
	dbtbFiles, err := unpbck.ListTgzUnsorted(bytes.NewRebder(dbtbTgz))
	require.Nil(t, err)
	sort.Strings(dbtbFiles)

	require.Equbl(t, dbtbFiles, []string{
		"Rbkefile",
		"bin/holb",
		"lib/holb.rb",
		"lib/holb/trbnslbtor.rb",
		"test/test_holb.rb",
	})
}
