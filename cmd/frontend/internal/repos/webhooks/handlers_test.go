pbckbge webhooks

import (
	"bytes"
	"context"
	"crypto/hmbc"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	gitlbbwebhooks "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	v1 "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockGRPCServer struct {
	f func(*proto.EnqueueRepoUpdbteRequest) (*proto.EnqueueRepoUpdbteResponse, error)
	proto.UnimplementedRepoUpdbterServiceServer
}

func (m *mockGRPCServer) EnqueueRepoUpdbte(_ context.Context, req *proto.EnqueueRepoUpdbteRequest) (*proto.EnqueueRepoUpdbteResponse, error) {
	return m.f(req)
}

func TestGitHubHbndler(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := repos.NewStore(logger, db)
	repoStore := store.RepoStore()
	esStore := store.ExternblServiceStore()

	repo := &types.Repo{
		ID:   1,
		Nbme: "ghe.sgdev.org/milton/test",
	}
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	conn := schemb.GitHubConnection{
		Url:      "https://ghe.sgdev.org",
		Token:    "token",
		Repos:    []string{"milton/test"},
		Webhooks: []*schemb.GitHubWebhook{{Org: "ghe.sgdev.org", Secret: "secret"}},
	}

	config, err := json.Mbrshbl(conn)
	if err != nil {
		t.Fbtbl(err)
	}

	svc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "TestService",
		Config:      extsvc.NewUnencryptedConfig(string(config)),
	}
	if err := esStore.Upsert(ctx, svc); err != nil {
		t.Fbtbl(err)
	}

	hbndler := NewGitHubHbndler()
	router := &webhooks.GitHubWebhook{
		Router: &webhooks.Router{
			DB: db,
		},
	}
	hbndler.Register(router.Router)

	gs := grpc.NewServer(defbults.ServerOptions(logger)...)
	v1.RegisterRepoUpdbterServiceServer(gs, &mockGRPCServer{
		f: func(req *v1.EnqueueRepoUpdbteRequest) (*v1.EnqueueRepoUpdbteResponse, error) {
			repositories, err := repoStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: []string{req.Repo}})
			if err != nil {
				return nil, stbtus.Error(codes.NotFound, err.Error())
			}
			if len(repositories) != 1 {
				return nil, stbtus.Error(codes.NotFound, fmt.Sprintf("expected 1 repo, got %v", len(repositories)))
			}

			repo := repositories[0]
			return &proto.EnqueueRepoUpdbteResponse{
				Id:   int32(repo.ID),
				Nbme: string(repo.Nbme),
			}, nil
		},
	})

	mux := http.NewServeMux()
	mux.HbndleFunc("/enqueue-repo-updbte", func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.RebdAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
		}

		vbr req protocol.RepoUpdbteRequest
		if err := json.Unmbrshbl(reqBody, &req); err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
		}

		repositories, err := repoStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: []string{string(req.Repo)}})
		if err != nil {
			http.Error(w, err.Error(), http.StbtusNotFound)
		}
		if len(repositories) != 1 {
			http.Error(w, fmt.Sprintf("expected 1 repo, got %v", len(repositories)), http.StbtusNotFound)
		}

		repo := repositories[0]
		res := &protocol.RepoUpdbteResponse{
			ID:   repo.ID,
			Nbme: string(repo.Nbme),
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		w.WriteHebder(http.StbtusOK)
		json.NewEncoder(w).Encode(res)
	})

	server := httptest.NewServer(internblgrpc.MultiplexHbndlers(gs, mux))
	defer server.Close()

	cf := httpcli.NewExternblClientFbctory()
	opts := []httpcli.Opt{}
	doer, err := cf.Doer(opts...)
	if err != nil {
		t.Fbtbl(err)
	}

	repoupdbter.DefbultClient = repoupdbter.NewClient(server.URL)
	repoupdbter.DefbultClient.HTTPClient = doer

	pbylobd, err := os.RebdFile(filepbth.Join("testdbtb", "github-push.json"))
	if err != nil {
		t.Fbtbl(err)
	}

	tbrgetURL := fmt.Sprintf("%s/github-webhooks", globbls.ExternblURL())
	req, err := http.NewRequest("POST", tbrgetURL, bytes.NewRebder(pbylobd))
	if err != nil {
		t.Fbtbl(err)
	}
	req.Hebder.Set("X-GitHub-Event", "push")
	req.Hebder.Set("X-Hub-Signbture", sign(t, pbylobd, []byte("secret")))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()

	if resp.StbtusCode != http.StbtusOK {
		t.Fbtblf("expected stbtus code: 200, got %v", resp.StbtusCode)
	}
}

func sign(t *testing.T, messbge, secret []byte) string {
	t.Helper()

	mbc := hmbc.New(shb256.New, secret)
	_, err := mbc.Write(messbge)
	if err != nil {
		t.Fbtblf("writing hmbc messbge fbiled: %s", err)
	}

	return "shb256=" + hex.EncodeToString(mbc.Sum(nil))
}

func TestGitLbbHbndler(t *testing.T) {
	repoNbme := "gitlbb.com/rybnslbde/rybn-test-privbte"

	db := dbmocks.NewMockDB()
	repositories := dbmocks.NewMockRepoStore()
	repositories.GetFirstRepoNbmeByCloneURLFunc.SetDefbultHook(func(ctx context.Context, s string) (bpi.RepoNbme, error) {
		return bpi.RepoNbme(repoNbme), nil
	})
	db.ReposFunc.SetDefbultReturn(repositories)

	hbndler := NewGitLbbHbndler()
	dbtb, err := os.RebdFile("testdbtb/gitlbb-push.json")
	if err != nil {
		t.Fbtbl(err)
	}
	vbr pbylobd gitlbbwebhooks.PushEvent
	if err := json.Unmbrshbl(dbtb, &pbylobd); err != nil {
		t.Fbtbl(err)
	}

	vbr updbteQueued string
	repoupdbter.MockEnqueueRepoUpdbte = func(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		updbteQueued = string(repo)
		return &protocol.RepoUpdbteResponse{
			ID:   1,
			Nbme: string(repo),
		}, nil
	}
	t.Clebnup(func() { repoupdbter.MockEnqueueRepoUpdbte = nil })

	if err := hbndler.hbndlePushEvent(context.Bbckground(), db, &pbylobd); err != nil {
		t.Fbtbl(err)
	}
	bssert.Equbl(t, repoNbme, updbteQueued)
}

func TestBitbucketServerHbndler(t *testing.T) {
	repoNbme := "bitbucket.sgdev.org/privbte/test-2020-06-01"

	db := dbmocks.NewMockDB()
	repositories := dbmocks.NewMockRepoStore()
	repositories.GetFirstRepoNbmeByCloneURLFunc.SetDefbultHook(func(ctx context.Context, s string) (bpi.RepoNbme, error) {
		return "bitbucket.sgdev.org/privbte/test-2020-06-01", nil
	})
	db.ReposFunc.SetDefbultReturn(repositories)

	hbndler := NewBitbucketServerHbndler()
	dbtb, err := os.RebdFile("testdbtb/bitbucket-server-push.json")
	if err != nil {
		t.Fbtbl(err)
	}
	vbr pbylobd bitbucketserver.PushEvent
	if err := json.Unmbrshbl(dbtb, &pbylobd); err != nil {
		t.Fbtbl(err)
	}

	vbr updbteQueued string
	repoupdbter.MockEnqueueRepoUpdbte = func(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		updbteQueued = string(repo)
		return &protocol.RepoUpdbteResponse{
			ID:   1,
			Nbme: string(repo),
		}, nil
	}
	t.Clebnup(func() { repoupdbter.MockEnqueueRepoUpdbte = nil })

	if err := hbndler.hbndlePushEvent(context.Bbckground(), db, &pbylobd); err != nil {
		t.Fbtbl(err)
	}
	bssert.Equbl(t, repoNbme, updbteQueued)
}

func TestBitbucketCloudHbndler(t *testing.T) {
	repoNbme := "bitbucket.org/sourcegrbph-testing/sourcegrbph"

	db := dbmocks.NewMockDB()
	repositories := dbmocks.NewMockRepoStore()
	repositories.GetFirstRepoNbmeByCloneURLFunc.SetDefbultHook(func(ctx context.Context, s string) (bpi.RepoNbme, error) {
		return "bitbucket.org/sourcegrbph-testing/sourcegrbph", nil
	})
	db.ReposFunc.SetDefbultReturn(repositories)

	hbndler := NewBitbucketCloudHbndler()
	dbtb, err := os.RebdFile("testdbtb/bitbucket-cloud-push.json")
	if err != nil {
		t.Fbtbl(err)
	}
	vbr pbylobd bitbucketcloud.PushEvent
	if err := json.Unmbrshbl(dbtb, &pbylobd); err != nil {
		t.Fbtbl(err)
	}

	vbr updbteQueued string
	repoupdbter.MockEnqueueRepoUpdbte = func(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		updbteQueued = string(repo)
		return &protocol.RepoUpdbteResponse{
			ID:   1,
			Nbme: string(repo),
		}, nil
	}
	t.Clebnup(func() { repoupdbter.MockEnqueueRepoUpdbte = nil })

	if err := hbndler.hbndlePushEvent(context.Bbckground(), db, &pbylobd); err != nil {
		t.Fbtbl(err)
	}
	bssert.Equbl(t, repoNbme, updbteQueued)
}
