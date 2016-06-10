package localstore

import (
	"fmt"
	"sort"

	"gopkg.in/gorp.v1"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// globalRefs is a DB-backed implementation of the Examples store.
//
// TODO: currently reusing data structures from globalRefs for simplicity,
// change these as it makes sense.
type examples struct{}

func (e *examples) Get(ctx context.Context, opt *sourcegraph.DefsListExamplesOp) (*sourcegraph.RefLocationsList, error) {
	defRepo, err := (&repos{}).Get(ctx, opt.Def.Repo)
	if err != nil {
		return nil, err
	}
	defRepoPath := defRepo.URI

	if opt == nil {
		opt = &sourcegraph.DefsListExamplesOp{}
	}

	defKeyID, err := graphDBH(ctx).SelectInt(
		"SELECT id FROM def_keys WHERE repo=$1 AND unit_type=$2 AND unit=$3 AND path=$4",
		defRepoPath, opt.Def.UnitType, opt.Def.Unit, opt.Def.Path)
	if err != nil {
		return nil, err
	} else if defKeyID == 0 {
		// DefKey was not found
		return &sourcegraph.RefLocationsList{RepoRefs: []*sourcegraph.DefRepoRef{}}, nil
	}

	// dbExamplesResult holds the result of the SELECT query for fetching examples.
	type dbExamplesResult struct {
		Repo  string
		File  string
		Count int
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	// Over-compensate the amount of refs we fetch since some of them may be
	// filtered out by filterVisibleRepos below.
	//
	// TODO limit this query in a more reliable way.
	const rowLimit = 100
	sql := `
SELECT DISTINCT ON (repo) repo, file, count
FROM global_refs_new
WHERE def_key_id=` + arg(defKeyID) + fmt.Sprintf(" LIMIT %s", arg(rowLimit))

	var examplesResult []*dbExamplesResult
	if _, err := graphDBH(ctx).Select(&examplesResult, sql, args...); err != nil {
		return nil, err
	}

	// Sort examples into groups by repo.
	var repoRefs []*sourcegraph.DefRepoRef
	refsByRepo := make(map[string]*sourcegraph.DefRepoRef)
	for _, r := range examplesResult {
		if _, ok := refsByRepo[r.Repo]; !ok {
			refsByRepo[r.Repo] = &sourcegraph.DefRepoRef{
				Repo: r.Repo,
			}
			repoRefs = append(repoRefs, refsByRepo[r.Repo])
		}
		if r.File != "" {
			refsByRepo[r.Repo].Files = append(refsByRepo[r.Repo].Files, &sourcegraph.DefFileRef{
				Path: r.File,
			})
		}
	}

	// SECURITY: filter private repos this user doesn't have access to.
	repoRefs, err = filterVisibleRepos(ctx, repoRefs)
	if err != nil {
		return nil, err
	}

	limit := opt.PerPageOrDefault()
	if len(repoRefs) > limit {
		repoRefs = repoRefs[:limit]
	}

	// Return Files in a consistent order
	for _, r := range repoRefs {
		sort.Sort(defFileRefByScore(r.Files))
	}

	return &sourcegraph.RefLocationsList{
		RepoRefs: repoRefs,
	}, nil
}
