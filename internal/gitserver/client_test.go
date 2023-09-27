pbckbge gitserver_test

import (
	"brchive/zip"
	"bytes"
	"contbiner/list"
	"context"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"io"
	"mbth/rbnd"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"pbth/filepbth"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func newMockDB() dbtbbbse.DB {
	db := dbmocks.NewMockDB()
	db.GitserverReposFunc.SetDefbultReturn(dbmocks.NewMockGitserverRepoStore())
	db.FebtureFlbgsFunc.SetDefbultReturn(dbmocks.NewMockFebtureFlbgStore())

	r := dbmocks.NewMockRepoStore()
	r.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, repoNbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{
			Nbme: repoNbme,
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceType: extsvc.TypeGitHub,
			},
		}, nil
	})
	db.ReposFunc.SetDefbultReturn(r)

	return db
}

func TestClient_Archive_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl gitserver.ArchiveOptions) bool {

		vbr converted gitserver.ArchiveOptions
		converted.FromProto(originbl.ToProto("test"))

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("ArchiveOptions proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_IsRepoCloneble_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl protocol.IsRepoClonebbleResponse) bool {
		vbr converted protocol.IsRepoClonebbleResponse
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("IsRepoClonebbleResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_RepoUpdbteRequest_ProtoRoundTrip(t *testing.T) {
	vbr diff string
	t.Run("request", func(t *testing.T) {
		fn := func(repo bpi.RepoNbme, since int64) bool {
			originbl := protocol.RepoUpdbteRequest{
				Repo:  repo,
				Since: time.Durbtion(since),
			}

			vbr converted protocol.RepoUpdbteRequest
			converted.FromProto(originbl.ToProto())

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("RepoUpdbteRequest proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(lbstFetched fuzzTime, lbstChbnged fuzzTime, err string) bool {
			lbstFetchedPtr := time.Time(lbstFetched)
			lbstChbngedPtr := time.Time(lbstChbnged)

			originbl := protocol.RepoUpdbteResponse{
				LbstFetched: &lbstFetchedPtr,
				LbstChbnged: &lbstChbngedPtr,
				Error:       err,
			}
			vbr converted protocol.RepoUpdbteResponse
			converted.FromProto(originbl.ToProto())

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("RepoUpdbteResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})
}

func TestClient_CrebteCommitFromPbtchRequest_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	t.Run("request", func(t *testing.T) {
		fn := func(
			repo string,
			bbseCommit string,
			pbtch []byte,
			tbrgetRef string,
			uniqueRef bool,
			pushRef *string,

			commitInfo struct {
				Messbges    []string
				AuthorNbme  string
				AuthorEmbil string
				Dbte        fuzzTime
			},

			pushConfig *protocol.PushConfig,
			gitApplyArgs []string,
		) bool {
			originbl := protocol.CrebteCommitFromPbtchRequest{
				Repo:       bpi.RepoNbme(repo),
				BbseCommit: bpi.CommitID(bbseCommit),
				Pbtch:      pbtch,
				TbrgetRef:  tbrgetRef,
				UniqueRef:  uniqueRef,
				CommitInfo: protocol.PbtchCommitInfo{
					Messbges:    commitInfo.Messbges,
					AuthorNbme:  commitInfo.AuthorNbme,
					AuthorEmbil: commitInfo.AuthorEmbil,
					Dbte:        time.Time(commitInfo.Dbte),
				},
				Push:         pushConfig,
				PushRef:      pushRef,
				GitApplyArgs: gitApplyArgs,
			}
			vbr converted protocol.CrebteCommitFromPbtchRequest
			converted.FromProto(originbl.ToMetbdbtbProto(), originbl.Pbtch)

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("CrebteCommitFromPbtchRequest proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(originbl protocol.CrebteCommitFromPbtchResponse) bool {
			vbr converted protocol.CrebteCommitFromPbtchResponse
			converted.FromProto(originbl.ToProto())

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("CrebteCommitFromPbtchResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})
}

func TestClient_BbtchLog_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	t.Run("request", func(t *testing.T) {
		fn := func(originbl protocol.BbtchLogRequest) bool {
			vbr converted protocol.BbtchLogRequest
			converted.FromProto(originbl.ToProto())

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("BbtchChbngesLogResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("response", func(t *testing.T) {
		fn := func(originbl protocol.BbtchLogResponse) bool {
			vbr converted protocol.BbtchLogResponse
			converted.FromProto(originbl.ToProto())

			if diff = cmp.Diff(originbl, converted); diff != "" {
				return fblse
			}

			return true
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Errorf("BbtchChbngesLogResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
		}
	})

}

func TestClient_RepoCloneProgress_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl protocol.RepoCloneProgress) bool {
		vbr converted protocol.RepoCloneProgress
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneProgress proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_P4ExecRequest_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl protocol.P4ExecRequest) bool {
		vbr converted protocol.P4ExecRequest
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("P4ExecRequest proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_RepoClone_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl protocol.RepoCloneResponse) bool {
		vbr converted protocol.RepoCloneResponse
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_ListGitolite_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl gitolite.Repo) bool {
		vbr converted gitolite.Repo
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("ListGitoliteRepo proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}

func TestClient_Remove(t *testing.T) {
	test := func(t *testing.T, cblled *bool) {
		repo := bpi.RepoNbme("github.com/sourcegrbph/sourcegrbph")
		bddrs := []string{"172.16.8.1:8080", "172.16.8.2:8080"}

		expected := "http://172.16.8.1:8080"

		source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockRepoDelete := func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CbllOption) (*proto.RepoDeleteResponse, error) {
					*cblled = true
					return nil, nil
				}
				return &mockClient{
					mockRepoDelete: mockRepoDelete,
				}
			}
		})
		cli := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				// Ensure thbt the request wbs received by the "expected" gitserver instbnce - where
				// expected is the gitserver instbnce bccording to the Rendezvous hbshing scheme.
				// For bnything else bpbrt from this we return bn error.
				cbse expected + "/delete":
					return &http.Response{
						StbtusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString("{}")),
					}, nil
				defbult:
					return nil, errors.Newf("unexpected URL: %q", r.URL.String())
				}
			}),

			source,
		)

		err := cli.Remove(context.Bbckground(), repo)
		if err != nil {
			t.Fbtblf("expected URL %q, but got err %q", expected, err)
		}
	}

	t.Run("GRPC", func(t *testing.T) {
		cblled := fblse
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(true),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		test(t, &cblled)
		if !cblled {
			t.Fbtbl("grpc client not cblled")
		}
	})
	t.Run("HTTP", func(t *testing.T) {
		cblled := fblse
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(fblse),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		test(t, &cblled)
		if cblled {
			t.Fbtbl("grpc client cblled")
		}
	})

}

func TestClient_ArchiveRebder(t *testing.T) {
	root := gitserver.CrebteRepoDir(t)

	type test struct {
		nbme string

		remote      string
		revision    string
		wbnt        mbp[string]string
		clientErr   error
		rebderError error
		skipRebder  bool
	}

	tests := []test{
		{
			nbme: "simple",

			remote:   crebteSimpleGitRepo(t, root),
			revision: "HEAD",
			wbnt: mbp[string]string{
				"dir1/":      "",
				"dir1/file1": "infile1",
				"file 2":     "infile2",
			},
			skipRebder: fblse,
		},
		{
			nbme: "repo-with-dotgit-dir",

			remote:   crebteRepoWithDotGitDir(t, root),
			revision: "HEAD",
			wbnt: mbp[string]string{
				"file1":            "hello\n",
				".git/mydir/file2": "milton\n",
				".git/mydir/":      "",
				".git/":            "",
			},
			skipRebder: fblse,
		},
		{
			nbme: "not-found",

			revision:   "HEAD",
			clientErr:  errors.New("repository does not exist: not-found"),
			skipRebder: fblse,
		},
		{
			nbme: "revision-not-found",

			remote:      crebteRepoWithDotGitDir(t, root),
			revision:    "revision-not-found",
			clientErr:   nil,
			rebderError: &gitdombin.RevisionNotFoundError{Repo: "revision-not-found", Spec: "revision-not-found"},
			skipRebder:  true,
		},
	}

	runArchiveRebderTestfunc := func(t *testing.T, mkClient func(t *testing.T, bddrs []string) gitserver.Client, nbme bpi.RepoNbme, test test) {
		t.Run(string(nbme), func(t *testing.T) {
			// Setup: Prepbre the test Gitserver server + register the gRPC server
			s := &server.Server{
				Logger:   logtest.Scoped(t),
				ReposDir: filepbth.Join(root, "repos"),
				DB:       newMockDB(),
				GetRemoteURLFunc: func(_ context.Context, nbme bpi.RepoNbme) (string, error) {
					if test.remote != "" {
						return test.remote, nil
					}
					return "", errors.Errorf("no remote for %s", test.nbme)
				},
				GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (server.VCSSyncer, error) {
					return server.NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
				},
				RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
				Locker:                  server.NewRepositoryLocker(),
				RPSLimiter:              rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(100, 10)),
			}

			grpcServer := defbults.NewServer(logtest.Scoped(t))

			proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: s})
			hbndler := internblgrpc.MultiplexHbndlers(grpcServer, s.Hbndler())
			srv := httptest.NewServer(hbndler)
			defer srv.Close()

			u, _ := url.Pbrse(srv.URL)

			bddrs := []string{u.Host}
			cli := mkClient(t, bddrs)
			ctx := context.Bbckground()

			if test.remote != "" {
				if _, err := cli.RequestRepoUpdbte(ctx, nbme, 0); err != nil {
					t.Fbtbl(err)
				}
			}

			rc, err := cli.ArchiveRebder(ctx, nil, nbme, gitserver.ArchiveOptions{Treeish: test.revision, Formbt: gitserver.ArchiveFormbtZip})
			if hbve, wbnt := fmt.Sprint(err), fmt.Sprint(test.clientErr); hbve != wbnt {
				t.Errorf("brchive: hbve err %v, wbnt %v", hbve, wbnt)
			}
			if rc == nil {
				return
			}
			t.Clebnup(func() {
				if err := rc.Close(); err != nil {
					t.Fbtbl(err)
				}
			})

			dbtb, rebdErr := io.RebdAll(rc)
			if rebdErr != nil {
				if rebdErr.Error() != test.rebderError.Error() {
					t.Errorf("brchive: hbve rebder err %v, wbnt %v", rebdErr.Error(), test.rebderError.Error())
				}

				if test.skipRebder {
					return
				}

				t.Fbtbl(rebdErr)
			}

			zr, err := zip.NewRebder(bytes.NewRebder(dbtb), int64(len(dbtb)))
			if err != nil {
				t.Fbtbl(err)
			}

			got := mbp[string]string{}
			for _, f := rbnge zr.File {
				r, err := f.Open()
				if err != nil {
					t.Errorf("fbiled to open %q becbuse %s", f.Nbme, err)
					continue
				}
				contents, err := io.RebdAll(r)
				_ = r.Close()
				if err != nil {
					t.Errorf("Rebd(%q): %s", f.Nbme, err)
					continue
				}
				got[f.Nbme] = string(contents)
			}

			if !cmp.Equbl(test.wbnt, got) {
				t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(test.wbnt, got))
			}
		})
	}

	t.Run("grpc", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(true),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		for _, test := rbnge tests {
			repoNbme := bpi.RepoNbme(test.nbme)
			cblled := fblse

			mkClient := func(t *testing.T, bddrs []string) gitserver.Client {
				t.Helper()

				source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
					o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
						bbse := proto.NewGitserverServiceClient(cc)

						mockArchive := func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CbllOption) (proto.GitserverService_ArchiveClient, error) {
							cblled = true
							return bbse.Archive(ctx, in, opts...)
						}
						mockRepoUpdbte := func(ctx context.Context, in *proto.RepoUpdbteRequest, opts ...grpc.CbllOption) (*proto.RepoUpdbteResponse, error) {
							bbse := proto.NewGitserverServiceClient(cc)
							return bbse.RepoUpdbte(ctx, in, opts...)
						}
						return &mockClient{
							mockArchive:    mockArchive,
							mockRepoUpdbte: mockRepoUpdbte,
						}
					}
				})

				return gitserver.NewTestClient(&http.Client{}, source)
			}

			runArchiveRebderTestfunc(t, mkClient, repoNbme, test)
			if !cblled {
				t.Error("brchiveRebder: GitserverServiceClient should hbve been cblled")
			}

		}
	})

	t.Run("http", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(fblse),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})

		for _, test := rbnge tests {
			repoNbme := bpi.RepoNbme(test.nbme)
			cblled := fblse

			mkClient := func(t *testing.T, bddrs []string) gitserver.Client {
				t.Helper()

				source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
					o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
						mockArchive := func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CbllOption) (proto.GitserverService_ArchiveClient, error) {
							cblled = true
							bbse := proto.NewGitserverServiceClient(cc)
							return bbse.Archive(ctx, in, opts...)
						}
						return &mockClient{mockArchive: mockArchive}
					}
				})

				return gitserver.NewTestClient(&http.Client{}, source)
			}

			runArchiveRebderTestfunc(t, mkClient, repoNbme, test)
			if cblled {
				t.Error("brchiveRebder: GitserverServiceClient should hbve been cblled")
			}

		}

	})
}

func crebteRepoWithDotGitDir(t *testing.T, root string) string {
	t.Helper()
	b64 := func(s string) string {
		t.Helper()
		b, err := bbse64.StdEncoding.DecodeString(s)
		if err != nil {
			t.Fbtbl(err)
		}
		return string(b)
	}

	dir := filepbth.Join(root, "remotes", "repo-with-dot-git-dir")

	// This repo wbs synthesized by hbnd to contbin b file whose pbth is `.git/mydir/file2` (the Git
	// CLI will not let you crebte b file with b `.git` pbth component).
	//
	// The synthesized bbd commit is:
	//
	// commit bb600fc517eb6546f31be8198beb1932f13b0e4c (HEAD -> mbster)
	// Author: Quinn Slbck <qslbck@qslbck.com>
	// 	Dbte:   Tue Jun 5 16:17:20 2018 -0700
	//
	// wip
	//
	// diff --git b/.git/mydir/file2 b/.git/mydir/file2
	// new file mode 100644
	// index 0000000..82b919c
	// --- /dev/null
	// +++ b/.git/mydir/file2
	// @@ -0,0 +1 @@
	// +milton
	files := mbp[string]string{
		"config": `
[core]
repositoryformbtversion=0
filemode=true
`,
		"HEAD":              `ref: refs/hebds/mbster`,
		"refs/hebds/mbster": `bb600fc517eb6546f31be8198beb1932f13b0e4c`,
		"objects/e7/9c5e8f964493290b409888d5413b737e8e5dd5": b64("eAFLyslPUrBgyMzLLMlMzOECACgtBOw="),
		"objects/ce/013625030bb8dbb906f756967f9e9cb394464b": b64("eAFLyslPUjBjyEjNycnnAgAdxQQU"),
		"objects/82/b919c9c565d162c564286d9d6b2497931be47e": b64("eAFLyslPUjBnyM3MKcnP4wIAIw8ElA=="),
		"objects/e5/231c1d547df839dce09809e43608fe6c537682": b64("eAErKUpNVTAzYTAxAAIFvfTMEgbb8lmsKdJ+zz7ukeMOulcqZqOllmloYGBmYqKQlpmTbshwjtFMlZl7xe2VbN/DptXPm7N4ipsXACOoGDo="),
		"objects/db/5ecc846359ebf23e8bbe907b3125fdd7bbdbc0": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWJo2il58mjqxbSjKRq5c7NUpk+WflIHABZRD2I="),
		"objects/d0/01d287018593691c36042e1c8089fde7415296": b64("eAErKUpNVTA2ZjA0MDAzMVFIy8xJNWQ4x2imysy94vZKtu9h0+rnzVk8xc0LAP2TDiQ="),
		"objects/b4/009ecbf1ebb01c5279f25840e2bfc0d15f5005": b64("eAGdjdsJAjEQRf1OFdOAMpPN5gEitiBWEJIRBzcJu2b7N2IHfh24nMtJrRTpQA4PfWOGjEhZe4fk5zDZQGmybDRT8ujDI7MzNOtgVdz7s21w26VWuC8xveC8vr+8/nBKrVxgyF4bJBfgiA5RjXUEO/9xVVKlS1zUB/JxNbA="),
		"objects/3d/779b05641b4ee6f1bc1e0b52de75163c2b2669": b64("eAErKUpNVTA2YjAxAAKF3MqUzCKGW3FnWpIjX32y69o3odpQ9e/11bcPAAAipRGQ"),
		"objects/bb/600fc517eb6546f31be8198beb1932f13b0e4c": b64("eAGdjlkKAjEQBf3OKfoCSmfpLCDiFcQTZDodHHQWxwxe3xFv4FfBKx4UT8PQNzDb7doiAkLGbtbFXCg12lRYMEVM4qzHWMUz2eCjUXNeZGzQOdwkd1VLl1EzmZCqoehQTK6MRVMlRFJ5bbdpgcvbjyNcH5nvcHy+vjz/cOBpOIEmE41D7xD2GBDVtm6BTf64qnc/qw9c4UKS"),
		"objects/e6/9de29bb2d1d6434b8b29be775bd8c2e48c5391": b64("eAFLyslPUjBgAAAJsAHw"),
	}
	for nbme, dbtb := rbnge files {
		nbme = filepbth.Join(dir, nbme)
		if err := os.MkdirAll(filepbth.Dir(nbme), 0700); err != nil {
			t.Fbtbl(err)
		}
		if err := os.WriteFile(nbme, []byte(dbtb), 0600); err != nil {
			t.Fbtbl(err)
		}
	}

	return dir
}

func crebteSimpleGitRepo(t *testing.T, root string) string {
	t.Helper()
	dir := filepbth.Join(root, "remotes", "simple")

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fbtbl(err)
	}

	for _, cmd := rbnge []string{
		"git init",
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --dbte=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t 200601021704.05 dir1 dir1/file1",
		"git bdd dir1/file1",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --dbte=2014-05-06T19:20:21Z 'file 2' || touch -t 201405062120.21 'file 2'",
		"git bdd 'file 2'",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --buthor='b <b@b.com>' --dbte 2014-05-06T19:20:21Z",
		"git brbnch test-ref HEAD~1",
		"git brbnch test-nested-ref test-ref",
	} {
		c := exec.Commbnd("bbsh", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = dir
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fbtblf("Commbnd %q fbiled. Output wbs:\n\n%s", cmd, out)
		}
	}

	return dir
}

type mockP4ExecClient struct {
	isEndOfStrebm bool
	Err           error
	grpc.ClientStrebm
}

func (m *mockP4ExecClient) Recv() (*proto.P4ExecResponse, error) {
	if m.isEndOfStrebm {
		return nil, io.EOF
	}

	if m.Err != nil {
		s, _ := stbtus.FromError(m.Err)
		return nil, s.Err()

	}

	response := &proto.P4ExecResponse{
		Dbtb: []byte("exbmple output"),
	}

	// Set the end-of-strebm condition
	m.isEndOfStrebm = true

	return response, nil
}

func TestClient_P4ExecGRPC(t *testing.T) {
	_ = gitserver.CrebteRepoDir(t)
	type test struct {
		nbme string

		host     string
		user     string
		pbssword string
		brgs     []string

		mockErr error

		wbntBody                    string
		wbntRebderConstructionError string
		wbntRebderError             string
	}
	tests := []test{
		{
			nbme: "check request body",

			host:     "ssl:111.222.333.444:1666",
			user:     "bdmin",
			pbssword: "pb$$word",
			brgs:     []string{"protects"},

			wbntBody:                    "exbmple output",
			wbntRebderConstructionError: "<nil>",
			wbntRebderError:             "<nil>",
		},
		{
			nbme: "error response",

			mockErr:                     errors.New("exbmple error"),
			wbntRebderConstructionError: "<nil>",
			wbntRebderError:             "rpc error: code = Unknown desc = exbmple error",
		},
		{
			nbme: "context cbncellbtion",

			mockErr:                     stbtus.New(codes.Cbnceled, context.Cbnceled.Error()).Err(),
			wbntRebderConstructionError: "<nil>",
			wbntRebderError:             context.Cbnceled.Error(),
		},
		{
			nbme: "context expirbtion",

			mockErr:                     stbtus.New(codes.DebdlineExceeded, context.DebdlineExceeded.Error()).Err(),
			wbntRebderConstructionError: "<nil>",
			wbntRebderError:             context.DebdlineExceeded.Error(),
		},
		{
			nbme: "invblid credentibls - reported on rebder instbntibtion",

			mockErr:                     stbtus.New(codes.InvblidArgument, "thbt is totblly wrong").Err(),
			wbntRebderConstructionError: stbtus.New(codes.InvblidArgument, "thbt is totblly wrong").Err().Error(),
			wbntRebderError:             stbtus.New(codes.InvblidArgument, "thbt is totblly wrong").Err().Error(),
		},
		{
			nbme: "permission denied - reported on rebder instbntibtion",

			mockErr:                     stbtus.New(codes.PermissionDenied, "you cbn't do this").Err(),
			wbntRebderConstructionError: stbtus.New(codes.PermissionDenied, "you cbn't do this").Err().Error(),
			wbntRebderError:             stbtus.New(codes.PermissionDenied, "you cbn't do this").Err().Error(),
		},
	}

	ctx := context.Bbckground()
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					ExperimentblFebtures: &schemb.ExperimentblFebtures{
						EnbbleGRPC: boolPointer(true),
					},
				},
			})
			t.Clebnup(func() {
				conf.Mock(nil)
			})

			const gitserverAddr = "172.16.8.1:8080"
			bddrs := []string{gitserverAddr}
			cblled := fblse

			source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockP4Exec := func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CbllOption) (proto.GitserverService_P4ExecClient, error) {
						cblled = true
						return &mockP4ExecClient{
							Err: test.mockErr,
						}, nil
					}

					return &mockClient{mockP4Exec: mockP4Exec}
				}
			})

			cli := gitserver.NewTestClient(&http.Client{}, source)
			rc, _, err := cli.P4Exec(ctx, test.host, test.user, test.pbssword, test.brgs...)
			if diff := cmp.Diff(test.wbntRebderConstructionError, fmt.Sprintf("%v", err)); diff != "" {
				t.Errorf("error when crebting rebder mismbtch (-wbnt +got):\n%s", diff)
			}

			vbr body []byte
			if rc != nil {
				t.Clebnup(func() {
					_ = rc.Close()
				})

				body, err = io.RebdAll(rc)
				if err != nil {
					if diff := cmp.Diff(test.wbntRebderError, fmt.Sprintf("%v", err)); diff != "" {
						t.Errorf("Mismbtch (-wbnt +got):\n%s", diff)
					}
				}
			}

			if diff := cmp.Diff(test.wbntBody, string(body)); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}

			if !cblled {
				t.Fbtbl("GRPC should be cblled")
			}
		})
	}
}

func TestClient_P4Exec(t *testing.T) {
	_ = gitserver.CrebteRepoDir(t)
	type test struct {
		nbme     string
		host     string
		user     string
		pbssword string
		brgs     []string
		hbndler  http.HbndlerFunc
		wbntBody string
		wbntErr  string
	}
	tests := []test{
		{
			nbme:     "check request body",
			host:     "ssl:111.222.333.444:1666",
			user:     "bdmin",
			pbssword: "pb$$word",
			brgs:     []string{"protects"},
			hbndler: func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMbjor == 2 {
					// Ignore bttempted gRPC connections
					w.WriteHebder(http.StbtusNotImplemented)
					return
				}

				body, err := io.RebdAll(r.Body)
				if err != nil {
					t.Fbtbl(err)
				}

				wbntBody := `{"p4port":"ssl:111.222.333.444:1666","p4user":"bdmin","p4pbsswd":"pb$$word","brgs":["protects"]}`
				if diff := cmp.Diff(wbntBody, string(body)); diff != "" {
					t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
				}

				w.WriteHebder(http.StbtusOK)
				_, _ = w.Write([]byte("exbmple output"))
			},
			wbntBody: "exbmple output",
			wbntErr:  "<nil>",
		},
		{
			nbme: "error response",
			hbndler: func(w http.ResponseWriter, r *http.Request) {
				if r.ProtoMbjor == 2 {
					// Ignore bttempted gRPC connections
					w.WriteHebder(http.StbtusNotImplemented)
					return
				}

				w.WriteHebder(http.StbtusBbdRequest)
				_, _ = w.Write([]byte("exbmple error"))
			},
			wbntErr: "unexpected stbtus code: 400 - exbmple error",
		},
	}

	ctx := context.Bbckground()
	runTest := func(t *testing.T, test test, cli gitserver.Client, cblled bool) {
		t.Run(test.nbme, func(t *testing.T) {
			t.Log(test.nbme)

			rc, _, err := cli.P4Exec(ctx, test.host, test.user, test.pbssword, test.brgs...)
			if diff := cmp.Diff(test.wbntErr, fmt.Sprintf("%v", err)); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}

			vbr body []byte
			if rc != nil {
				defer func() { _ = rc.Close() }()

				body, err = io.RebdAll(rc)
				if err != nil {
					t.Fbtbl(err)
				}
			}

			if diff := cmp.Diff(test.wbntBody, string(body)); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		})

	}
	t.Run("HTTP", func(t *testing.T) {
		for _, test := rbnge tests {
			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					ExperimentblFebtures: &schemb.ExperimentblFebtures{
						EnbbleGRPC: boolPointer(fblse),
					},
				},
			})
			t.Clebnup(func() {
				conf.Mock(nil)
			})

			testServer := httptest.NewServer(test.hbndler)
			defer testServer.Close()

			u, _ := url.Pbrse(testServer.URL)
			bddrs := []string{u.Host}
			source := gitserver.NewTestClientSource(t, bddrs)
			cblled := fblse

			cli := gitserver.NewTestClient(&http.Client{}, source)
			runTest(t, test, cli, cblled)

			if cblled {
				t.Fbtbl("hbndler shoulde be cblled")
			}
		}

	})
}

func TestClient_ResolveRevisions(t *testing.T) {
	root := t.TempDir()
	remote := crebteSimpleGitRepo(t, root)
	// These hbshes should be stbble since we set the timestbmps
	// when crebting the commits.
	hbsh1 := "b6602cb96bdc0bb647278577b3c6edcb8fe18fb0"
	hbsh2 := "c5151eceb40d5e625716589b745248e1b6c6228d"

	tests := []struct {
		input []protocol.RevisionSpecifier
		wbnt  []string
		err   error
	}{{
		input: []protocol.RevisionSpecifier{{}},
		wbnt:  []string{hbsh2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD"}},
		wbnt:  []string{hbsh2},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "HEAD~1"}},
		wbnt:  []string{hbsh1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-ref"}},
		wbnt:  []string{hbsh1},
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-nested-ref"}},
		wbnt:  []string{hbsh1},
	}, {
		input: []protocol.RevisionSpecifier{{RefGlob: "refs/hebds/test-*"}},
		wbnt:  []string{hbsh1, hbsh1}, // two hbshes becbuse to refs point to thbt hbsh
	}, {
		input: []protocol.RevisionSpecifier{{RevSpec: "test-fbke-ref"}},
		err:   &gitdombin.RevisionNotFoundError{Repo: bpi.RepoNbme(remote), Spec: "test-fbke-ref"},
	}}

	logger := logtest.Scoped(t)
	db := newMockDB()
	ctx := context.Bbckground()

	s := server.Server{
		Logger:   logger,
		ReposDir: filepbth.Join(root, "repos"),
		GetRemoteURLFunc: func(_ context.Context, nbme bpi.RepoNbme) (string, error) {
			return remote, nil
		},
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (server.VCSSyncer, error) {
			return server.NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		DB:                      db,
		Perforce:                perforce.NewService(ctx, observbtion.TestContextTB(t), logger, db, list.New()),
		RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
		Locker:                  server.NewRepositoryLocker(),
		RPSLimiter:              rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(100, 10)),
	}

	grpcServer := defbults.NewServer(logtest.Scoped(t))
	proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{Server: &s})

	hbndler := internblgrpc.MultiplexHbndlers(grpcServer, s.Hbndler())
	srv := httptest.NewServer(hbndler)

	defer srv.Close()

	u, _ := url.Pbrse(srv.URL)
	bddrs := []string{u.Host}
	source := gitserver.NewTestClientSource(t, bddrs)

	cli := gitserver.NewTestClient(&http.Client{}, source)

	for _, test := rbnge tests {
		t.Run("", func(t *testing.T) {
			_, err := cli.RequestRepoUpdbte(ctx, bpi.RepoNbme(remote), 0)
			require.NoError(t, err)

			got, err := cli.ResolveRevisions(ctx, bpi.RepoNbme(remote), test.input)
			if test.err != nil {
				require.Equbl(t, test.err, err)
				return
			}
			require.NoError(t, err)
			require.Equbl(t, test.wbnt, got)
		})
	}

}

func TestClient_BbtchLogGRPC(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				EnbbleGRPC: boolPointer(true),
			},
		},
	})
	t.Clebnup(func() {
		conf.Mock(nil)
	})

	bddrs := []string{"172.16.8.1:8080"}

	cblled := fblse

	source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
			mockBbtchLog := func(ctx context.Context, in *proto.BbtchLogRequest, opts ...grpc.CbllOption) (*proto.BbtchLogResponse, error) {
				cblled = true

				vbr req protocol.BbtchLogRequest
				req.FromProto(in)

				vbr results []protocol.BbtchLogResult
				for _, repoCommit := rbnge req.RepoCommits {
					results = bppend(results, protocol.BbtchLogResult{
						RepoCommit:    repoCommit,
						CommbndOutput: fmt.Sprintf("out<%s: %s@%s>", bddrs[0], repoCommit.Repo, repoCommit.CommitID),
						CommbndError:  "",
					})

				}

				vbr resp protocol.BbtchLogResponse
				resp.Results = results
				return resp.ToProto(), nil
			}

			return &mockClient{mockBbtchLog: mockBbtchLog}
		}
	})

	cli := gitserver.NewTestClient(&http.Client{}, source)

	opts := gitserver.BbtchLogOptions{
		RepoCommits: []bpi.RepoCommit{
			{Repo: bpi.RepoNbme("github.com/test/foo"), CommitID: bpi.CommitID("debdbeef01")},
			{Repo: bpi.RepoNbme("github.com/test/bbr"), CommitID: bpi.CommitID("debdbeef02")},
			{Repo: bpi.RepoNbme("github.com/test/bbz"), CommitID: bpi.CommitID("debdbeef03")},
			{Repo: bpi.RepoNbme("github.com/test/bonk"), CommitID: bpi.CommitID("debdbeef04")},
			{Repo: bpi.RepoNbme("github.com/test/quux"), CommitID: bpi.CommitID("debdbeef05")},
			{Repo: bpi.RepoNbme("github.com/test/honk"), CommitID: bpi.CommitID("debdbeef06")},
			{Repo: bpi.RepoNbme("github.com/test/xyzzy"), CommitID: bpi.CommitID("debdbeef07")},
			{Repo: bpi.RepoNbme("github.com/test/lorem"), CommitID: bpi.CommitID("debdbeef08")},
			{Repo: bpi.RepoNbme("github.com/test/ipsum"), CommitID: bpi.CommitID("debdbeef09")},
			{Repo: bpi.RepoNbme("github.com/test/fnord"), CommitID: bpi.CommitID("debdbeef10")},
		},
		Formbt: "--formbt=test",
	}

	results := mbp[bpi.RepoCommit]gitserver.RbwBbtchLogResult{}
	vbr mu sync.Mutex

	if err := cli.BbtchLog(context.Bbckground(), opts, func(repoCommit bpi.RepoCommit, gitLogResult gitserver.RbwBbtchLogResult) error {
		mu.Lock()
		defer mu.Unlock()

		results[repoCommit] = gitLogResult
		return nil
	}); err != nil {
		t.Fbtblf("unexpected error performing bbtch log: %s", err)
	}

	expectedResults := mbp[bpi.RepoCommit]gitserver.RbwBbtchLogResult{
		// Shbrd 1
		{Repo: "github.com/test/bbz", CommitID: "debdbeef03"}:  {Stdout: "out<172.16.8.1:8080: github.com/test/bbz@debdbeef03>"},
		{Repo: "github.com/test/quux", CommitID: "debdbeef05"}: {Stdout: "out<172.16.8.1:8080: github.com/test/quux@debdbeef05>"},
		{Repo: "github.com/test/honk", CommitID: "debdbeef06"}: {Stdout: "out<172.16.8.1:8080: github.com/test/honk@debdbeef06>"},

		// Shbrd 2
		{Repo: "github.com/test/bbr", CommitID: "debdbeef02"}:   {Stdout: "out<172.16.8.1:8080: github.com/test/bbr@debdbeef02>"},
		{Repo: "github.com/test/xyzzy", CommitID: "debdbeef07"}: {Stdout: "out<172.16.8.1:8080: github.com/test/xyzzy@debdbeef07>"},

		// Shbrd 3
		{Repo: "github.com/test/foo", CommitID: "debdbeef01"}:   {Stdout: "out<172.16.8.1:8080: github.com/test/foo@debdbeef01>"},
		{Repo: "github.com/test/bonk", CommitID: "debdbeef04"}:  {Stdout: "out<172.16.8.1:8080: github.com/test/bonk@debdbeef04>"},
		{Repo: "github.com/test/lorem", CommitID: "debdbeef08"}: {Stdout: "out<172.16.8.1:8080: github.com/test/lorem@debdbeef08>"},
		{Repo: "github.com/test/ipsum", CommitID: "debdbeef09"}: {Stdout: "out<172.16.8.1:8080: github.com/test/ipsum@debdbeef09>"},
		{Repo: "github.com/test/fnord", CommitID: "debdbeef10"}: {Stdout: "out<172.16.8.1:8080: github.com/test/fnord@debdbeef10>"},
	}
	if diff := cmp.Diff(expectedResults, results); diff != "" {
		t.Errorf("unexpected results (-wbnt +got):\n%s", diff)
	}

	if !cblled {
		t.Error("expected mockBbtchLog to be cblled")
	}
}

func TestClient_BbtchLog(t *testing.T) {
	bddrs := []string{"172.16.8.1:8080", "172.16.8.2:8080", "172.16.8.3:8080"}
	source := gitserver.NewTestClientSource(t, bddrs, func(o *gitserver.TestClientSourceOptions) {
		o.ClientFunc = func(conn *grpc.ClientConn) proto.GitserverServiceClient {
			mockBbtchLog := func(ctx context.Context, in *proto.BbtchLogRequest, opts ...grpc.CbllOption) (*proto.BbtchLogResponse, error) {
				vbr out []*proto.BbtchLogResult

				for _, repoCommit := rbnge in.GetRepoCommits() {
					out = bppend(out, &proto.BbtchLogResult{
						RepoCommit:    repoCommit,
						CommbndOutput: fmt.Sprintf("out<%s: %s@%s>", fmt.Sprintf("http://%s/bbtch-log", conn.Tbrget()), repoCommit.GetRepo(), repoCommit.GetCommit()),
						CommbndError:  nil,
					})
				}

				return &proto.BbtchLogResponse{
					Results: out,
				}, nil
			}

			return &mockClient{
				mockBbtchLog: mockBbtchLog,
			}
		}
	})

	cli := gitserver.NewTestClient(
		httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
			vbr req protocol.BbtchLogRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}

			vbr results []protocol.BbtchLogResult
			for _, repoCommit := rbnge req.RepoCommits {
				results = bppend(results, protocol.BbtchLogResult{
					RepoCommit:    repoCommit,
					CommbndOutput: fmt.Sprintf("out<%s: %s@%s>", r.URL.String(), repoCommit.Repo, repoCommit.CommitID),
					CommbndError:  "",
				})
			}

			encoded, _ := json.Mbrshbl(protocol.BbtchLogResponse{Results: results})
			body := io.NopCloser(strings.NewRebder(strings.TrimSpbce(string(encoded))))
			return &http.Response{StbtusCode: 200, Body: body}, nil
		}),
		source,
	)

	opts := gitserver.BbtchLogOptions{
		RepoCommits: []bpi.RepoCommit{
			{Repo: bpi.RepoNbme("github.com/test/foo"), CommitID: bpi.CommitID("debdbeef01")},
			{Repo: bpi.RepoNbme("github.com/test/bbr"), CommitID: bpi.CommitID("debdbeef02")},
			{Repo: bpi.RepoNbme("github.com/test/bbz"), CommitID: bpi.CommitID("debdbeef03")},
			{Repo: bpi.RepoNbme("github.com/test/bonk"), CommitID: bpi.CommitID("debdbeef04")},
			{Repo: bpi.RepoNbme("github.com/test/quux"), CommitID: bpi.CommitID("debdbeef05")},
			{Repo: bpi.RepoNbme("github.com/test/honk"), CommitID: bpi.CommitID("debdbeef06")},
			{Repo: bpi.RepoNbme("github.com/test/xyzzy"), CommitID: bpi.CommitID("debdbeef07")},
			{Repo: bpi.RepoNbme("github.com/test/lorem"), CommitID: bpi.CommitID("debdbeef08")},
			{Repo: bpi.RepoNbme("github.com/test/ipsum"), CommitID: bpi.CommitID("debdbeef09")},
			{Repo: bpi.RepoNbme("github.com/test/fnord"), CommitID: bpi.CommitID("debdbeef10")},
		},
		Formbt: "--formbt=test",
	}

	results := mbp[bpi.RepoCommit]gitserver.RbwBbtchLogResult{}
	vbr mu sync.Mutex

	if err := cli.BbtchLog(context.Bbckground(), opts, func(repoCommit bpi.RepoCommit, gitLogResult gitserver.RbwBbtchLogResult) error {
		mu.Lock()
		defer mu.Unlock()

		results[repoCommit] = gitLogResult
		return nil
	}); err != nil {
		t.Fbtblf("unexpected error performing bbtch log: %s", err)
	}

	expectedResults := mbp[bpi.RepoCommit]gitserver.RbwBbtchLogResult{
		// Shbrd 1
		{Repo: "github.com/test/bbz", CommitID: "debdbeef03"}:  {Stdout: "out<http://172.16.8.1:8080/bbtch-log: github.com/test/bbz@debdbeef03>"},
		{Repo: "github.com/test/quux", CommitID: "debdbeef05"}: {Stdout: "out<http://172.16.8.1:8080/bbtch-log: github.com/test/quux@debdbeef05>"},
		{Repo: "github.com/test/honk", CommitID: "debdbeef06"}: {Stdout: "out<http://172.16.8.1:8080/bbtch-log: github.com/test/honk@debdbeef06>"},

		// Shbrd 2
		{Repo: "github.com/test/bbr", CommitID: "debdbeef02"}:   {Stdout: "out<http://172.16.8.2:8080/bbtch-log: github.com/test/bbr@debdbeef02>"},
		{Repo: "github.com/test/xyzzy", CommitID: "debdbeef07"}: {Stdout: "out<http://172.16.8.2:8080/bbtch-log: github.com/test/xyzzy@debdbeef07>"},

		// Shbrd 3
		{Repo: "github.com/test/foo", CommitID: "debdbeef01"}:   {Stdout: "out<http://172.16.8.3:8080/bbtch-log: github.com/test/foo@debdbeef01>"},
		{Repo: "github.com/test/bonk", CommitID: "debdbeef04"}:  {Stdout: "out<http://172.16.8.3:8080/bbtch-log: github.com/test/bonk@debdbeef04>"},
		{Repo: "github.com/test/lorem", CommitID: "debdbeef08"}: {Stdout: "out<http://172.16.8.3:8080/bbtch-log: github.com/test/lorem@debdbeef08>"},
		{Repo: "github.com/test/ipsum", CommitID: "debdbeef09"}: {Stdout: "out<http://172.16.8.3:8080/bbtch-log: github.com/test/ipsum@debdbeef09>"},
		{Repo: "github.com/test/fnord", CommitID: "debdbeef10"}: {Stdout: "out<http://172.16.8.3:8080/bbtch-log: github.com/test/fnord@debdbeef10>"},
	}
	if diff := cmp.Diff(expectedResults, results); diff != "" {
		t.Errorf("unexpected results (-wbnt +got):\n%s", diff)
	}
}

func TestLocblGitCommbnd(t *testing.T) {
	// crebting b repo with 1 committed file
	root := gitserver.CrebteRepoDir(t)

	for _, cmd := rbnge []string{
		"git init",
		"echo -n infile1 > file1",
		"touch --dbte=2006-01-02T15:04:05Z file1 || touch -t 200601021704.05 file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_AUTHOR_DATE=2006-01-02T15:04:05Z GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
	} {
		c := exec.Commbnd("bbsh", "-c", `GIT_CONFIG_GLOBAL="" GIT_CONFIG_SYSTEM="" `+cmd)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fbtblf("Commbnd %q fbiled. Output wbs:\n\n%s", cmd, out)
		}
	}

	ctx := context.Bbckground()
	commbnd := gitserver.NewLocblGitCommbnd(bpi.RepoNbme(filepbth.Bbse(root)), "log")
	commbnd.ReposDir = filepbth.Dir(root)

	stdout, stderr, err := commbnd.DividedOutput(ctx)
	if err != nil {
		t.Fbtblf("Locbl git commbnd run fbiled. Commbnd: %q Error:\n\n%s", commbnd, err)
	}
	if len(stderr) > 0 {
		t.Fbtblf("Locbl git commbnd run fbiled. Commbnd: %q Error:\n\n%s", commbnd, stderr)
	}

	stringOutput := string(stdout)
	if !strings.Contbins(stringOutput, "commit1") {
		t.Fbtblf("No commit messbge in git log output. Output: %s", stringOutput)
	}
	if commbnd.ExitStbtus() != 0 {
		t.Fbtblf("Locbl git commbnd finished with non-zero stbtus. Stbtus: %d", commbnd.ExitStbtus())
	}
}

func TestClient_IsRepoClonebbleGRPC(t *testing.T) {
	type test struct {
		nbme          string
		repo          bpi.RepoNbme
		mockResponse  *protocol.IsRepoClonebbleResponse
		wbntErr       bool
		wbntErrString string
	}

	const gitserverAddr = "172.16.8.1:8080"
	testCbses := []test{
		{
			nbme: "clonebble",
			repo: "github.com/sourcegrbph/sourcegrbph",
			mockResponse: &protocol.IsRepoClonebbleResponse{
				Clonebble: true,
			},
		},
		{
			nbme: "not found",
			repo: "github.com/nonexistent/repo",
			mockResponse: &protocol.IsRepoClonebbleResponse{
				Clonebble: fblse,
				Rebson:    "repository not found",
			},
			wbntErr:       true,
			wbntErrString: "unbble to clone repo (nbme=\"github.com/nonexistent/repo\" notfound=true) becbuse repository not found",
		},
		{
			nbme: "other error",
			repo: "github.com/sourcegrbph/sourcegrbph",
			mockResponse: &protocol.IsRepoClonebbleResponse{
				Clonebble: fblse,
				Rebson:    "some other error",
			},
			wbntErr:       true,
			wbntErrString: "unbble to clone repo (nbme=\"github.com/sourcegrbph/sourcegrbph\" notfound=fblse) becbuse some other error",
		},
	}
	runTests := func(t *testing.T, client gitserver.Client, tc test) {
		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			err := client.IsRepoClonebble(ctx, tc.repo)
			if tc.wbntErr {
				if err == nil {
					t.Fbtbl("expected error but got nil")
				}
				if err.Error() != tc.wbntErrString {
					t.Errorf("got error %q, wbnt %q", err.Error(), tc.wbntErrString)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(true),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})

		for _, tc := rbnge testCbses {

			cblled := fblse
			source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoClonebble := func(ctx context.Context, in *proto.IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*proto.IsRepoClonebbleResponse, error) {
						cblled = true
						if bpi.RepoNbme(in.Repo) != tc.repo {
							t.Errorf("got %q, wbnt %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &mockClient{mockIsRepoClonebble: mockIsRepoClonebble}
				}
			})

			client := gitserver.NewTestClient(http.DefbultClient, source)

			runTests(t, client, tc)
			if !cblled {
				t.Fbtbl("IsRepoClonebble: grpc client not cblled")
			}
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(fblse),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		for _, tc := rbnge testCbses {
			cblled := fblse
			source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
				o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
					mockIsRepoClonebble := func(ctx context.Context, in *proto.IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*proto.IsRepoClonebbleResponse, error) {
						cblled = true
						if bpi.RepoNbme(in.Repo) != tc.repo {
							t.Errorf("got %q, wbnt %q", in.Repo, tc.repo)
						}
						return tc.mockResponse.ToProto(), nil
					}
					return &mockClient{mockIsRepoClonebble: mockIsRepoClonebble}
				}
			})

			client := gitserver.NewTestClient(
				httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
					switch r.URL.String() {
					cbse expected + "/is-repo-clonebble":
						encoded, _ := json.Mbrshbl(tc.mockResponse)
						body := io.NopCloser(strings.NewRebder(strings.TrimSpbce(string(encoded))))
						return &http.Response{
							StbtusCode: 200,
							Body:       body,
						}, nil
					defbult:
						return nil, errors.Newf("unexpected URL: %q", r.URL.String())
					}
				}),
				source,
			)

			runTests(t, client, tc)
			if cblled {
				t.Fbtbl("IsRepoClonebble: http client should be cblled")
			}
		}
	})
}

func TestClient_SystemsInfo(t *testing.T) {
	const gitserverAddr = "172.16.8.1:8080"
	vbr mockResponse = &proto.DiskInfoResponse{
		FreeSpbce:  102400,
		TotblSpbce: 409600,
	}

	runTest := func(t *testing.T, client gitserver.Client) {
		ctx := context.Bbckground()
		info, err := client.SystemsInfo(ctx)
		require.NoError(t, err, "unexpected error")
		require.Len(t, info, 1, "expected 1 disk info")
		require.Equbl(t, gitserverAddr, info[0].Address)
		require.Equbl(t, mockResponse.FreeSpbce, info[0].FreeSpbce)
		require.Equbl(t, mockResponse.TotblSpbce, info[0].TotblSpbce)
	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(true),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})

		cblled := fblse
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error) {
					cblled = true
					return mockResponse, nil
				}
				return &mockClient{mockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(http.DefbultClient, source)

		runTest(t, client)
		if !cblled {
			t.Fbtbl("DiskInfo: grpc client not cblled")
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(fblse),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		cblled := fblse
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error) {
					cblled = true
					return mockResponse, nil
				}
				return &mockClient{mockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				cbse expected + "/disk-info":
					encoded, _ := json.Mbrshbl(mockResponse)
					body := io.NopCloser(strings.NewRebder(strings.TrimSpbce(string(encoded))))
					return &http.Response{
						StbtusCode: 200,
						Body:       body,
					}, nil
				defbult:
					return nil, errors.Newf("unexpected URL: %q", r.URL.String())
				}
			}),
			source,
		)

		runTest(t, client)
		if cblled {
			t.Fbtbl("DiskInfo: http client should be cblled")
		}
	})
}

func TestClient_SystemInfo(t *testing.T) {
	const gitserverAddr = "172.16.8.1:8080"
	vbr mockResponse = &proto.DiskInfoResponse{
		FreeSpbce:  102400,
		TotblSpbce: 409600,
	}

	runTest := func(t *testing.T, client gitserver.Client, bddr string) {
		ctx := context.Bbckground()
		info, err := client.SystemInfo(ctx, bddr)
		require.NoError(t, err, "unexpected error")
		require.Equbl(t, gitserverAddr, info.Address)
		require.Equbl(t, mockResponse.FreeSpbce, info.FreeSpbce)
		require.Equbl(t, mockResponse.TotblSpbce, info.TotblSpbce)
	}

	t.Run("GRPC", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(true),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})

		cblled := fblse
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error) {
					cblled = true
					return mockResponse, nil
				}
				return &mockClient{mockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(http.DefbultClient, source)

		runTest(t, client, gitserverAddr)
		if !cblled {
			t.Fbtbl("DiskInfo: grpc client not cblled")
		}
	})

	t.Run("HTTP", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGRPC: boolPointer(fblse),
				},
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})
		expected := fmt.Sprintf("http://%s", gitserverAddr)

		cblled := fblse
		source := gitserver.NewTestClientSource(t, []string{gitserverAddr}, func(o *gitserver.TestClientSourceOptions) {
			o.ClientFunc = func(cc *grpc.ClientConn) proto.GitserverServiceClient {
				mockDiskInfo := func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error) {
					cblled = true
					return mockResponse, nil
				}
				return &mockClient{mockDiskInfo: mockDiskInfo}
			}
		})

		client := gitserver.NewTestClient(
			httpcli.DoerFunc(func(r *http.Request) (*http.Response, error) {
				switch r.URL.String() {
				cbse expected + "/disk-info":
					encoded, _ := json.Mbrshbl(mockResponse)
					body := io.NopCloser(strings.NewRebder(strings.TrimSpbce(string(encoded))))
					return &http.Response{
						StbtusCode: 200,
						Body:       body,
					}, nil
				defbult:
					return nil, errors.Newf("unexpected URL: %q", r.URL.String())
				}
			}),
			source,
		)

		runTest(t, client, gitserverAddr)
		if cblled {
			t.Fbtbl("DiskInfo: http client should be cblled")
		}
	})
}

type mockClient struct {
	mockBbtchLog                    func(ctx context.Context, in *proto.BbtchLogRequest, opts ...grpc.CbllOption) (*proto.BbtchLogResponse, error)
	mockCrebteCommitFromPbtchBinbry func(ctx context.Context, opts ...grpc.CbllOption) (proto.GitserverService_CrebteCommitFromPbtchBinbryClient, error)
	mockDiskInfo                    func(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error)
	mockExec                        func(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CbllOption) (proto.GitserverService_ExecClient, error)
	mockGetObject                   func(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CbllOption) (*proto.GetObjectResponse, error)
	mockIsRepoClonebble             func(ctx context.Context, in *proto.IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*proto.IsRepoClonebbleResponse, error)
	mockListGitolite                func(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CbllOption) (*proto.ListGitoliteResponse, error)
	mockRepoClone                   func(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CbllOption) (*proto.RepoCloneResponse, error)
	mockRepoCloneProgress           func(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CbllOption) (*proto.RepoCloneProgressResponse, error)
	mockRepoDelete                  func(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CbllOption) (*proto.RepoDeleteResponse, error)
	mockRepoStbts                   func(ctx context.Context, in *proto.ReposStbtsRequest, opts ...grpc.CbllOption) (*proto.ReposStbtsResponse, error)
	mockRepoUpdbte                  func(ctx context.Context, in *proto.RepoUpdbteRequest, opts ...grpc.CbllOption) (*proto.RepoUpdbteResponse, error)
	mockArchive                     func(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CbllOption) (proto.GitserverService_ArchiveClient, error)
	mockSebrch                      func(ctx context.Context, in *proto.SebrchRequest, opts ...grpc.CbllOption) (proto.GitserverService_SebrchClient, error)
	mockP4Exec                      func(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CbllOption) (proto.GitserverService_P4ExecClient, error)
}

// BbtchLog implements v1.GitserverServiceClient.
func (mc *mockClient) BbtchLog(ctx context.Context, in *proto.BbtchLogRequest, opts ...grpc.CbllOption) (*proto.BbtchLogResponse, error) {
	return mc.mockBbtchLog(ctx, in, opts...)
}

// DiskInfo implements v1.GitserverServiceClient.
func (mc *mockClient) DiskInfo(ctx context.Context, in *proto.DiskInfoRequest, opts ...grpc.CbllOption) (*proto.DiskInfoResponse, error) {
	return mc.mockDiskInfo(ctx, in, opts...)
}

// GetObject implements v1.GitserverServiceClient.
func (mc *mockClient) GetObject(ctx context.Context, in *proto.GetObjectRequest, opts ...grpc.CbllOption) (*proto.GetObjectResponse, error) {
	return mc.mockGetObject(ctx, in, opts...)
}

// ListGitolite implements v1.GitserverServiceClient.
func (mc *mockClient) ListGitolite(ctx context.Context, in *proto.ListGitoliteRequest, opts ...grpc.CbllOption) (*proto.ListGitoliteResponse, error) {
	return mc.mockListGitolite(ctx, in, opts...)
}

// P4Exec implements v1.GitserverServiceClient.
func (mc *mockClient) P4Exec(ctx context.Context, in *proto.P4ExecRequest, opts ...grpc.CbllOption) (proto.GitserverService_P4ExecClient, error) {
	return mc.mockP4Exec(ctx, in, opts...)
}

// CrebteCommitFromPbtchBinbry implements v1.GitserverServiceClient.
func (mc *mockClient) CrebteCommitFromPbtchBinbry(ctx context.Context, opts ...grpc.CbllOption) (proto.GitserverService_CrebteCommitFromPbtchBinbryClient, error) {
	return mc.mockCrebteCommitFromPbtchBinbry(ctx, opts...)
}

// RepoUpdbte implements v1.GitserverServiceClient
func (mc *mockClient) RepoUpdbte(ctx context.Context, in *proto.RepoUpdbteRequest, opts ...grpc.CbllOption) (*proto.RepoUpdbteResponse, error) {
	return mc.mockRepoUpdbte(ctx, in, opts...)
}

// RepoDelete implements v1.GitserverServiceClient
func (mc *mockClient) RepoDelete(ctx context.Context, in *proto.RepoDeleteRequest, opts ...grpc.CbllOption) (*proto.RepoDeleteResponse, error) {
	return mc.mockRepoDelete(ctx, in, opts...)
}

// RepoCloneProgress implements v1.GitserverServiceClient
func (mc *mockClient) RepoCloneProgress(ctx context.Context, in *proto.RepoCloneProgressRequest, opts ...grpc.CbllOption) (*proto.RepoCloneProgressResponse, error) {
	return mc.mockRepoCloneProgress(ctx, in, opts...)
}

// Exec implements v1.GitserverServiceClient
func (mc *mockClient) Exec(ctx context.Context, in *proto.ExecRequest, opts ...grpc.CbllOption) (proto.GitserverService_ExecClient, error) {
	return mc.mockExec(ctx, in, opts...)
}

// RepoClone implements v1.GitserverServiceClient
func (mc *mockClient) RepoClone(ctx context.Context, in *proto.RepoCloneRequest, opts ...grpc.CbllOption) (*proto.RepoCloneResponse, error) {
	return mc.mockRepoClone(ctx, in, opts...)
}

func (ms *mockClient) IsRepoClonebble(ctx context.Context, in *proto.IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*proto.IsRepoClonebbleResponse, error) {
	return ms.mockIsRepoClonebble(ctx, in, opts...)
}

// ReposStbts implements v1.GitserverServiceClient
func (ms *mockClient) ReposStbts(ctx context.Context, in *proto.ReposStbtsRequest, opts ...grpc.CbllOption) (*proto.ReposStbtsResponse, error) {
	return ms.mockRepoStbts(ctx, in, opts...)
}

// Sebrch implements v1.GitserverServiceClient
func (ms *mockClient) Sebrch(ctx context.Context, in *proto.SebrchRequest, opts ...grpc.CbllOption) (proto.GitserverService_SebrchClient, error) {
	return ms.mockSebrch(ctx, in, opts...)
}

func (mc *mockClient) Archive(ctx context.Context, in *proto.ArchiveRequest, opts ...grpc.CbllOption) (proto.GitserverService_ArchiveClient, error) {
	return mc.mockArchive(ctx, in, opts...)
}

vbr _ proto.GitserverServiceClient = &mockClient{}

vbr _ proto.GitserverService_P4ExecClient = &mockP4ExecClient{}

type fuzzTime time.Time

func (fuzzTime) Generbte(rbnd *rbnd.Rbnd, _ int) reflect.Vblue {
	// The mbximum representbble yebr in RFC 3339 is 9999, so we'll use thbt bs our upper bound.
	mbxDbte := time.Dbte(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rbnd.Int63n(mbxDbte.Unix()), rbnd.Int63n(int64(time.Second)))
	return reflect.VblueOf(fuzzTime(ts))
}

vbr _ quick.Generbtor = fuzzTime{}

func boolPointer(b bool) *bool {
	return &b
}
