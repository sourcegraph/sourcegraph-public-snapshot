package localstore

import (
	"fmt"

	"gopkg.in/gorp.v1"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// globalRefs is a DB-backed implementation of the Examples store.
//
// TODO: currently reusing data structures from globalRefs for simplicity,
// change these as it makes sense.
type examples struct{}

func (e *examples) Get(ctx context.Context, opt *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExamplesList, error) {
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
		return &sourcegraph.ExamplesList{RepoRefs: []*sourcegraph.DefRepoRef{}}, nil
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

	// =============== START HACK ===============
	limit := opt.PerPageOrDefault()
	selectedRefs := make(map[string]int32)
	selectedRepoRefs := make([]*sourcegraph.DefRepoRef, limit)
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, rr := range repoRefs {
		res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: rr.Repo})
		if err != nil {
			return nil, err
		}
		paths := make([]string, len(rr.Files))
		for _, f := range rr.Files {
			paths = append(paths, f.Path)
		}
		revRes, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: res.Repo, Rev: ""})
		if err != nil {
			return nil, err
		}

		refs, err := cl.Defs.ListRefs(ctx, &sourcegraph.DefsListRefsOp{
			Def: opt.Def,
			Opt: &sourcegraph.DefListRefsOptions{
				Repo:     res.Repo,
				CommitID: revRes.CommitID,
				Files:    paths,
				ListOptions: sourcegraph.ListOptions{
					Page:    1,
					PerPage: 10000,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		if len(refs.Refs) == 0 {
			continue
		}
		treeEntries := make(map[string][]byte)
		for i, r := range refs.Refs {
			if _, ok := treeEntries[r.File]; !ok {
				entry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
					Entry: sourcegraph.TreeEntrySpec{
						RepoRev: sourcegraph.RepoRevSpec{
							Repo:     res.Repo,
							CommitID: revRes.CommitID,
						},
						Path: r.File,
					},
					Opt: &sourcegraph.RepoTreeGetOptions{},
				})
				if err != nil {
					return nil, err
				}
				treeEntries[r.File] = entry.Contents
			}
			// contents := treeEntries[r.File]
			// after := contents[r.Start:r.End]
			if true {
				selectedRepoRefs = append(selectedRepoRefs, rr)
				selectedRefs[selectedRefKey(rr.Repo, r.File)] = int32(i)
				break
			}
		}
		if len(selectedRepoRefs) >= limit {
			break
		}
	}
	// =============== END HACK ===============

	var examples *sourcegraph.ExamplesList
	if len(selectedRepoRefs) >= limit {
		examples = &sourcegraph.ExamplesList{
			RepoRefs:     repoRefs,
			SelectedRefs: selectedRefs,
		}
	} else {
		if len(repoRefs) > limit {
			repoRefs = repoRefs[:limit]
		}
		examples = &sourcegraph.ExamplesList{
			RepoRefs:     repoRefs,
			SelectedRefs: nil,
		}
	}
	return examples, nil
}

func selectedRefKey(repo, file string) string {
	return fmt.Sprintf("%s/%s", repo, file)
}
