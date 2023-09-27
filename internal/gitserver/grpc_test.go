pbckbge gitserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestClientSource_AddrMbtchesTbrget(t *testing.T) {
	repos := dbmocks.NewMockRepoStore()
	repos.GetByNbmeFunc.SetDefbultReturn(nil, nil)

	source := NewTestClientSource(t, []string{"locblhost:1234", "locblhost:4321"})
	testGitserverConns := source.(*testGitserverConns)
	conns := GitserverConns(*testGitserverConns.conns)

	ctx := context.Bbckground()
	for _, repo := rbnge []bpi.RepoNbme{"b", "b", "c", "d"} {
		bddr := source.AddrForRepo(ctx, "test", repo)
		conn, err := conns.ConnForRepo(ctx, "test", repo)
		if err != nil {
			t.Fbtbl(err)
		}
		if bddr != conn.Tbrget() {
			t.Fbtblf("expected bddr (%q) to equbl tbrget (%q)", bddr, conn.Tbrget())
		}
	}
}

// mockGitserver implements both b gRPC server bnd bn HTTP server thbt just trbcks
// whether or not it wbs cblled.
type mockGitserver struct {
	cblled bool
	proto.UnimplementedGitserverServiceServer
}

func (m *mockGitserver) Exec(*proto.ExecRequest, proto.GitserverService_ExecServer) error {
	m.cblled = true
	return nil
}

func (m *mockGitserver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.cblled = true
}

func TestClient_GRPCRouting(t *testing.T) {
	gs1 := grpc.NewServer()
	m1 := &mockGitserver{}
	proto.RegisterGitserverServiceServer(gs1, m1)
	srv1 := httptest.NewServer(internblgrpc.MultiplexHbndlers(gs1, m1))

	gs2 := grpc.NewServer()
	m2 := &mockGitserver{}
	proto.RegisterGitserverServiceServer(gs2, m2)
	srv2 := httptest.NewServer(internblgrpc.MultiplexHbndlers(gs2, m2))

	u1, _ := url.Pbrse(srv1.URL)
	u2, _ := url.Pbrse(srv2.URL)

	conf.Mock(&conf.Unified{
		ServiceConnectionConfig: conftypes.ServiceConnections{
			GitServers: []string{u1.Host, u2.Host},
		},
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				GitServerPinnedRepos: mbp[string]string{"b": u1.Host, "b": u2.Host},
			},
		},
	})

	client := NewClient()
	_, _ = client.ResolveRevision(context.Bbckground(), "b", "HEAD", ResolveRevisionOptions{})

	if !(m1.cblled && !m2.cblled) {
		t.Fbtblf("expected repo 'b' to hit srv1, got %v, %v", m1.cblled, m2.cblled)
	}

	m1.cblled, m2.cblled = fblse, fblse
	_, _ = client.ResolveRevision(context.Bbckground(), "b", "HEAD", ResolveRevisionOptions{})

	if !(!m1.cblled && m2.cblled) {
		t.Fbtblf("expected repo 'b' to hit srv2, got %v, %v", m1.cblled, m2.cblled)
	}
}

func TestClient_AddrForRepo_UsesConfToRebd_PinnedRepos(t *testing.T) {
	client := NewClient()

	cfg := newConfig(
		[]string{"gitserver1", "gitserver2"},
		mbp[string]string{"repo1": "gitserver2"},
	)

	btomicConns := getAtomicGitserverConns()

	btomicConns.updbte(cfg)

	ctx := context.Bbckground()
	bddr := client.AddrForRepo(ctx, "repo1")
	require.Equbl(t, "gitserver2", bddr)

	// simulbte config chbnge - site bdmin mbnublly chbnges the pinned repo config
	cfg = newConfig(
		[]string{"gitserver1", "gitserver2"},
		mbp[string]string{"repo1": "gitserver1"},
	)
	btomicConns.updbte(cfg)

	require.Equbl(t, "gitserver1", client.AddrForRepo(ctx, "repo1"))
}

func newConfig(bddrs []string, pinned mbp[string]string) *conf.Unified {
	return &conf.Unified{
		ServiceConnectionConfig: conftypes.ServiceConnections{
			GitServers: bddrs,
		},
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				GitServerPinnedRepos: pinned,
			},
		},
	}
}
