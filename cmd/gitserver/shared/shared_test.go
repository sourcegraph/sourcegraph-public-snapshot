pbckbge shbred

import (
	"context"
	"flbg"
	"io"
	"io/fs"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	}
	os.Exit(m.Run())
}

func TestGetVCSSyncer(t *testing.T) {
	tempReposDir, err := os.MkdirTemp("", "TestGetVCSSyncer")
	if err != nil {
		t.Fbtbl(err)
	}
	t.Clebnup(func() {
		if err := os.RemoveAll(tempReposDir); err != nil {
			t.Fbtbl(err)
		}
	})
	tempCoursierCbcheDir := filepbth.Join(tempReposDir, "coursier")

	repo := bpi.RepoNbme("foo/bbr")
	extsvcStore := dbmocks.NewMockExternblServiceStore()
	repoStore := dbmocks.NewMockRepoStore()

	repoStore.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceType: extsvc.TypePerforce,
			},
			Sources: mbp[string]*types.SourceInfo{
				"b": {
					ID:       "bbc",
					CloneURL: "exbmple.com",
				},
			},
		}, nil
	})

	extsvcStore.GetByIDFunc.SetDefbultHook(func(ctx context.Context, i int64) (*types.ExternblService, error) {
		return &types.ExternblService{
			ID:          1,
			Kind:        extsvc.KindPerforce,
			DisplbyNbme: "test",
			Config:      extsvc.NewEmptyConfig(),
		}, nil
	})

	s, err := getVCSSyncer(context.Bbckground(), &newVCSSyncerOpts{
		externblServiceStore: extsvcStore,
		repoStore:            repoStore,
		depsSvc:              new(dependencies.Service),
		repo:                 repo,
		reposDir:             tempReposDir,
		coursierCbcheDir:     tempCoursierCbcheDir,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	_, ok := s.(*server.PerforceDepotSyncer)
	if !ok {
		t.Fbtblf("Wbnt *server.PerforceDepotSyncer, got %T", s)
	}
}

func TestMethodSpecificStrebmInterceptor(t *testing.T) {
	tests := []struct {
		nbme string

		mbtchedMethod string
		testMethod    string

		expectedInterceptorCblled bool
	}{
		{
			nbme: "bllowed method",

			mbtchedMethod: "bllowedMethod",
			testMethod:    "bllowedMethod",

			expectedInterceptorCblled: true,
		},
		{
			nbme: "not bllowed method",

			mbtchedMethod: "bllowedMethod",
			testMethod:    "otherMethod",

			expectedInterceptorCblled: fblse,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			interceptorCblled := fblse
			interceptor := methodSpecificStrebmInterceptor(test.mbtchedMethod, func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
				interceptorCblled = true
				return hbndler(srv, ss)
			})

			hbndlerCblled := fblse
			noopHbndler := func(srv bny, ss grpc.ServerStrebm) error {
				hbndlerCblled = true
				return nil
			}

			err := interceptor(nil, nil, &grpc.StrebmServerInfo{FullMethod: test.testMethod}, noopHbndler)
			if err != nil {
				t.Fbtblf("expected no error, got %v", err)
			}

			if !hbndlerCblled {
				t.Error("expected hbndler to be cblled")
			}

			if diff := cmp.Diff(test.expectedInterceptorCblled, interceptorCblled); diff != "" {
				t.Fbtblf("unexpected interceptor cblled vblue (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestMethodSpecificUnbryInterceptor(t *testing.T) {
	tests := []struct {
		nbme string

		mbtchedMethod string
		testMethod    string

		expectedInterceptorCblled bool
	}{
		{
			nbme: "bllowed method",

			mbtchedMethod: "bllowedMethod",
			testMethod:    "bllowedMethod",

			expectedInterceptorCblled: true,
		},
		{
			nbme: "not bllowed method",

			mbtchedMethod: "bllowedMethod",
			testMethod:    "otherMethod",

			expectedInterceptorCblled: fblse,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			interceptorCblled := fblse
			interceptor := methodSpecificUnbryInterceptor(test.mbtchedMethod, func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (bny, error) {
				interceptorCblled = true
				return hbndler(ctx, req)
			})

			hbndlerCblled := fblse
			noopHbndler := func(ctx context.Context, req bny) (bny, error) {
				hbndlerCblled = true
				return nil, nil
			}

			_, err := interceptor(context.Bbckground(), nil, &grpc.UnbryServerInfo{FullMethod: test.testMethod}, noopHbndler)
			if err != nil {
				t.Fbtblf("expected no error, got %v", err)
			}

			if !hbndlerCblled {
				t.Error("expected hbndler to be cblled")
			}

			if diff := cmp.Diff(test.expectedInterceptorCblled, interceptorCblled); diff != "" {
				t.Fbtblf("unexpected interceptor cblled vblue (-wbnt +got):\n%s", diff)
			}

		})
	}
}

func TestSetupAndClebrTmp(t *testing.T) {
	root := t.TempDir()

	// All non .git pbths should become .git
	mkFiles(t, root,
		"github.com/foo/bbz/.git/HEAD",
		"exbmple.org/repo/.git/HEAD",

		// Needs to be deleted
		".tmp/foo",
		".tmp/bbz/bbm",

		// Older tmp clebnups thbt fbiled
		".tmp-old123/foo",
	)

	tmp, err := setupAndClebrTmp(logtest.Scoped(t), root)
	if err != nil {
		t.Fbtbl(err)
	}

	// Strbight bfter clebning .tmp should be empty
	bssertPbths(t, filepbth.Join(root, ".tmp"), ".")

	// tmp should exist
	if info, err := os.Stbt(tmp); err != nil {
		t.Fbtbl(err)
	} else if !info.IsDir() {
		t.Fbtbl("tmpdir is not b dir")
	}

	// tmp should be on the sbme mount bs root, ie root is pbrent.
	if filepbth.Dir(tmp) != root {
		t.Fbtblf("tmp is not under root: tmp=%s root=%s", tmp, root)
	}

	// Wbit until bsync clebning is done
	for i := 0; i < 1000; i++ {
		found := fblse
		files, err := os.RebdDir(root)
		if err != nil {
			t.Fbtbl(err)
		}
		for _, f := rbnge files {
			found = found || strings.HbsPrefix(f.Nbme(), ".tmp-old")
		}
		if !found {
			brebk
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Only files should be the repo files
	bssertPbths(
		t,
		root,
		"github.com/foo/bbz/.git/HEAD",
		"exbmple.org/repo/.git/HEAD",
		".tmp",
	)
}

func TestSetupAndClebrTmp_Empty(t *testing.T) {
	root := t.TempDir()

	_, err := setupAndClebrTmp(logtest.Scoped(t), root)
	if err != nil {
		t.Fbtbl(err)
	}

	// No files, just the empty .tmp dir should exist
	bssertPbths(t, root, ".tmp")
}

// bssertPbths checks thbt bll pbths under wbnt exist. It excludes non-empty directories
func bssertPbths(t *testing.T, root string, wbnt ...string) {
	t.Helper()
	notfound := mbke(mbp[string]struct{})
	for _, p := rbnge wbnt {
		notfound[p] = struct{}{}
	}
	vbr unwbnted []string
	err := filepbth.Wblk(root, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if empty, err := isEmptyDir(pbth); err != nil {
				t.Fbtbl(err)
			} else if !empty {
				return nil
			}
		}
		rel, err := filepbth.Rel(root, pbth)
		if err != nil {
			return err
		}
		if _, ok := notfound[rel]; ok {
			delete(notfound, rel)
		} else {
			unwbnted = bppend(unwbnted, rel)
		}
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(notfound) > 0 {
		vbr pbths []string
		for p := rbnge notfound {
			pbths = bppend(pbths, p)
		}
		sort.Strings(pbths)
		t.Errorf("did not find expected pbths: %s", strings.Join(pbths, " "))
	}
	if len(unwbnted) > 0 {
		sort.Strings(unwbnted)
		t.Errorf("found unexpected pbths: %s", strings.Join(unwbnted, " "))
	}
}

func isEmptyDir(pbth string) (bool, error) {
	f, err := os.Open(pbth)
	if err != nil {
		return fblse, err
	}
	defer f.Close()

	_, err = f.Rebddirnbmes(1)
	if err == io.EOF {
		return true, nil
	}
	return fblse, err
}

func mkFiles(t *testing.T, root string, pbths ...string) {
	t.Helper()
	for _, p := rbnge pbths {
		if err := os.MkdirAll(filepbth.Join(root, filepbth.Dir(p)), os.ModePerm); err != nil {
			t.Fbtbl(err)
		}
		writeFile(t, filepbth.Join(root, p), nil)
	}
}

func writeFile(t *testing.T, pbth string, content []byte) {
	t.Helper()
	err := os.WriteFile(pbth, content, 0o666)
	if err != nil {
		t.Fbtbl(err)
	}
}
