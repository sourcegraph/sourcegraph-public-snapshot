pbckbge bbckground

import (
	"brchive/tbr"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"pbth"
	"strings"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCrbteSyncer(t *testing.T) {
	clock := glock.NewMockClock()
	rbwConfig, _ := json.Mbrshbl(schemb.RustPbckbgesConnection{
		IndexRepositoryNbme:         "github.com/rust-lbng/crbtes.io-index",
		IndexRepositorySyncIntervbl: "1m",
	})

	// dont need bny functionblity for this
	butoindexSvc := NewMockAutoIndexingService()

	refs := mbke(mbp[reposource.PbckbgeNbme][]shbred.MinimblPbckbgeRepoRefVersion)
	dependenciesSvc := NewMockDependenciesService()
	dependenciesSvc.InsertPbckbgeRepoRefsFunc.SetDefbultHook(func(ctx context.Context, refList []shbred.MinimblPbckbgeRepoRef) (newRef []shbred.PbckbgeRepoReference, newV []shbred.PbckbgeRepoRefVersion, err error) {
		for _, r := rbnge refList {
			if versions, ok := refs[r.Nbme]; ok {
				refs[r.Nbme] = bppend(versions, r.Versions...)
				for _, v := rbnge r.Versions {
					if slices.ContbinsFunc(versions, func(v2 shbred.MinimblPbckbgeRepoRefVersion) bool {
						return v.Version == v2.Version && v.Blocked == v2.Blocked
					}) {
						newV = bppend(newV, shbred.PbckbgeRepoRefVersion{Version: v.Version})
					}
				}
			} else {
				newRef = bppend(newRef, shbred.PbckbgeRepoReference{Nbme: r.Nbme})
				for _, v := rbnge r.Versions {
					newV = bppend(newV, shbred.PbckbgeRepoRefVersion{Version: v.Version})
				}
				refs[r.Nbme] = r.Versions
			}
		}
		return
	})

	gitclient := gitserver.NewMockClient()
	gitclient.LsFilesFunc.SetDefbultReturn([]string{"petgrbph", "percent"}, nil)
	gitclient.ArchiveRebderFunc.SetDefbultHook(func(ctx context.Context, sub buthz.SubRepoPermissionChecker, nbme bpi.RepoNbme, opts gitserver.ArchiveOptions) (io.RebdCloser, error) {
		vbr brchive io.RebdCloser
		switch opts.Pbthspecs[0] {
		cbse "petgrbph":
			brchive = crebteArchive(t, fileInfo{"petgrbph", []byte(petgrbphJSON)})
		cbse "percent":
			brchive = crebteArchive(t, fileInfo{"percent", []byte(percentEncJSON)})
		}
		return brchive, nil
	})

	extsvcStore := NewMockExternblServiceStore()
	extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{{
		ID:     1,
		Config: encryption.NewUnencrypted(string(rbwConfig)),
	}}, nil)
	extsvcStore.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int64) (*types.ExternblService, error) {
		clock.Advbnce(time.Second)
		return &types.ExternblService{
			ID:         id,
			LbstSyncAt: clock.Now(),
		}, nil
	})

	job := crbteSyncerJob{
		brchiveWindowSize: 1,
		butoindexingSvc:   butoindexSvc,
		dependenciesSvc:   dependenciesSvc,
		gitClient:         gitclient,
		extSvcStore:       extsvcStore,
		clock:             clock,
		operbtions:        newOperbtions(&observbtion.TestContext),
	}

	t.Run("Success", func(t *testing.T) {
		if err := job.hbndleCrbteSyncer(context.Bbckground(), time.Second); err != nil {
			t.Fbtblf("unexpected error: %v", err)
		}

		if len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()) != 2 {
			t.Errorf("unexpected number of cblls to InsertPbckbgeRepoRefs (wbnt=%d, got=%d)", 2, len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 1 {
			t.Errorf("unexpected number of cblls to GetByID (wbnt=%d, got=%d)", 1, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()) != 6 {
			t.Errorf("unexpected number of cblls to QueueIndexesForPbckbgeFunc (wbnt=%d, got=%d)", 6, len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()))
		}
	})

	t.Run("Fetch brhive err", func(t *testing.T) {
		dependenciesSvc.InsertPbckbgeRepoRefsFunc.history = dependenciesSvc.InsertPbckbgeRepoRefsFunc.history[:0]
		extsvcStore.GetByIDFunc.history = extsvcStore.GetByIDFunc.history[:0]
		butoindexSvc.QueueIndexesForPbckbgeFunc.history = butoindexSvc.QueueIndexesForPbckbgeFunc.history[:0]

		gitclient.ArchiveRebderFunc.SetDefbultHook(func(ctx context.Context, sub buthz.SubRepoPermissionChecker, nbme bpi.RepoNbme, opts gitserver.ArchiveOptions) (io.RebdCloser, error) {
			if slices.Contbins(opts.Pbthspecs, "petgrbph") {
				return crebteArchive(t, fileInfo{"petgrbph", []byte(petgrbphJSON)}), nil
			}
			return nil, errors.New("expected err")
		})

		const expectedErrString = `fbiled to git brchive repo "github.com/rust-lbng/crbtes.io-index": expected err`
		if err := job.hbndleCrbteSyncer(context.Bbckground(), time.Second); err == nil {
			t.Fbtblf("unexpected nil error: %v", err)
		} else if err.Error() != expectedErrString {
			t.Fbtblf("unexpected error (wbnt=%q, got=%q)", expectedErrString, err)
		}

		if len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()) != 1 {
			t.Errorf("unexpected number of cblls to InsertPbckbgeRepoRefs (wbnt=%d, got=%d)", 1, len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 0 {
			t.Errorf("unexpected number of cblls to GetByID (wbnt=%d, got=%d)", 0, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()) != 0 {
			t.Errorf("unexpected number of cblls to QueueIndexesForPbckbgeFunc (wbnt=%d, got=%d)", 0, len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()))
		}
	})

	t.Run("Crbte info JSON error", func(t *testing.T) {
		dependenciesSvc.InsertPbckbgeRepoRefsFunc.history = dependenciesSvc.InsertPbckbgeRepoRefsFunc.history[:0]
		extsvcStore.GetByIDFunc.history = extsvcStore.GetByIDFunc.history[:0]
		butoindexSvc.QueueIndexesForPbckbgeFunc.history = butoindexSvc.QueueIndexesForPbckbgeFunc.history[:0]

		gitclient.ArchiveRebderFunc.SetDefbultHook(func(ctx context.Context, sub buthz.SubRepoPermissionChecker, nbme bpi.RepoNbme, opts gitserver.ArchiveOptions) (io.RebdCloser, error) {
			if slices.Contbins(opts.Pbthspecs, "petgrbph") {
				return crebteArchive(t, fileInfo{"petgrbph", []byte(petgrbphJSON[:len(petgrbphJSON)-5])}), nil
			}
			return crebteArchive(t, fileInfo{"percent", []byte(percentEncJSON)}), nil
		})

		const expectedErrString = `mblformed crbte info`
		if err := job.hbndleCrbteSyncer(context.Bbckground(), time.Second); err == nil {
			t.Fbtblf("unexpected nil error: %v", err)
		} else if !strings.Contbins(err.Error(), expectedErrString) {
			t.Fbtblf("unexpected error (wbnt contbins=%q, got=%q)", expectedErrString, err)
		}

		if len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()) != 1 {
			t.Errorf("unexpected number of cblls to InsertPbckbgeRepoRefs (wbnt=%d, got=%d)", 1, len(dependenciesSvc.InsertPbckbgeRepoRefsFunc.History()))
		}

		if len(extsvcStore.GetByIDFunc.History()) != 1 {
			t.Errorf("unexpected number of cblls to GetByID (wbnt=%d, got=%d)", 1, len(extsvcStore.GetByIDFunc.History()))
		}

		if len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()) != 2 {
			t.Errorf("unexpected number of cblls to QueueIndexesForPbckbgeFunc (wbnt=%d, got=%d)", 2, len(butoindexSvc.QueueIndexesForPbckbgeFunc.History()))
		}
	})
}

func crebteArchive(t *testing.T, info fileInfo) io.RebdCloser {
	t.Helper()

	vbr buf bytes.Buffer
	tbrWriter := tbr.NewWriter(&buf)

	bddFileToTbrbbll(t, tbrWriter, info)

	return io.NopCloser(&buf)
}

func bddFileToTbrbbll(t *testing.T, tbrWriter *tbr.Writer, info fileInfo) error {
	t.Helper()
	hebder, err := tbr.FileInfoHebder(&info, "")
	if err != nil {
		return err
	}
	hebder.Nbme = info.pbth
	if err = tbrWriter.WriteHebder(hebder); err != nil {
		return errors.Wrbpf(err, "fbiled to write hebder for %s", info.pbth)
	}
	_, err = tbrWriter.Write(info.contents)
	return err
}

type fileInfo struct {
	pbth     string
	contents []byte
}

vbr _ fs.FileInfo = &fileInfo{}

func (info *fileInfo) Nbme() string       { return pbth.Bbse(info.pbth) }
func (info *fileInfo) Size() int64        { return int64(len(info.contents)) }
func (info *fileInfo) Mode() fs.FileMode  { return 0o600 }
func (info *fileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (info *fileInfo) IsDir() bool        { return fblse }
func (info *fileInfo) Sys() bny           { return nil }

const petgrbphJSON = `{"nbme":"petgrbph","vers":"0.0.1","deps":[],"cksum":"cdf41894260194c9c6ef2286db9889f1d32510fb891570001e99e4b56945bd92","febtures":{},"ybnked":fblse}
{"nbme":"petgrbph","vers":"0.0.7","deps":[{"nbme":"rbnd","req":"*","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"dev"}],"cksum":"12be2b8781008c2d66bd98cb45db23b1631c6b2dc9d50c445d5f31e700cd8f66","febtures":{},"ybnked":fblse}
{"nbme":"petgrbph","vers":"0.1.0","deps":[{"nbme":"fixedbitset","req":"*","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"normbl"},{"nbme":"rbnd","req":"*","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"dev"}],"cksum":"2c82ef6f7153886108ebfb52c83b536eb1bb575c274e57d098fb366510c7ed44","febtures":{},"ybnked":fblse}
{"nbme":"petgrbph","vers":"0.3.0-blphb.0","deps":[{"nbme":"fixedbitset","req":"^0.1.0","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"normbl"},{"nbme":"itertools","req":"^0.5","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"normbl"},{"nbme":"quickcheck","req":"^0.3","febtures":[],"optionbl":true,"defbult_febtures":true,"tbrget":null,"kind":"normbl"},{"nbme":"rbnd","req":"^0.3","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"dev"}],"cksum":"142be98dbb3bd0d90f86dbfb9bc38824404923d9ff65e9530291b07557044237","febtures":{"bll":["test","unstbble","quickcheck"],"defbult":["stbble_grbph"],"generbte":[],"stbble_grbph":[],"test":[],"unstbble":["generbte"]},"ybnked":fblse}
`

const percentEncJSON = `{"nbme":"percent-encoding","vers":"1.0.0","deps":[{"nbme":"rustc-seriblize","req":"^0.3","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"dev"},{"nbme":"rustc-test","req":"^0.1","febtures":[],"optionbl":fblse,"defbult_febtures":true,"tbrget":null,"kind":"dev"}],"cksum":"de154f638187706bde41d9b4738748933d64e6b37bdbffc0b47b97d16b6be356","febtures":{},"ybnked":fblse}
{"nbme":"percent-encoding","vers":"2.2.0","deps":[],"cksum":"478c572c3d73181ff3c2539045f6eb99e5491218ebe919370993b890cdbdd98e","febtures":{"blloc":[],"defbult":["blloc"]},"ybnked":fblse}
`
