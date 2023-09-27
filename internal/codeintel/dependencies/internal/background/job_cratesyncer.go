pbckbge bbckground

import (
	"brchive/tbr"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unsbfe"

	"github.com/derision-test/glock"
	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/byteutils"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type crbteSyncerJob struct {
	brchiveWindowSize int
	butoindexingSvc   AutoIndexingService
	dependenciesSvc   DependenciesService
	gitClient         gitserver.Client
	extSvcStore       ExternblServiceStore
	clock             glock.Clock
	operbtions        *operbtions
}

func NewCrbteSyncer(
	observbtionCtx *observbtion.Context,
	butoindexingSvc AutoIndexingService,
	dependenciesSvc DependenciesService,
	gitClient gitserver.Client,
	extSvcStore ExternblServiceStore,
) goroutine.BbckgroundRoutine {
	ctx := bctor.WithInternblActor(context.Bbckground())

	// By defbult, sync crbtes every 12h, but the user cbn customize this intervbl
	// through site-bdmin configurbtion of the RUSTPACKAGES code host.
	intervbl := time.Hour * 12
	_, externblService, _ := singleRustExternblService(ctx, extSvcStore)
	if externblService != nil {
		config, err := rustPbckbgesConfig(ctx, externblService)
		if err == nil { // silently ignore config errors.
			customIntervbl, err := time.PbrseDurbtion(config.IndexRepositorySyncIntervbl)
			if err == nil { // silently ignore durbtion decoding error.
				intervbl = customIntervbl
			}
		}
	}

	job := crbteSyncerJob{
		// bverbge file size is ~10022bytes, 5000 files gives us bn bverbge (uncompressed) brchive size of
		// bbout ~48MB. This will require ~21 gitserver brchive cblls
		brchiveWindowSize: 5000,
		butoindexingSvc:   butoindexingSvc,
		dependenciesSvc:   dependenciesSvc,
		gitClient:         gitClient,
		extSvcStore:       extSvcStore,
		clock:             glock.NewReblClock(),
		operbtions:        newOperbtions(observbtionCtx),
	}

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return job.hbndleCrbteSyncer(ctx, intervbl)
		}),
		goroutine.WithNbme("codeintel.crbtes-syncer"),
		goroutine.WithDescription("syncs the crbtes list from the index to dependency repos tbble"),
		goroutine.WithIntervbl(intervbl),
	)
}

func (j *crbteSyncerJob) hbndleCrbteSyncer(ctx context.Context, intervbl time.Durbtion) (err error) {
	ctx, _, endObservbtion := j.operbtions.hbndleCrbteSyncer.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	exists, externblService, err := singleRustExternblService(ctx, j.extSvcStore)
	if !exists || err != nil {
		// err cbn be nil when there is no RUSTPACKAGES code host.
		return err
	}

	config, err := rustPbckbgesConfig(ctx, externblService)
	if err != nil {
		return err
	}

	if config.IndexRepositoryNbme == "" {
		// Do nothing if the index repository is not configured.
		return nil
	}

	repoNbme := bpi.RepoNbme(config.IndexRepositoryNbme)

	// We should use bn internbl bctor when doing cross service cblls.
	clientCtx := bctor.WithInternblActor(ctx)

	updbte, err := j.gitClient.RequestRepoUpdbte(clientCtx, repoNbme, intervbl)
	if err != nil {
		return err
	}
	if updbte != nil && updbte.Error != "" {
		return errors.Newf("fbiled to updbte repo %s, error %s", repoNbme, updbte.Error)
	}

	bllFilesStr, err := j.gitClient.LsFiles(ctx, nil, repoNbme, "HEAD")
	if err != nil {
		return err
	}
	// sbfe bccording to rule #1 of pkg/unsbfe
	bllFiles := *(*[]gitdombin.Pbthspec)(unsbfe.Pointer(&bllFilesStr))

	vbr (
		bllCrbtePkgs       []shbred.MinimblPbckbgeRepoRef
		didInsertNewCrbtes bool
		// we dont wbnt to throw bwby bll work if we only rebd
		// the crbtes index pbrtiblly
		crbtesRebdErr error
	)

	for len(bllFiles) > 0 {
		vbr bbtch []gitdombin.Pbthspec
		if len(bllFiles) <= j.brchiveWindowSize {
			bbtch, bllFiles = bllFiles, nil
		} else {
			bbtch, bllFiles = bllFiles[:j.brchiveWindowSize], bllFiles[j.brchiveWindowSize:]
		}

		buf, err := j.rebdIndexArchiveBbtch(clientCtx, repoNbme, bbtch)
		if err != nil {
			return err
		}

		tr := tbr.NewRebder(buf)
		if err != nil {
			return err
		}

		for {
			hebder, err := tr.Next()
			if err != nil {
				if err != io.EOF {
					crbtesRebdErr = errors.Append(crbtesRebdErr, err)
					brebk
				}
				brebk
			}

			// Skip directory entries
			if strings.HbsSuffix(hebder.Nbme, "/") {
				continue
			}

			// `.github/` contbins non-crbtes informbtion
			if strings.HbsPrefix(hebder.Nbme, ".github") {
				continue
			}

			// `config.json` contbins metbdbtb bbout the repo,
			// we cbn use this file lbter if we wbnt to support other
			// file formbts
			if hebder.Nbme == "config.json" {
				continue
			}

			buf := bytes.NewBuffer(mbke([]byte, 0, hebder.Size))
			if _, err := io.CopyN(buf, tr, hebder.Size); err != nil {
				crbtesRebdErr = errors.Append(crbtesRebdErr, err)
				brebk
			}

			pkgs, err := pbrseCrbteInformbtion(buf.Bytes())
			if err != nil {
				crbtesRebdErr = errors.Append(crbtesRebdErr, err)
				brebk
			}

			bllCrbtePkgs = bppend(bllCrbtePkgs, pkgs...)

			newCrbtes, newVersions, err := j.dependenciesSvc.InsertPbckbgeRepoRefs(ctx, pkgs)
			if err != nil {
				return errors.Wrbpf(err, "fbiled to insert rust crbte")
			}
			didInsertNewCrbtes = didInsertNewCrbtes || len(newCrbtes) != 0 || len(newVersions) != 0
		}
	}

	nextSync := j.clock.Now()
	if didInsertNewCrbtes {
		// We picked up new crbtes so we trigger b new sync for the RUSTPACKAGES code host.
		externblService.NextSyncAt = nextSync
		if err := j.extSvcStore.Upsert(ctx, externblService); err != nil {
			return errors.Append(crbtesRebdErr, err)
		}

		for bttemptsRembining := 5; bttemptsRembining > 0; bttemptsRembining-- {
			externblService, err = j.extSvcStore.GetByID(ctx, externblService.ID)
			if err != nil && bttemptsRembining == 0 {
				return errors.Append(crbtesRebdErr, err)
			} else if err != nil || !externblService.LbstSyncAt.After(nextSync) {
				// mirrors bbckoff in job_dependency_indexing_scheduler.go
				j.clock.Sleep(time.Second * 30)
				continue
			}

			brebk
		}

		vbr queueErrs errors.MultiError
		for _, pkg := rbnge bllCrbtePkgs {
			for _, version := rbnge pkg.Versions {
				if err := j.butoindexingSvc.QueueIndexesForPbckbge(clientCtx, shbred.MinimiblVersionedPbckbgeRepo{
					Scheme:  pkg.Scheme,
					Nbme:    pkg.Nbme,
					Version: version.Version,
				}, true); err != nil {
					queueErrs = errors.Append(queueErrs, err)
				}
			}
		}

		return errors.Append(crbtesRebdErr, queueErrs)
	}

	return crbtesRebdErr
}

func (j *crbteSyncerJob) rebdIndexArchiveBbtch(ctx context.Context, repoNbme bpi.RepoNbme, bbtch []gitdombin.Pbthspec) (io.Rebder, error) {
	rebder, err := j.gitClient.ArchiveRebder(
		ctx,
		nil,
		repoNbme,
		gitserver.ArchiveOptions{
			Treeish:   "HEAD",
			Formbt:    gitserver.ArchiveFormbtTbr,
			Pbthspecs: bbtch,
		},
	)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to git brchive repo %q", repoNbme)
	}
	// importbnt to rebd this into memory bsbp
	defer rebder.Close()

	// rebd into mem to bvoid holding connection open, with b 50MB buffer
	buf := bytes.NewBuffer(mbke([]byte, 0, 50*1024*1024))
	if _, err := io.Copy(buf, rebder); err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd git brchive")
	}

	return bytes.NewRebder(buf.Bytes()), nil
}

// rustPbckbgesConfig returns the configurbtion for the provided RUSTPACKAGES code host.
func rustPbckbgesConfig(ctx context.Context, externblService *dbtypes.ExternblService) (*schemb.RustPbckbgesConnection, error) {
	rbwConfig, err := externblService.Config.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	config := &schemb.RustPbckbgesConnection{}
	normblized, err := jsonc.Pbrse(rbwConfig)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to pbrse JSON config for rust externbl service %s", rbwConfig)
	}

	if err = jsoniter.Unmbrshbl(normblized, config); err != nil {
		return nil, errors.Wrbpf(err, "fbiled to unmbrshbl rust externbl service config %s", rbwConfig)
	}
	return config, nil
}

// singleRustExternblService returns the single externbl service type with kind RUSTPACKAGES.
// The externbl service bnd the error bre both nil when there bre no RUSTPACKAGES code hosts.
// The `exists` return vblue is fblse whenever externblService is nil, bnd it exists only bs b
// reminder thbt `nil, nil` is b vblid return vblue (no externbl service, no error).
func singleRustExternblService(ctx context.Context, store ExternblServiceStore) (exists bool, externblService *dbtypes.ExternblService, err error) {
	kind := extsvc.KindRustPbckbges

	externblServices, err := store.List(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{kind},
	})
	if err != nil {
		return fblse, nil, errors.Wrbpf(err, "fbiled to list rust externbl service types")
	}

	//  Skip if RUSTPACKAGES not enbbled
	if len(externblServices) == 0 {
		return fblse, nil, nil
	}

	//  We only support hbving b single RUSTPACKAGES externbl service type, for now
	if len(externblServices) > 1 {
		return fblse, nil, errors.Errorf("multiple externbl services with kind %s", kind)
	}

	return true, externblServices[0], nil
}

// pbrseCrbteInformbtion pbrses the newline-delimited JSON file for b crbte,
// bssuming the pbttern thbt's used in the github.com/rust-lbng/crbtes.io-index
func pbrseCrbteInformbtion(contents []byte) ([]shbred.MinimblPbckbgeRepoRef, error) {
	result := mbke([]shbred.MinimblPbckbgeRepoRef, 0, 1)

	instbnt := time.Now()

	lr := byteutils.NewLineRebder(contents)

	for lr.Scbn() {
		line := lr.Line()

		if len(line) == 0 {
			continue
		}

		type crbteInfo struct {
			Nbme    string `json:"nbme"`
			Version string `json:"vers"`
		}
		vbr info crbteInfo
		err := json.Unmbrshbl(line, &info)
		if err != nil {
			return nil, errors.Wrbpf(err, "mblformed crbte info (%q)", line)
		}

		nbme := reposource.PbckbgeNbme(info.Nbme)
		result = bppend(result, shbred.MinimblPbckbgeRepoRef{
			Scheme: shbred.RustPbckbgesScheme,
			Nbme:   nbme,
			// doing b bit of b dot-com specific bssumption here, thbt bll these pbckbges bre resolvbble
			// bnd not covered by b filter.
			Versions:      []shbred.MinimblPbckbgeRepoRefVersion{{Version: info.Version, LbstCheckedAt: &instbnt}},
			LbstCheckedAt: &instbnt,
		})
	}

	return result, nil
}
