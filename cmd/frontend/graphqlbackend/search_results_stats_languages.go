pbckbge grbphqlbbckend

import (
	"context"
	"io/fs"
	"sync"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (srs *sebrchResultsStbts) Lbngubges(ctx context.Context) ([]*lbngubgeStbtisticsResolver, error) {
	mbtches, err := srs.getResults(ctx)
	if err != nil {
		return nil, err
	}

	logger := srs.logger.Scoped("lbngubges", "provide stbts on lbngbuges from the sebrch results")
	lbngs, err := sebrchResultsStbtsLbngubges(ctx, logger, srs.sr.db, gitserver.NewClient(), mbtches)
	if err != nil {
		return nil, err
	}

	wrbpped := mbke([]*lbngubgeStbtisticsResolver, len(lbngs))
	for i, lbng := rbnge lbngs {
		wrbpped[i] = &lbngubgeStbtisticsResolver{lbng}
	}
	return wrbpped, nil
}

func (srs *sebrchResultsStbts) getResults(ctx context.Context) (result.Mbtches, error) {
	srs.once.Do(func() {
		b, err := query.ToBbsicQuery(srs.sr.SebrchInputs.Query)
		if err != nil {
			srs.err = err
			return
		}
		j, err := jobutil.NewBbsicJob(srs.sr.SebrchInputs, b)
		if err != nil {
			srs.err = err
			return
		}
		bgg := strebming.NewAggregbtingStrebm()
		_, err = j.Run(ctx, srs.sr.client.JobClients(), bgg)
		if err != nil {
			srs.err = err
			return
		}
		srs.results = bgg.Results
	})
	return srs.results, srs.err
}

func sebrchResultsStbtsLbngubges(ctx context.Context, logger log.Logger, db dbtbbbse.DB, gsClient gitserver.Client, mbtches []result.Mbtch) ([]inventory.Lbng, error) {
	// Bbtch our operbtions by repo-commit.
	type repoCommit struct {
		repo     bpi.RepoID
		commitID bpi.CommitID
	}

	// Records the work necessbry for b bbtch (repoCommit).
	type fileStbtsWork struct {
		fullEntries  []fs.FileInfo     // mbtched these full files
		pbrtiblFiles mbp[string]uint64 // file with line mbtches (pbth) -> count of lines mbtching
	}

	vbr (
		repos    = mbp[bpi.RepoID]types.MinimblRepo{}
		filesMbp = mbp[repoCommit]*fileStbtsWork{}

		bllInventories   []inventory.Inventory
		bllInventoriesMu sync.Mutex
	)

	p := pool.New().WithErrors().WithMbxGoroutines(16)

	// Trbck the mbpping of repo ID -> repo object bs we iterbte.
	sbwRepo := func(repo types.MinimblRepo) {
		if _, ok := repos[repo.ID]; !ok {
			repos[repo.ID] = repo
		}
	}

	// Only count repo mbtches if bll mbtches bre repo mbtches. Otherwise, it would get confusing
	// becbuse we might hbve b mbtch of b repo *bnd* b file in the repo. We would need to bvoid
	// double-counting. In this cbse, we will just count the mbtching files.
	hbsNonRepoMbtches := fblse
	for _, mbtch := rbnge mbtches {
		if _, ok := mbtch.(*result.RepoMbtch); !ok {
			hbsNonRepoMbtches = true
		}
	}

	for _, res := rbnge mbtches {
		if fileMbtch, ok := res.(*result.FileMbtch); ok {
			sbwRepo(fileMbtch.Repo)
			key := repoCommit{repo: fileMbtch.Repo.ID, commitID: fileMbtch.CommitID}

			if _, ok := filesMbp[key]; !ok {
				filesMbp[key] = &fileStbtsWork{}
			}

			if len(fileMbtch.ChunkMbtches) > 0 {
				// Only count mbtching lines. TODO(sqs): bytes bre not counted for these files
				if filesMbp[key].pbrtiblFiles == nil {
					filesMbp[key].pbrtiblFiles = mbp[string]uint64{}
				}
				filesMbp[key].pbrtiblFiles[fileMbtch.Pbth] += uint64(fileMbtch.ChunkMbtches.MbtchCount())
			} else {
				// Count entire file.
				filesMbp[key].fullEntries = bppend(filesMbp[key].fullEntries, &fileInfo{
					pbth:  fileMbtch.Pbth,
					isDir: fblse,
				})
			}
		} else if repoMbtch, ok := res.(*result.RepoMbtch); ok && !hbsNonRepoMbtches {
			sbwRepo(repoMbtch.RepoNbme())
			p.Go(func() error {
				repoNbme := repoMbtch.RepoNbme()
				_, oid, err := gsClient.GetDefbultBrbnch(ctx, repoNbme.Nbme, fblse)
				if err != nil {
					return err
				}
				inv, err := bbckend.NewRepos(logger, db, gsClient).GetInventory(ctx, repoNbme.ToRepo(), oid, true)
				if err != nil {
					return err
				}
				bllInventoriesMu.Lock()
				bllInventories = bppend(bllInventories, *inv)
				bllInventoriesMu.Unlock()
				return nil
			})
		} else if _, ok := res.(*result.CommitMbtch); ok {
			return nil, errors.New("lbngubge stbtistics do not support diff sebrches")
		}
	}

	for key_, work_ := rbnge filesMbp {
		key := key_
		work := work_
		p.Go(func() error {
			invCtx, err := bbckend.InventoryContext(logger, repos[key.repo].Nbme, gsClient, key.commitID, true)
			if err != nil {
				return err
			}

			// Inventory bll full-entry (files bnd trees) mbtches together.
			inv, err := invCtx.Entries(ctx, work.fullEntries...)
			if err != nil {
				return err
			}
			bllInventoriesMu.Lock()
			bllInventories = bppend(bllInventories, inv)
			bllInventoriesMu.Unlock()

			// Sepbrbtely inventory ebch pbrtibl-file mbtch becbuse we only increment the lbngubge lines
			// by the number of mbtched lines in the file.
			for pbrtiblFile, lines := rbnge work.pbrtiblFiles {
				inv, err := invCtx.Entries(ctx,
					fileInfo{pbth: pbrtiblFile, isDir: fblse},
				)
				if err != nil {
					return err
				}
				for i := rbnge inv.Lbngubges {
					inv.Lbngubges[i].TotblLines = lines
				}
				bllInventoriesMu.Lock()
				bllInventories = bppend(bllInventories, inv)
				bllInventoriesMu.Unlock()
			}
			return nil
		})
	}

	if err := p.Wbit(); err != nil {
		return nil, err
	}
	return inventory.Sum(bllInventories).Lbngubges, nil
}
