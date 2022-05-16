package versions

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestGetAndStoreVersions(t *testing.T) {
	es := []*types.ExternalService{
		{Kind: extsvc.KindGitHub, DisplayName: "github.com 1", Config: `{"url": "https://github.com"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "github.com 2", Config: `{"url": "https://github.com"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "github enterprise", Config: `{"url": "https://github.example.com"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "gitlab", Config: `{"url": "https://gitlab.example.com"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "gitlab.com", Config: `{"url": "https://gitlab.com"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "bitbucket server", Config: `{"url": "https://bitbucket.sgdev.org"}`},
		{Kind: extsvc.KindGitHub, DisplayName: "another bitbucket server", Config: `{"url": "https://bitbucket2.sgdev.org"}`},
	}
	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(es, nil)

	t.Run("success", func(t *testing.T) {
		src := &fakeVersionSource{version: "1.2.3.4", err: nil, es: es}

		got, err := loadVersions(context.Background(), logtest.Scoped(t), externalServices, newFakeSourcer(src))
		require.NoError(t, err)
		assert.Len(t, got, 6)
	})

	t.Run("error fetching version", func(t *testing.T) {
		testErr := errors.Errorf("what is up")
		src := &fakeVersionSource{version: "1.2.3.4", err: testErr, es: es}

		_, err := loadVersions(context.Background(), logtest.Scoped(t), externalServices, newFakeSourcer(src))
		require.NoError(t, err)
	})

	t.Run("error parsing external service config", func(t *testing.T) {
		invalidEs := []*types.ExternalService{
			{Kind: extsvc.KindGitHub, DisplayName: "github.com 1", Config: `invalid bogus`},
		}
		externalServices.ListFunc.SetDefaultReturn(invalidEs, nil)

		src := &fakeVersionSource{version: "1.2.3.4", err: nil, es: invalidEs}

		_, err := loadVersions(context.Background(), logtest.Scoped(t), externalServices, newFakeSourcer(src))
		require.Error(t, err)
	})
}

type fakeVersionSource struct {
	version string
	err     error

	es types.ExternalServices
}

func (f *fakeVersionSource) ListRepos(ctx context.Context, res chan repos.SourceResult) {}
func (f *fakeVersionSource) ExternalServices() types.ExternalServices {
	return f.es
}
func (f *fakeVersionSource) Version(context.Context) (string, error) {
	return f.version, f.err
}

func newFakeSourcer(fakeSource *fakeVersionSource) repos.Sourcer {
	return func(e *types.ExternalService) (repos.Source, error) {
		return fakeSource, nil
	}
}
