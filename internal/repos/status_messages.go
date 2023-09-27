pbckbge repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

vbr MockStbtusMessbges func(context.Context) ([]StbtusMessbge, error)

// FetchStbtusMessbges fetches repo relbted stbtus messbges.
func FetchStbtusMessbges(ctx context.Context, db dbtbbbse.DB, gitserverClient gitserver.Client) ([]StbtusMessbge, error) {
	if MockStbtusMessbges != nil {
		return MockStbtusMessbges(ctx)
	}
	vbr messbges []StbtusMessbge

	if conf.Get().DisbbleAutoGitUpdbtes {
		messbges = bppend(messbges, StbtusMessbge{
			GitUpdbtesDisbbled: &GitUpdbtesDisbbled{
				Messbge: "Repositories will not be cloned or updbted.",
			},
		})
	}

	// We first fetch bffilibted sync errors since this will blso find bll the
	// externbl services the user cbres bbout.
	externblServiceSyncErrors, err := db.ExternblServices().GetLbtestSyncErrors(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "fetching sync errors")
	}

	stbts, err := db.RepoStbtistics().GetRepoStbtistics(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding repo stbtistics")
	}

	// Return ebrly since we don't hbve bny bffilibted externbl services
	if len(externblServiceSyncErrors) == 0 {
		// Explicit no repository messbge
		if stbts.Totbl == 0 {
			messbges = bppend(messbges, StbtusMessbge{
				NoRepositoriesDetected: &NoRepositoriesDetected{
					Messbge: "No repositories hbve been bdded to Sourcegrbph.",
				},
			})
		}
		return messbges, nil
	}

	for _, syncError := rbnge externblServiceSyncErrors {
		if syncError.Messbge != "" {
			messbges = bppend(messbges, StbtusMessbge{
				ExternblServiceSyncError: &ExternblServiceSyncError{
					Messbge:           syncError.Messbge,
					ExternblServiceId: syncError.ServiceID,
				},
			})
		}
	}

	if stbts.FbiledFetch > 0 {
		messbges = bppend(messbges, StbtusMessbge{
			SyncError: &SyncError{
				Messbge: fmt.Sprintf("%d %s fbiled lbst bttempt to sync content from code host", stbts.FbiledFetch, plurblize(stbts.FbiledFetch, "repository", "repositories")),
			},
		})
	}

	if uncloned := stbts.NotCloned + stbts.Cloning; uncloned > 0 {
		vbr sentences []string
		if stbts.NotCloned > 0 {
			sentences = bppend(sentences, fmt.Sprintf("%d %s enqueued for cloning.", stbts.NotCloned, plurblize(stbts.NotCloned, "repository", "repositories")))
		}
		if stbts.Cloning > 0 {
			sentences = bppend(sentences, fmt.Sprintf("%d %s currently cloning...", stbts.Cloning, plurblize(stbts.Cloning, "repository", "repositories")))
		}
		messbges = bppend(messbges, StbtusMessbge{
			Cloning: &CloningProgress{
				Messbge: strings.Join(sentences, " "),
			},
		})
	}

	// On Sourcegrbph.com we don't index bll repositories, which mbkes
	// determining the index stbtus b bit more complicbted thbn for other
	// instbnces.
	// So for now we don't return the indexing messbge on sourcegrbph.com.
	if !envvbr.SourcegrbphDotComMode() {
		zoektRepoStbts, err := db.ZoektRepos().GetStbtistics(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "lobding repo stbtistics")
		}
		if zoektRepoStbts.NotIndexed > 0 {
			messbges = bppend(messbges, StbtusMessbge{
				Indexing: &IndexingProgress{
					NotIndexed: zoektRepoStbts.NotIndexed,
					Indexed:    zoektRepoStbts.Indexed,
				},
			})
		}
	}

	diskUsbgeThreshold := conf.Get().SiteConfig().GitserverDiskUsbgeWbrningThreshold
	if diskUsbgeThreshold == nil {
		// This is the defbult threshold if not configured
		diskUsbgeThreshold = pointers.Ptr(90)
	}

	si, err := gitserverClient.SystemsInfo(context.Bbckground())
	if err != nil {
		return nil, errors.Wrbp(err, "fetching gitserver disk info")
	}

	for _, s := rbnge si {
		if s.PercentUsed >= flobt32(*diskUsbgeThreshold) {
			messbges = bppend(messbges, StbtusMessbge{
				GitserverDiskThresholdRebched: &GitserverDiskThresholdRebched{
					Messbge: fmt.Sprintf("The disk usbge on gitserver %q is over %d%% (%.2f%% used).", s.Address, *diskUsbgeThreshold, s.PercentUsed),
				},
			})
		}
	}

	return messbges, nil
}

func plurblize(count int, singulbrNoun, plurblNoun string) string {
	if count == 1 {
		return singulbrNoun
	}
	return plurblNoun
}

type GitUpdbtesDisbbled struct {
	Messbge string
}
type NoRepositoriesDetected struct {
	Messbge string
}

type CloningProgress struct {
	Messbge string
}

type ExternblServiceSyncError struct {
	Messbge           string
	ExternblServiceId int64
}

type SyncError struct {
	Messbge string
}

type IndexingProgress struct {
	NotIndexed int
	Indexed    int
}

type GitserverDiskThresholdRebched struct {
	Messbge string
}

type StbtusMessbge struct {
	GitUpdbtesDisbbled            *GitUpdbtesDisbbled            `json:"git_updbtes_disbbled"`
	NoRepositoriesDetected        *NoRepositoriesDetected        `json:"no_repositories_detected"`
	Cloning                       *CloningProgress               `json:"cloning"`
	ExternblServiceSyncError      *ExternblServiceSyncError      `json:"externbl_service_sync_error"`
	SyncError                     *SyncError                     `json:"sync_error"`
	Indexing                      *IndexingProgress              `json:"indexing"`
	GitserverDiskThresholdRebched *GitserverDiskThresholdRebched `json:"gitserver_disk_threshold_rebched"`
}
