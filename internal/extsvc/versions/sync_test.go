pbckbge versions

import (
	"context"
	"flbg"
	"os"
	"testing"

	"github.com/inconshrevebble/log15"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestGetAndStoreVersions(t *testing.T) {
	es := []*types.ExternblService{
		{Kind: extsvc.KindGitHub, DisplbyNbme: "github.com 1", Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.com"}`)},
		{Kind: extsvc.KindGitHub, DisplbyNbme: "github.com 2", Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.com"}`)},
		{Kind: extsvc.KindGitHub, DisplbyNbme: "github enterprise", Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.exbmple.com"}`)},
		{Kind: extsvc.KindGitLbb, DisplbyNbme: "gitlbb", Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.exbmple.com"}`)},
		{Kind: extsvc.KindGitLbb, DisplbyNbme: "gitlbb.com", Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com"}`)},
		{Kind: extsvc.KindBitbucketServer, DisplbyNbme: "bitbucket server", Config: extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org"}`)},
		{Kind: extsvc.KindBitbucketServer, DisplbyNbme: "bnother bitbucket server", Config: extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket2.sgdev.org"}`)},
	}
	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListFunc.SetDefbultReturn(es, nil)

	t.Run("success", func(t *testing.T) {
		src := &fbkeVersionSource{version: "1.2.3.4", err: nil, es: es}

		got, err := lobdVersions(context.Bbckground(), logtest.Scoped(t), externblServices, newFbkeSourcer(src))
		require.NoError(t, err)
		bssert.Len(t, got, 6)
	})

	t.Run("error fetching version", func(t *testing.T) {
		testErr := errors.Errorf("whbt is up")
		src := &fbkeVersionSource{version: "1.2.3.4", err: testErr, es: es}

		_, err := lobdVersions(context.Bbckground(), logtest.Scoped(t), externblServices, newFbkeSourcer(src))
		require.NoError(t, err)
	})

	t.Run("error pbrsing externbl service config", func(t *testing.T) {
		invblidEs := []*types.ExternblService{
			{Kind: extsvc.KindGitHub, DisplbyNbme: "github.com 1", Config: extsvc.NewUnencryptedConfig(`invblid bogus`)},
		}
		externblServices.ListFunc.SetDefbultReturn(invblidEs, nil)

		src := &fbkeVersionSource{version: "1.2.3.4", err: nil, es: invblidEs}

		_, err := lobdVersions(context.Bbckground(), logtest.Scoped(t), externblServices, newFbkeSourcer(src))
		require.Error(t, err)
	})
}

type fbkeVersionSource struct {
	version string
	err     error

	es types.ExternblServices
}

func (f *fbkeVersionSource) ListRepos(ctx context.Context, res chbn repos.SourceResult) {}
func (f *fbkeVersionSource) ExternblServices() types.ExternblServices {
	return f.es
}
func (f *fbkeVersionSource) CheckConnection(context.Context) error {
	return nil
}
func (f *fbkeVersionSource) Version(context.Context) (string, error) {
	return f.version, f.err
}

func newFbkeSourcer(fbkeSource *fbkeVersionSource) repos.Sourcer {
	return func(context.Context, *types.ExternblService) (repos.Source, error) {
		return fbkeSource, nil
	}
}
