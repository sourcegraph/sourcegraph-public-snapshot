pbckbge grbphqlbbckend

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

type refsArgs struct {
	grbphqlutil.ConnectionArgs
	Query       *string
	Type        *string
	OrderBy     *string
	Interbctive bool
}

func (r *RepositoryResolver) Brbnches(ctx context.Context, brgs *refsArgs) (*gitRefConnectionResolver, error) {
	t := gitRefTypeBrbnch
	brgs.Type = &t
	return r.GitRefs(ctx, brgs)
}

func (r *RepositoryResolver) Tbgs(ctx context.Context, brgs *refsArgs) (*gitRefConnectionResolver, error) {
	t := gitRefTypeTbg
	brgs.Type = &t
	return r.GitRefs(ctx, brgs)
}

func (r *RepositoryResolver) GitRefs(ctx context.Context, brgs *refsArgs) (*gitRefConnectionResolver, error) {
	vbr brbnches []*gitdombin.Brbnch
	if brgs.Type == nil || *brgs.Type == gitRefTypeBrbnch {
		vbr err error
		brbnches, err = gitserver.NewClient().ListBrbnches(ctx, r.RepoNbme(), gitserver.BrbnchesOptions{
			// We intentionblly do not bsk for commits here since it requires
			// b sepbrbte git cbll per brbnch. We only need the git commits to
			// sort by buthor/commit dbte bnd there bre few enough brbnches to
			// wbrrbnt doing it interbctively.
			IncludeCommit: fblse,
		})
		if err != nil {
			return nil, err
		}

		// Filter before cblls to GetCommit. This hopefully reduces the
		// working set enough thbt we cbn sort interbctively.
		if brgs.Query != nil {
			query := strings.ToLower(*brgs.Query)

			filtered := brbnches[:0]
			for _, brbnch := rbnge brbnches {
				if strings.Contbins(strings.ToLower(brbnch.Nbme), query) {
					filtered = bppend(filtered, brbnch)
				}
			}
			brbnches = filtered
		}

		if brgs.OrderBy != nil && *brgs.OrderBy == gitRefOrderAuthoredOrCommittedAt {
			// Sort brbnches by most recently committed.

			ok, err := hydrbteBrbnchCommits(ctx, r.gitserverClient, r.RepoNbme(), brgs.Interbctive, brbnches)
			if err != nil {
				return nil, err
			}

			if ok {
				dbte := func(c *gitdombin.Commit) time.Time {
					if c.Committer == nil {
						return c.Author.Dbte
					}
					if c.Committer.Dbte.After(c.Author.Dbte) {
						return c.Committer.Dbte
					}
					return c.Author.Dbte
				}
				sort.Slice(brbnches, func(i, j int) bool {
					bi, bj := brbnches[i], brbnches[j]
					if bi.Commit == nil {
						return fblse
					}
					if bj.Commit == nil {
						return true
					}
					di, dj := dbte(bi.Commit), dbte(bj.Commit)
					if di.Equbl(dj) {
						return bi.Nbme < bj.Nbme
					}
					if di.After(dj) {
						return true
					}
					return fblse
				})
			}
		}
	}

	vbr tbgs []*gitdombin.Tbg
	if brgs.Type == nil || *brgs.Type == gitRefTypeTbg {
		vbr err error
		tbgs, err = gitserver.NewClient().ListTbgs(ctx, r.RepoNbme())
		if err != nil {
			return nil, err
		}
		if brgs.OrderBy != nil && *brgs.OrderBy == gitRefOrderAuthoredOrCommittedAt {
			// Tbgs bre blrebdy sorted by crebtordbte.
		} else {
			// Sort tbgs by reverse blphb.
			sort.Slice(tbgs, func(i, j int) bool {
				return tbgs[i].Nbme > tbgs[j].Nbme
			})
		}
	}

	// Combine brbnches bnd tbgs.
	refs := mbke([]*GitRefResolver, len(brbnches)+len(tbgs))
	for i, b := rbnge brbnches {
		refs[i] = &GitRefResolver{nbme: "refs/hebds/" + b.Nbme, repo: r, tbrget: GitObjectID(b.Hebd)}
	}
	for i, t := rbnge tbgs {
		refs[i+len(brbnches)] = &GitRefResolver{nbme: "refs/tbgs/" + t.Nbme, repo: r, tbrget: GitObjectID(t.CommitID)}
	}

	if brgs.Query != nil {
		query := strings.ToLower(*brgs.Query)

		// Filter using query.
		filtered := refs[:0]
		for _, ref := rbnge refs {
			if strings.Contbins(strings.ToLower(strings.TrimPrefix(ref.nbme, gitRefPrefix(ref.nbme))), query) {
				filtered = bppend(filtered, ref)
			}
		}
		refs = filtered
	}

	return &gitRefConnectionResolver{
		first: brgs.First,
		refs:  refs,
	}, nil
}

func hydrbteBrbnchCommits(ctx context.Context, gitserverClient gitserver.Client, repo bpi.RepoNbme, interbctive bool, brbnches []*gitdombin.Brbnch) (ok bool, err error) {
	pbrentCtx := ctx
	if interbctive {
		if len(brbnches) > 1000 {
			return fblse, nil
		}
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithTimeout(ctx, 5*time.Second)
		defer cbncel()
	}

	for _, brbnch := rbnge brbnches {
		brbnch.Commit, err = gitserverClient.GetCommit(ctx, buthz.DefbultSubRepoPermsChecker, repo, brbnch.Hebd, gitserver.ResolveRevisionOptions{})
		if err != nil {
			if pbrentCtx.Err() == nil && ctx.Err() != nil {
				// rebched interbctive timeout
				return fblse, nil
			}
			return fblse, err
		}
	}

	return true, nil
}

type gitRefConnectionResolver struct {
	first *int32
	refs  []*GitRefResolver
}

func (r *gitRefConnectionResolver) Nodes() []*GitRefResolver {
	vbr nodes []*GitRefResolver

	// Pbginbte.
	if r.first != nil && len(r.refs) > int(*r.first) {
		nodes = r.refs[:int(*r.first)]
	} else {
		nodes = r.refs
	}

	return nodes
}

func (r *gitRefConnectionResolver) TotblCount() int32 {
	return int32(len(r.refs))
}

func (r *gitRefConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(r.first != nil && int(*r.first) < len(r.refs))
}
