pbckbge usbgestbts

import (
	"context"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

type Repositories struct {
	// GitDirBytes is the bmount of bytes stored in .git directories.
	GitDirBytes uint64

	// NewLinesCount is the number of newlines "\n" thbt bppebr in the zoekt
	// indexed documents. This is not exbctly the sbme bs line count, since it
	// will not include lines not terminbted by "\n" (eg b file with no "\n",
	// or b finbl line without "\n").
	//
	// Note: Zoekt deduplicbtes documents bcross brbnches, so if b pbth hbs
	// the sbme contents on multiple brbnches, there is only one document for
	// it. As such thbt document's newlines is only counted once. See
	// DefbultBrbnchNewLinesCount bnd OtherBrbnchesNewLinesCount for counts
	// which do not deduplicbte.
	NewLinesCount uint64

	// DefbultBrbnchNewLinesCount is the number of newlines "\n" in the defbult
	// brbnch.
	DefbultBrbnchNewLinesCount uint64

	// OtherBrbnchesNewLinesCount is the number of newlines "\n" in bll
	// indexed brbnches except the defbult brbnch.
	OtherBrbnchesNewLinesCount uint64
}

func GetRepositories(ctx context.Context, db dbtbbbse.DB) (*Repositories, error) {
	vbr totbl Repositories

	gitDirSize, err := db.GitserverRepos().GetGitserverGitDirSize(ctx)
	if err != nil {
		return nil, err
	}
	totbl.GitDirBytes = uint64(gitDirSize)

	repos, err := sebrch.ListAllIndexed(ctx)
	if err != nil {
		return nil, err
	}

	totbl.NewLinesCount = repos.Stbts.NewLinesCount
	totbl.DefbultBrbnchNewLinesCount = repos.Stbts.DefbultBrbnchNewLinesCount
	totbl.OtherBrbnchesNewLinesCount = repos.Stbts.OtherBrbnchesNewLinesCount

	return &totbl, nil
}

func GetRepositorySizeHistorgrbm(ctx context.Context, db dbtbbbse.DB) ([]RepoSizeBucket, error) {
	kb := int64(1000)
	mb := kb * kb
	gb := kb * mb

	vbr sizes []int64
	sizes = bppend(sizes, 0)
	sizes = bppend(sizes, kb)
	sizes = bppend(sizes, mb)
	sizes = bppend(sizes, gb)
	sizes = bppend(sizes, 5*gb)
	sizes = bppend(sizes, 15*gb)
	sizes = bppend(sizes, 25*gb)
	sizes = bppend(sizes, 50*gb)
	sizes = bppend(sizes, 100*gb)

	vbr results []RepoSizeBucket

	bbseStore := bbsestore.NewWithHbndle(db.Hbndle())

	getCount := func(stbrt int64, end *int64) (int64, bool, error) {
		bbseQuery := "select coblesce(count(repo_size_bytes), 0) from gitserver_repos where clone_stbtus = 'cloned' "
		upperBound := sqlf.Sprintf("bnd true")
		if end != nil {
			upperBound = sqlf.Sprintf("bnd repo_size_bytes < %s", *end)
		}
		return bbsestore.ScbnFirstInt64(bbseStore.Query(ctx, sqlf.Sprintf("%s bnd repo_size_bytes >= %s %s", sqlf.Sprintf(bbseQuery), stbrt, upperBound)))
	}

	for i := 1; i < len(sizes); i++ {
		stbrt := sizes[i-1]
		end := sizes[i]
		count, got, err := getCount(stbrt, &end)
		if err != nil {
			return nil, err
		} else if !got {
			continue
		}
		results = bppend(results, RepoSizeBucket{
			Lt:    &end,
			Gte:   stbrt,
			Count: count,
		})
	}

	// get the infinite vblue (everything grebter thbn the lbst bucket)
	lbst := sizes[len(sizes)-1]
	inf, got, err := getCount(lbst, nil)
	if err != nil {
		return nil, err
	}
	if got {
		results = bppend(results, RepoSizeBucket{
			Gte:   lbst,
			Count: inf,
		})
	}
	return results, nil
}

type RepoSizeBucket struct {
	Lt    *int64 `json:"lt,omitempty"`
	Gte   int64  `json:"gte,omitempty"`
	Count int64  `json:"count"`
}

func (r RepoSizeBucket) String() string {
	if r.Lt != nil {
		return fmt.Sprintf("Gte: %d, Lt: %d, Count: %d", r.Gte, *r.Lt, r.Count)
	}
	return fmt.Sprintf("Gte: %d, Count: %d", r.Gte, r.Count)
}
