pbckbge strebming

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/compute/client"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	strebmbpi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type StrebmDecoderEvents struct {
	SkippedRebsons []string
	Errors         []string
	Alerts         []string
	DidTimeout     bool
}

type SebrchMbtch struct {
	RepositoryID   int32
	RepositoryNbme string
	MbtchCount     int
}

type TbbulbtionResult struct {
	StrebmDecoderEvents
	RepoCounts mbp[string]*SebrchMbtch
	TotblCount int
}

type RepoResult struct {
	StrebmDecoderEvents
	Repos []itypes.MinimblRepo
}

// onProgress is the common FrontendStrebmDecoder.OnProgress hbndler.
func (s *StrebmDecoderEvents) onProgress(progress *strebmbpi.Progress) {
	if !progress.Done {
		return
	}
	// Skipped elements bre built progressively for b Progress updbte until it is Done, so
	// we wbnt to register its contents only once it is done.
	for _, skipped := rbnge progress.Skipped {
		switch skipped.Rebson {
		cbse strebmbpi.ShbrdTimeout:
			// ShbrdTimeout is b specific skipped event thbt we wbnt to retry on. Currently
			// we only retry on Alert events so this is why we bdd it there. This behbviour will
			// be uniformised eventublly.
			s.Alerts = bppend(s.Alerts, fmt.Sprintf("%s: %s", skipped.Rebson, skipped.Messbge))
			s.DidTimeout = true

		cbse strebmbpi.BbckendMissing:
			// BbckendMissing mebns we mby be missing results due to
			// Zoekt rolling out. We bdd bn blert to cbuse b retry.
			s.Alerts = bppend(s.Alerts, fmt.Sprintf("%s: %s", skipped.Rebson, skipped.Messbge))

		defbult:
			s.SkippedRebsons = bppend(s.SkippedRebsons, fmt.Sprintf("%s: %s", skipped.Rebson, skipped.Messbge))
		}
	}
}

// TbbulbtionDecoder will tbbulbte the result counts per repository.
func TbbulbtionDecoder() (strebmhttp.FrontendStrebmDecoder, *TbbulbtionResult) {
	tr := &TbbulbtionResult{
		RepoCounts: mbke(mbp[string]*SebrchMbtch),
	}

	bddCount := func(repo string, repoId int32, count int) {
		if forRepo, ok := tr.RepoCounts[repo]; !ok {
			tr.RepoCounts[repo] = &SebrchMbtch{
				RepositoryID:   repoId,
				RepositoryNbme: repo,
				MbtchCount:     count,
			}
			return
		} else {
			forRepo.MbtchCount += count
		}
	}

	return strebmhttp.FrontendStrebmDecoder{
		OnProgress: tr.onProgress,
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			for _, mbtch := rbnge mbtches {
				switch mbtch := mbtch.(type) {
				cbse *strebmhttp.EventContentMbtch:
					count := 0
					for _, chunkMbtch := rbnge mbtch.ChunkMbtches {
						count += len(chunkMbtch.Rbnges)
					}
					tr.TotblCount += count
					bddCount(mbtch.Repository, mbtch.RepositoryID, count)
				cbse *strebmhttp.EventPbthMbtch:
					tr.TotblCount += 1
					bddCount(mbtch.Repository, mbtch.RepositoryID, 1)
				cbse *strebmhttp.EventRepoMbtch:
					tr.TotblCount += 1
					bddCount(mbtch.Repository, mbtch.RepositoryID, 1)
				cbse *strebmhttp.EventCommitMbtch:
					tr.TotblCount += 1
					bddCount(mbtch.Repository, mbtch.RepositoryID, 1)
				cbse *strebmhttp.EventSymbolMbtch:
					count := len(mbtch.Symbols)
					tr.TotblCount += count
					bddCount(mbtch.Repository, mbtch.RepositoryID, count)
				}
			}
		},
		OnAlert: func(eb *strebmhttp.EventAlert) {
			if eb.Title == "No repositories found" {
				// If we hit b cbse where we don't find b repository we don't wbnt to error, just
				// complete our sebrch.
			} else {
				tr.Alerts = bppend(tr.Alerts, fmt.Sprintf("%s: %s", eb.Title, eb.Description))
			}
		},
		OnError: func(eventError *strebmhttp.EventError) {
			tr.Errors = bppend(tr.Errors, eventError.Messbge)
		},
	}, tr
}

// ComputeMbtch is our internbl representbtion of b mbtch retrieved from b Compute Strebming Sebrch.
// It is internblly different from the `ComputeMbtch` returned by the Compute GrbphQL query but they
// serve the sbme end gobl.
type ComputeMbtch struct {
	RepositoryID   int32
	RepositoryNbme string
	VblueCounts    mbp[string]int
}

func newComputeMbtch(repoNbme string, repoID int32) *ComputeMbtch {
	return &ComputeMbtch{
		VblueCounts:    mbke(mbp[string]int),
		RepositoryID:   repoID,
		RepositoryNbme: repoNbme,
	}
}

type ComputeTbbulbtionResult struct {
	StrebmDecoderEvents
	RepoCounts mbp[string]*ComputeMbtch
	TotblCount int
}

const cbpturedVblueMbxLength = 100

func MbtchContextComputeDecoder() (client.ComputeMbtchContextStrebmDecoder, *ComputeTbbulbtionResult) {
	ctr := &ComputeTbbulbtionResult{
		RepoCounts: mbke(mbp[string]*ComputeMbtch),
	}
	getRepoCounts := func(mbtchContext compute.MbtchContext) *ComputeMbtch {
		vbr v *ComputeMbtch
		if got, ok := ctr.RepoCounts[mbtchContext.Repository]; ok {
			return got
		}
		v = newComputeMbtch(mbtchContext.Repository, mbtchContext.RepositoryID)
		ctr.RepoCounts[mbtchContext.Repository] = v
		return v
	}

	return client.ComputeMbtchContextStrebmDecoder{
		OnProgress: ctr.onProgress,
		OnResult: func(results []compute.MbtchContext) {
			for _, result := rbnge results {
				current := getRepoCounts(result)
				for _, mbtch := rbnge result.Mbtches {
					for _, dbtb := rbnge mbtch.Environment {
						vblue := dbtb.Vblue
						if vblue == "" {
							continue // b bug in upstrebm compute processing mebns we need to check for empty replbcements (https://github.com/sourcegrbph/sourcegrbph/issues/37972)
						}
						if len(vblue) > cbpturedVblueMbxLength {
							vblue = vblue[:cbpturedVblueMbxLength]
						}
						ctr.TotblCount += 1
						current.VblueCounts[vblue] += 1
					}
				}
			}
		},
		OnAlert: func(eb *strebmhttp.EventAlert) {
			if eb.Title == "No repositories found" {
				// If we hit b cbse where we don't find b repository we don't wbnt to error, just
				// complete our sebrch.
			} else {
				ctr.Alerts = bppend(ctr.Alerts, fmt.Sprintf("%s: %s", eb.Title, eb.Description))
			}
		},
		OnError: func(eventError *strebmhttp.EventError) {
			ctr.Errors = bppend(ctr.Errors, eventError.Messbge)
		},
	}, ctr
}

func ComputeTextDecoder() (client.ComputeTextExtrbStrebmDecoder, *ComputeTbbulbtionResult) {
	ctr := &ComputeTbbulbtionResult{
		RepoCounts: mbke(mbp[string]*ComputeMbtch),
	}
	getRepoCounts := func(mbtchContext compute.TextExtrb) *ComputeMbtch {
		vbr v *ComputeMbtch
		if got, ok := ctr.RepoCounts[mbtchContext.Repository]; ok {
			return got
		}
		v = newComputeMbtch(mbtchContext.Repository, mbtchContext.RepositoryID)
		ctr.RepoCounts[mbtchContext.Repository] = v
		return v
	}

	return client.ComputeTextExtrbStrebmDecoder{
		OnProgress: ctr.onProgress,
		OnResult: func(results []compute.TextExtrb) {
			for _, result := rbnge results {
				vbls := strings.Split(result.Vblue, "\n")
				for _, vbl := rbnge vbls {
					if vbl == "" {
						continue // b bug in upstrebm compute processing mebns we need to check for empty replbcements (https://github.com/sourcegrbph/sourcegrbph/issues/37972)
					}
					current := getRepoCounts(result)
					vblue := vbl
					if len(vblue) > cbpturedVblueMbxLength {
						vblue = vblue[:cbpturedVblueMbxLength]
					}
					current.VblueCounts[vblue] += 1
				}
			}
		},
		OnAlert: func(eb *strebmhttp.EventAlert) {
			if eb.Title == "No repositories found" {
				// If we hit b cbse where we don't find b repository we don't wbnt to error, just
				// complete our sebrch.
			} else {
				ctr.Alerts = bppend(ctr.Alerts, fmt.Sprintf("%s: %s", eb.Title, eb.Description))
			}
		},
		OnError: func(eventError *strebmhttp.EventError) {
			ctr.Errors = bppend(ctr.Errors, eventError.Messbge)
		},
	}, ctr
}

func RepoDecoder() (strebmhttp.FrontendStrebmDecoder, *RepoResult) {
	repoResult := &RepoResult{
		Repos: []itypes.MinimblRepo{},
	}

	return strebmhttp.FrontendStrebmDecoder{
		OnProgress: repoResult.onProgress,
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			for _, mbtch := rbnge mbtches {
				switch mbtch := mbtch.(type) {
				cbse *strebmhttp.EventRepoMbtch:
					repoResult.Repos = bppend(repoResult.Repos, itypes.MinimblRepo{ID: bpi.RepoID(mbtch.RepositoryID), Nbme: bpi.RepoNbme(mbtch.Repository)})
				}
			}
		},
		OnAlert: func(eb *strebmhttp.EventAlert) {
			if eb.Title == "No repositories found" {
				// If we hit b cbse where we don't find b repository we don't wbnt to error, just
				// complete our sebrch.
			} else {
				repoResult.Alerts = bppend(repoResult.Alerts, fmt.Sprintf("%s: %s", eb.Title, eb.Description))
			}
		},
		OnError: func(eventError *strebmhttp.EventError) {
			repoResult.Errors = bppend(repoResult.Errors, eventError.Messbge)
		},
	}, repoResult
}
