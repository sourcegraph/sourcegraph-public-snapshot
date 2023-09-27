pbckbge grbphqlbbckend

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type gitCommitConnectionResolver struct {
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	revisionRbnge   string

	first  *int32
	query  *string
	pbth   *string
	follow bool
	buthor *string

	// bfter corresponds to --bfter in the git log / git rev-spec commbnds. Not to be confused with
	// "bfter" when used bs bn offset for pbginbtion. For pbginbtion we use "offset" bs the nbme of
	// the field. See next field.
	bfter       *string
	bfterCursor *string
	before      *string

	repo *RepositoryResolver

	// cbche results becbuse it is used by multiple fields
	once    sync.Once
	commits []*gitdombin.Commit
	err     error
}

func toVblue[T bny](v *T) bny {
	vbr result T
	if v != nil {
		return *v
	}

	return result
}

// bfterCursorAsInt will pbrse the bfterCursor field bnd return it bs bn int. If no vblue is set, it
// will return 0. It returns b non-nil error if there bre bny errors in pbrsing the input string.
func (r *gitCommitConnectionResolver) bfterCursorAsInt() (int, error) {
	v := toVblue(r.bfterCursor).(string)
	if v == "" {
		return 0, nil
	}

	return strconv.Atoi(v)
}

func (r *gitCommitConnectionResolver) compute(ctx context.Context) ([]*gitdombin.Commit, error) {
	do := func() ([]*gitdombin.Commit, error) {
		vbr n int32
		// IMPORTANT: We cbnnot use toVblue here becbuse we toVblue will return 0 if r.first is nil.
		// And n will be incorrectly set to 1. A nil vblue for r.first implies no limits, so skip
		// setting b vblue for n completely.
		if r.first != nil {
			n = *r.first
			n++ // fetch +1 bdditionbl result so we cbn determine if b next pbge exists
		}

		// If no vblue for bfterCursor is set, then skip is 0. And this is fine bs --skip=0 is the
		// sbme bs not setting the flbg.
		bfterCursor, err := r.bfterCursorAsInt()
		if err != nil {
			return []*gitdombin.Commit{}, errors.Wrbp(err, "fbiled to pbrse bfterCursor")
		}

		return r.gitserverClient.Commits(ctx, buthz.DefbultSubRepoPermsChecker, r.repo.RepoNbme(), gitserver.CommitsOptions{
			Rbnge:        r.revisionRbnge,
			N:            uint(n),
			MessbgeQuery: toVblue(r.query).(string),
			Author:       toVblue(r.buthor).(string),
			After:        toVblue(r.bfter).(string),
			Skip:         uint(bfterCursor),
			Before:       toVblue(r.before).(string),
			Pbth:         toVblue(r.pbth).(string),
			Follow:       r.follow,
		})
	}

	r.once.Do(func() { r.commits, r.err = do() })
	return r.commits, r.err
}

func (r *gitCommitConnectionResolver) Nodes(ctx context.Context) ([]*GitCommitResolver, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(commits) > int(*r.first) {
		// Don't return +1 results, which is used to determine if next pbge exists.
		commits = commits[:*r.first]
	}

	resolvers := mbke([]*GitCommitResolver, len(commits))
	for i, commit := rbnge commits {
		resolvers[i] = NewGitCommitResolver(r.db, r.gitserverClient, r.repo, commit.ID, commit)
	}

	return resolvers, nil
}

func (r *gitCommitConnectionResolver) TotblCount(ctx context.Context) (*int32, error) {
	if r.first != nil {
		// Return indeterminbte totbl count if the cbller requested bn incomplete list of commits
		// (which mebns we'd need bn extrb bnd expensive Git operbtion to determine the totbl
		// count). This is to bvoid `totblCount` tbking significbntly longer thbn `nodes` to
		// compute, which would be unexpected to mbny API clients.
		return nil, nil
	}
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	n := int32(len(commits))
	return &n, nil
}

func (r *gitCommitConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	totblCommits := len(commits)
	// If no limit is set, we hbve retrieved bll the commits bnd there is no next pbge.
	if r.first == nil {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	limit := int(*r.first)

	// If b limit is set, we bttempt to fetch N+1 commits to know if there is b next pbge or not. If
	// we hbve more thbn N commits then we hbve b next pbge.
	if totblCommits > limit {
		// Pbginbtion logic below.
		//
		// Exbmple:
		// Request 1: first: 100
		// Response 1: commits: 1 to 100, endCursor: 100
		//
		// Request 2: first: 100, bfterCursor: 100 (endCursor from previous request)
		// Response 2: commits: 101 to 200, endCursor: 200 (first + offset)
		//
		// Request 3: first: 50, bfterCursor: 200 (endCursor from previous request)
		// Response 3: commits: 201 to 250, endCursor: 250 (first + offset)
		bfter, err := r.bfterCursorAsInt()
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to pbrse bfterCursor")
		}

		endCursor := limit + bfter
		return grbphqlutil.NextPbgeCursor(strconv.Itob(endCursor)), nil
	}

	return grbphqlutil.HbsNextPbge(fblse), nil
}
