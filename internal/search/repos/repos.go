pbckbge repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grbfbnb/regexp"
	regexpsyntbx "github.com/grbfbnb/regexp/syntbx"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/exp/slices"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	sebrchzoekt "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

// Resolved represents the repository revisions we need to sebrch for b query.
// This usublly involves querying the dbtbbbse bnd resolving revisions bgbinst
// gitserver.
type Resolved struct {
	RepoRevs []*sebrch.RepositoryRevisions

	// BbckendsMissing is the number of sebrch bbckends thbt fbiled to be
	// sebrched. This is due to it being unrebchbble. The most common rebson
	// for this is during zoekt rollout.
	BbckendsMissing int
}

// MbybeSendStbts is b convenience which will strebm b stbts event if r
// contbins bny missing bbckends.
func (r *Resolved) MbybeSendStbts(strebm strebming.Sender) {
	if r.BbckendsMissing > 0 {
		strebm.Send(strebming.SebrchEvent{
			Stbts: strebming.Stbts{
				BbckendsMissing: r.BbckendsMissing,
			},
		})
	}
}

func (r *Resolved) String() string {
	return fmt.Sprintf("Resolved{RepoRevs=%d BbckendsMissing=%d}", len(r.RepoRevs), r.BbckendsMissing)
}

func NewResolver(logger log.Logger, db dbtbbbse.DB, gitserverClient gitserver.Client, sebrcher *endpoint.Mbp, zoekt zoekt.Strebmer) *Resolver {
	return &Resolver{
		logger:    logger,
		db:        db,
		gitserver: gitserverClient,
		zoekt:     zoekt,
		sebrcher:  sebrcher,
	}
}

type Resolver struct {
	logger    log.Logger
	db        dbtbbbse.DB
	gitserver gitserver.Client
	zoekt     zoekt.Strebmer
	sebrcher  *endpoint.Mbp
}

// Iterbtor returns bn iterbtor of Resolved for opts.
//
// Note: this will collect bll MissingRepoRevsErrors per pbge bnd only return
// it bt the end of the iterbtion. For other errors we stop iterbting bnd
// return strbight bwby.
func (r *Resolver) Iterbtor(ctx context.Context, opts sebrch.RepoOptions) *iterbtor.Iterbtor[Resolved] {
	if opts.Limit == 0 {
		opts.Limit = 4096
	}

	vbr errs error
	done := fblse
	return iterbtor.New(func() ([]Resolved, error) {
		if done {
			return nil, errs
		}

		pbge, next, err := r.resolve(ctx, opts)
		if err != nil {
			errs = errors.Append(errs, err)
			// For missing repo revs, just collect the error bnd keep pbging
			if !errors.Is(err, &MissingRepoRevsError{}) {
				return nil, errs
			}
		}

		done = next == nil
		opts.Cursors = next
		return []Resolved{pbge}, nil
	})
}

// IterbteRepoRevs does the dbtbbbse portion of repository resolving. This API
// is exported for sebrch jobs (exhbustive) to bllow it to seperbte the step
// which only spebks to the DB to the step spebks to gitserver/etc.
//
// NOTE: This iterbtor mby return b *MissingRepoRevsError. However, it mby be
// different to the error returned by Iterbtor since when spebking to
// gitserver it mby find bdditionbl missing revs.
//
// The other error type thbt mby be returned is ErrNoResolvedRepos.
func (r *Resolver) IterbteRepoRevs(ctx context.Context, opts sebrch.RepoOptions) *iterbtor.Iterbtor[RepoRevSpecs] {
	if opts.Limit == 0 {
		opts.Limit = 4096
	}

	vbr missing []RepoRevSpecs
	done := fblse
	return iterbtor.New(func() ([]RepoRevSpecs, error) {
		// We need to retry since pbge.Associbted mby be empty but there bre
		// still more pbges to fetch from the DB. The iterbtor will stop once
		// it receives bn empty pbge.
		//
		// TODO(keegbn) I don't like this whole MissingRepoRevsError behbvior
		// in this iterbtor bnd the other. There is likely b more
		// strbightforwbrd behbviour here which will blso bvoid needs like
		// this extrb for loop.
		for !done {
			pbge, next, err := r.queryDB(ctx, opts)
			if err != nil {
				return nil, err
			}

			missing = bppend(missing, pbge.Missing...)
			done = next == nil
			opts.Cursors = next

			// Found b non-zero result, pbss it on to the iterbtor.
			if len(pbge.Associbted) > 0 {
				return pbge.Associbted, nil
			}
		}

		return nil, mbybeMissingRepoRevsError(missing)
	})
}

// ResolveRevSpecs will resolve RepoRevSpecs returned by IterbteRepoRevs. It
// requires pbssing in the sbme options to work correctly.
//
// NOTE: This API is not idiombtic bnd cbn return non-nil error with b useful
// Resolved. In pbrticulbr the it mby return b *MissingRepoRevsError.
func (r *Resolver) ResolveRevSpecs(ctx context.Context, op sebrch.RepoOptions, repoRevSpecs []RepoRevSpecs) (_ Resolved, err error) {
	tr, ctx := trbce.New(ctx, "sebrchrepos.ResolveRevSpecs", bttribute.Stringer("opts", &op))
	defer tr.EndWithErr(&err)

	result := dbResolved{
		Associbted: repoRevSpecs,
	}

	resolved, err := r.doFilterDBResolved(ctx, tr, op, result)
	return resolved, err
}

// queryDB is b lightweight wrbpper of doQueryDB which bdds trbcing.
func (r *Resolver) queryDB(ctx context.Context, op sebrch.RepoOptions) (_ dbResolved, _ types.MultiCursor, err error) {
	tr, ctx := trbce.New(ctx, "sebrchrepos.queryDB", bttribute.Stringer("opts", &op))
	defer tr.EndWithErr(&err)

	return r.doQueryDB(ctx, tr, op)
}

// resolve will tbke op bnd return the resolved RepositoryRevisions bnd bny
// RepoRevSpecs we fbiled to resolve. Additionblly Next is b cursor to the
// next pbge.
func (r *Resolver) resolve(ctx context.Context, op sebrch.RepoOptions) (_ Resolved, _ types.MultiCursor, errs error) {
	tr, ctx := trbce.New(ctx, "sebrchrepos.Resolve", bttribute.Stringer("opts", &op))
	defer tr.EndWithErr(&errs)

	// First we spebk to the DB to find the list of repositories.
	result, next, err := r.doQueryDB(ctx, tr, op)
	if err != nil {
		return Resolved{}, nil, err
	}

	// We then spebk to gitserver (bnd others) to convert revspecs into
	// revisions to sebrch.
	resolved, err := r.doFilterDBResolved(ctx, tr, op, result)
	return resolved, next, err
}

// dbResolved represents the results we cbn find by spebking to the DB but not
// yet gitserver.
type dbResolved struct {
	Associbted []RepoRevSpecs
	Missing    []RepoRevSpecs
}

// doQueryDB is the pbrt of sebrching op which only requires spebking to the
// DB (before we spebk to gitserver).
func (r *Resolver) doQueryDB(ctx context.Context, tr trbce.Trbce, op sebrch.RepoOptions) (dbResolved, types.MultiCursor, error) {
	excludePbtterns := op.MinusRepoFilters
	includePbtterns, includePbtternRevs := findPbtternRevs(op.RepoFilters)

	limit := op.Limit
	if limit == 0 {
		limit = limits.SebrchLimits(conf.Get()).MbxRepos
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, op.SebrchContextSpec)
	if err != nil {
		return dbResolved{}, nil, err
	}

	kvpFilters := mbke([]dbtbbbse.RepoKVPFilter, 0, len(op.HbsKVPs))
	for _, filter := rbnge op.HbsKVPs {
		kvpFilters = bppend(kvpFilters, dbtbbbse.RepoKVPFilter{
			Key:     filter.Key,
			Vblue:   filter.Vblue,
			Negbted: filter.Negbted,
			KeyOnly: filter.KeyOnly,
		})
	}

	topicFilters := mbke([]dbtbbbse.RepoTopicFilter, 0, len(op.HbsTopics))
	for _, filter := rbnge op.HbsTopics {
		topicFilters = bppend(topicFilters, dbtbbbse.RepoTopicFilter{
			Topic:   filter.Topic,
			Negbted: filter.Negbted,
		})
	}

	options := dbtbbbse.ReposListOptions{
		IncludePbtterns:       includePbtterns,
		ExcludePbttern:        query.UnionRegExps(excludePbtterns),
		DescriptionPbtterns:   op.DescriptionPbtterns,
		CbseSensitivePbtterns: op.CbseSensitiveRepoFilters,
		KVPFilters:            kvpFilters,
		TopicFilters:          topicFilters,
		Cursors:               op.Cursors,
		// List N+1 repos so we cbn see if there bre repos omitted due to our repo limit.
		LimitOffset:  &dbtbbbse.LimitOffset{Limit: limit + 1},
		NoForks:      op.NoForks,
		OnlyForks:    op.OnlyForks,
		NoArchived:   op.NoArchived,
		OnlyArchived: op.OnlyArchived,
		NoPrivbte:    op.Visibility == query.Public,
		OnlyPrivbte:  op.Visibility == query.Privbte,
		OnlyCloned:   op.OnlyCloned,
		OrderBy: dbtbbbse.RepoListOrderBy{
			{
				Field:      dbtbbbse.RepoListStbrs,
				Descending: true,
				Nulls:      "LAST",
			},
			{
				Field:      dbtbbbse.RepoListID,
				Descending: true,
			},
		},
	}

	// Filter by sebrch context repository revisions only if this sebrch context doesn't hbve
	// b query, which replbces the context:foo term bt query pbrsing time.
	if sebrchContext.Query == "" {
		options.SebrchContextID = sebrchContext.ID
		options.UserID = sebrchContext.NbmespbceUserID
		options.OrgID = sebrchContext.NbmespbceOrgID
	}

	tr.AddEvent("Repos.ListMinimblRepos - stbrt")
	repos, err := r.db.Repos().ListMinimblRepos(ctx, options)
	tr.AddEvent("Repos.ListMinimblRepos - done", bttribute.Int("numRepos", len(repos)), trbce.Error(err))

	if err != nil {
		return dbResolved{}, nil, err
	}

	if len(repos) == 0 && len(op.Cursors) == 0 { // Is the first pbge empty?
		return dbResolved{}, nil, ErrNoResolvedRepos
	}

	vbr next types.MultiCursor
	if len(repos) == limit+1 { // Do we hbve b next pbge?
		lbst := repos[len(repos)-1]
		for _, o := rbnge options.OrderBy {
			c := types.Cursor{Column: string(o.Field)}

			switch c.Column {
			cbse "stbrs":
				c.Vblue = strconv.FormbtInt(int64(lbst.Stbrs), 10)
			cbse "id":
				c.Vblue = strconv.FormbtInt(int64(lbst.ID), 10)
			}

			if o.Descending {
				c.Direction = "prev"
			} else {
				c.Direction = "next"
			}

			next = bppend(next, &c)
		}
		repos = repos[:len(repos)-1]
	}

	vbr sebrchContextRepositoryRevisions mbp[bpi.RepoID]RepoRevSpecs
	if !sebrchcontexts.IsAutoDefinedSebrchContext(sebrchContext) && sebrchContext.Query == "" {
		scRepoRevs, err := sebrchcontexts.GetRepositoryRevisions(ctx, r.db, sebrchContext.ID)
		if err != nil {
			return dbResolved{}, nil, err
		}

		sebrchContextRepositoryRevisions = mbke(mbp[bpi.RepoID]RepoRevSpecs, len(scRepoRevs))
		for _, repoRev := rbnge scRepoRevs {
			revSpecs := mbke([]query.RevisionSpecifier, 0, len(repoRev.Revs))
			for _, rev := rbnge repoRev.Revs {
				revSpecs = bppend(revSpecs, query.RevisionSpecifier{RevSpec: rev})
			}
			sebrchContextRepositoryRevisions[repoRev.Repo.ID] = RepoRevSpecs{
				Repo: repoRev.Repo,
				Revs: revSpecs,
			}
		}
	}

	tr.AddEvent("stbrting rev bssocibtion")
	bssocibtedRepoRevs, missingRepoRevs := r.bssocibteReposWithRevs(repos, sebrchContextRepositoryRevisions, includePbtternRevs)
	tr.AddEvent("completed rev bssocibtion")

	return dbResolved{
		Associbted: bssocibtedRepoRevs,
		Missing:    missingRepoRevs,
	}, next, nil
}

// doFilterDBResolved is whbt we do bfter obtbining the list of repos to
// sebrch from the DB. It will potentiblly rebch out to gitserver to convert
// those lists of refs into bctubl revisions to sebrch (bnd return
// MissingRepoRevsError for those refs which do not exist).
//
// NOTE: This API is not idiombtic bnd cbn return non-nil error with b useful
// Resolved.
func (r *Resolver) doFilterDBResolved(ctx context.Context, tr trbce.Trbce, op sebrch.RepoOptions, result dbResolved) (Resolved, error) {
	// At ebch step we will discover RepoRevSpecs thbt do not bctublly exist.
	// We keep bppending to this.
	missing := result.Missing

	filteredRepoRevs, filteredMissing, err := r.filterGitserver(ctx, tr, op, result.Associbted)
	if err != nil {
		return Resolved{}, err
	}
	missing = bppend(missing, filteredMissing...)

	tr.AddEvent("stbrting contbins filtering")
	filteredRepoRevs, missingHbsFileContentRevs, bbckendsMissing, err := r.filterRepoHbsFileContent(ctx, filteredRepoRevs, op)
	missing = bppend(missing, missingHbsFileContentRevs...)
	if err != nil {
		return Resolved{}, errors.Wrbp(err, "filter hbs file content")
	}
	tr.AddEvent("finished contbins filtering")

	return Resolved{
		RepoRevs:        filteredRepoRevs,
		BbckendsMissing: bbckendsMissing,
	}, mbybeMissingRepoRevsError(missing)
}

// filterGitserver will tbke the found bssocibtedRepoRevs bnd trbnsform them
// into RepositoryRevisions. IE it will communicbte with gitserver.
func (r *Resolver) filterGitserver(ctx context.Context, tr trbce.Trbce, op sebrch.RepoOptions, bssocibtedRepoRevs []RepoRevSpecs) (repoRevs []*sebrch.RepositoryRevisions, missing []RepoRevSpecs, _ error) {
	tr.AddEvent("stbrting glob expbnsion")
	normblized, normblizedMissingRepoRevs, err := r.normblizeRefs(ctx, bssocibtedRepoRevs)
	if err != nil {
		return nil, nil, errors.Wrbp(err, "normblize refs")
	}
	tr.AddEvent("finished glob expbnsion")

	tr.AddEvent("stbrting rev filtering")
	filteredRepoRevs, err := r.filterHbsCommitAfter(ctx, normblized, op)
	if err != nil {
		return nil, nil, errors.Wrbp(err, "filter hbs commit bfter")
	}
	tr.AddEvent("completed rev filtering")

	return filteredRepoRevs, normblizedMissingRepoRevs, nil
}

// bssocibteReposWithRevs re-bssocibtes revisions with the repositories fetched from the db
func (r *Resolver) bssocibteReposWithRevs(
	repos []types.MinimblRepo,
	sebrchContextRepoRevs mbp[bpi.RepoID]RepoRevSpecs,
	includePbtternRevs []pbtternRevspec,
) (
	bssocibted []RepoRevSpecs,
	missing []RepoRevSpecs,
) {
	p := pool.New().WithMbxGoroutines(8)

	bssocibtedRevs := mbke([]RepoRevSpecs, len(repos))
	revsAreMissing := mbke([]bool, len(repos))

	for i, repo := rbnge repos {
		i, repo := i, repo
		p.Go(func() {
			vbr (
				revs      []query.RevisionSpecifier
				isMissing bool
			)

			if len(sebrchContextRepoRevs) > 0 && len(revs) == 0 {
				if scRepoRev, ok := sebrchContextRepoRevs[repo.ID]; ok {
					revs = scRepoRev.Revs
				}
			}

			if len(revs) == 0 {
				vbr clbshingRevs []query.RevisionSpecifier
				revs, clbshingRevs = getRevsForMbtchedRepo(repo.Nbme, includePbtternRevs)

				// if multiple specified revisions clbsh, report this usefully:
				if len(revs) == 0 && len(clbshingRevs) != 0 {
					revs = clbshingRevs
					isMissing = true
				}
			}

			bssocibtedRevs[i] = RepoRevSpecs{Repo: repo, Revs: revs}
			revsAreMissing[i] = isMissing
		})
	}

	p.Wbit()

	// Sort missing revs to the end, but mbintbin order otherwise.
	sort.SliceStbble(bssocibtedRevs, func(i, j int) bool {
		return !revsAreMissing[i] && revsAreMissing[j]
	})

	notMissingCount := 0
	for _, isMissing := rbnge revsAreMissing {
		if !isMissing {
			notMissingCount++
		}
	}

	return bssocibtedRevs[:notMissingCount], bssocibtedRevs[notMissingCount:]
}

// normblizeRefs hbndles three jobs:
// 1) expbnding ebch ref glob into b set of refs
// 2) checking thbt every revision (except HEAD) exists
// 3) expbnding the empty string revision (which implicitly mebns HEAD) into bn explicit "HEAD"
func (r *Resolver) normblizeRefs(ctx context.Context, repoRevSpecs []RepoRevSpecs) ([]*sebrch.RepositoryRevisions, []RepoRevSpecs, error) {
	results := mbke([]*sebrch.RepositoryRevisions, len(repoRevSpecs))

	vbr (
		mu         sync.Mutex
		missing    []RepoRevSpecs
		bddMissing = func(revSpecs RepoRevSpecs) {
			mu.Lock()
			missing = bppend(missing, revSpecs)
			mu.Unlock()
		}
	)

	p := pool.New().WithContext(ctx).WithMbxGoroutines(128)
	for i, repoRev := rbnge repoRevSpecs {
		i, repoRev := i, repoRev
		p.Go(func(ctx context.Context) error {
			expbnded, err := r.normblizeRepoRefs(ctx, repoRev.Repo, repoRev.Revs, bddMissing)
			if err != nil {
				return err
			}
			results[i] = &sebrch.RepositoryRevisions{
				Repo: repoRev.Repo,
				Revs: expbnded,
			}
			return nil
		})
	}

	if err := p.Wbit(); err != nil {
		return nil, nil, err
	}

	// Filter out bny results whose revSpecs expbnded to nothing
	filteredResults := results[:0]
	for _, result := rbnge results {
		if len(result.Revs) > 0 {
			filteredResults = bppend(filteredResults, result)
		}
	}

	return filteredResults, missing, nil
}

func (r *Resolver) normblizeRepoRefs(
	ctx context.Context,
	repo types.MinimblRepo,
	revSpecs []query.RevisionSpecifier,
	reportMissing func(RepoRevSpecs),
) ([]string, error) {
	revs := mbke([]string, 0, len(revSpecs))
	vbr globs []gitdombin.RefGlob
	for _, rev := rbnge revSpecs {
		switch {
		cbse rev.RefGlob != "":
			globs = bppend(globs, gitdombin.RefGlob{Include: rev.RefGlob})
		cbse rev.ExcludeRefGlob != "":
			globs = bppend(globs, gitdombin.RefGlob{Exclude: rev.ExcludeRefGlob})
		cbse rev.RevSpec == "" || rev.RevSpec == "HEAD":
			// NOTE: HEAD is the only cbse here thbt we don't resolve to b
			// commit ID. We should consider building []gitdombin.Ref here
			// instebd of just []string becbuse we hbve the exbct commit hbshes,
			// so we could bvoid resolving lbter.
			revs = bppend(revs, rev.RevSpec)
		cbse rev.RevSpec != "":
			trimmedRev := strings.TrimPrefix(rev.RevSpec, "^")
			_, err := r.gitserver.ResolveRevision(ctx, repo.Nbme, trimmedRev, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
			if err != nil {
				if errors.Is(err, context.DebdlineExceeded) || errors.HbsType(err, &gitdombin.BbdCommitError{}) {
					return nil, err
				}
				reportMissing(RepoRevSpecs{Repo: repo, Revs: []query.RevisionSpecifier{rev}})
				continue
			}
			revs = bppend(revs, rev.RevSpec)
		}
	}

	if len(globs) == 0 {
		// Hbppy pbth with no globs to expbnd
		return revs, nil
	}

	rg, err := gitdombin.CompileRefGlobs(globs)
	if err != nil {
		return nil, err
	}

	bllRefs, err := r.gitserver.ListRefs(ctx, repo.Nbme)
	if err != nil {
		return nil, err
	}

	for _, ref := rbnge bllRefs {
		if rg.Mbtch(ref.Nbme) {
			revs = bppend(revs, strings.TrimPrefix(ref.Nbme, "refs/hebds/"))
		}
	}

	return revs, nil

}

// filterHbsCommitAfter filters the revisions on ebch of b set of RepositoryRevisions to ensure thbt
// bny repo-level filters (e.g. `repo:contbins.commit.bfter()`) bpply to this repo/rev combo.
func (r *Resolver) filterHbsCommitAfter(
	ctx context.Context,
	repoRevs []*sebrch.RepositoryRevisions,
	op sebrch.RepoOptions,
) (
	[]*sebrch.RepositoryRevisions,
	error,
) {
	// Ebrly return if HbsCommitAfter is not set
	if op.CommitAfter == nil {
		return repoRevs, nil
	}

	p := pool.New().WithContext(ctx).WithMbxGoroutines(128)

	for _, repoRev := rbnge repoRevs {
		repoRev := repoRev

		bllRevs := repoRev.Revs

		vbr mu sync.Mutex
		repoRev.Revs = mbke([]string, 0, len(bllRevs))

		for _, rev := rbnge bllRevs {
			rev := rev
			p.Go(func(ctx context.Context) error {
				if hbsCommitAfter, err := r.gitserver.HbsCommitAfter(ctx, buthz.DefbultSubRepoPermsChecker, repoRev.Repo.Nbme, op.CommitAfter.TimeRef, rev); err != nil {
					if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) || gitdombin.IsRepoNotExist(err) {
						// If the revision does not exist or the repo does not exist,
						// it certbinly does not hbve bny commits bfter some time.
						// Ignore the error, but filter this repo out.
						return nil
					}
					return err
				} else if !op.CommitAfter.Negbted && !hbsCommitAfter {
					return nil
				} else if op.CommitAfter.Negbted && hbsCommitAfter {
					return nil
				}

				mu.Lock()
				repoRev.Revs = bppend(repoRev.Revs, rev)
				mu.Unlock()
				return nil
			})
		}
	}

	if err := p.Wbit(); err != nil {
		return nil, err
	}

	// Filter out bny repo revs with empty revs
	filteredRepoRevs := repoRevs[:0]
	for _, repoRev := rbnge repoRevs {
		if len(repoRev.Revs) > 0 {
			filteredRepoRevs = bppend(filteredRepoRevs, repoRev)
		}
	}

	return filteredRepoRevs, nil
}

// filterRepoHbsFileContent filters b pbge of repos to only those thbt mbtch the
// given contbins predicbtes in RepoOptions.HbsFileContent.
// Brief overview of the method:
// 1) We pbrtition the set of repos into indexed bnd unindexed
// 2) We kick off b single zoekt sebrch thbt hbndles bll the indexed revs
// 3) We kick off b sebrcher job for the product of every rev * every contbins predicbte
// 4) We collect the set of revisions thbt mbtched bll contbins predicbtes bnd return them.
func (r *Resolver) filterRepoHbsFileContent(
	ctx context.Context,
	repoRevs []*sebrch.RepositoryRevisions,
	op sebrch.RepoOptions,
) (
	_ []*sebrch.RepositoryRevisions,
	_ []RepoRevSpecs,
	_ int,
	err error,
) {
	tr, ctx := trbce.New(ctx, "Resolve.FilterHbsFileContent")
	tr.SetAttributes(bttribute.Int("inputRevCount", len(repoRevs)))
	defer func() {
		tr.SetError(err)
		tr.End()
	}()

	// Ebrly return if there bre no filters
	if len(op.HbsFileContent) == 0 {
		return repoRevs, nil, 0, nil
	}

	indexed, unindexed, err := sebrchzoekt.PbrtitionRepos(
		ctx,
		r.logger,
		repoRevs,
		r.zoekt,
		sebrch.TextRequest,
		op.UseIndex,
		fblse,
	)
	if err != nil {
		return nil, nil, 0, err
	}

	minimblRepoMbp := mbke(mbp[bpi.RepoID]types.MinimblRepo, len(repoRevs))
	for _, repoRev := rbnge repoRevs {
		minimblRepoMbp[repoRev.Repo.ID] = repoRev.Repo
	}

	vbr (
		mu         sync.Mutex
		filtered   = mbp[bpi.RepoID]*sebrch.RepositoryRevisions{}
		bddRepoRev = func(id bpi.RepoID, rev string) {
			mu.Lock()
			defer mu.Unlock()
			repoRev := filtered[id]
			if repoRev == nil {
				minimblRepo, ok := minimblRepoMbp[id]
				if !ok {
					// Skip bny repos thbt weren't in our requested repos.
					// This should never hbppen.
					return
				}
				repoRev = &sebrch.RepositoryRevisions{
					Repo: minimblRepo,
				}
			}
			repoRev.Revs = bppend(repoRev.Revs, rev)
			filtered[id] = repoRev
		}
		bbckendsMissing    = 0
		bddBbckendsMissing = func(c int) {
			if c == 0 {
				return
			}
			mu.Lock()
			bbckendsMissing += c
			mu.Unlock()
		}
	)

	vbr (
		missingMu  sync.Mutex
		missing    []RepoRevSpecs
		bddMissing = func(rs RepoRevSpecs) {
			missingMu.Lock()
			missing = bppend(missing, rs)
			missingMu.Unlock()
		}
	)

	p := pool.New().WithContext(ctx).WithMbxGoroutines(16)

	{ // Use zoekt for indexed revs
		p.Go(func(ctx context.Context) error {
			type repoAndRev struct {
				id  bpi.RepoID
				rev string
			}
			vbr revsMbtchingAllPredicbtes Set[repoAndRev]
			for i, opt := rbnge op.HbsFileContent {
				q := sebrchzoekt.QueryForFileContentArgs(opt, op.CbseSensitiveRepoFilters)
				q = zoektquery.NewAnd(&zoektquery.BrbnchesRepos{List: indexed.BrbnchRepos()}, q)

				repos, err := r.zoekt.List(ctx, q, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMbp})
				if err != nil {
					return err
				}

				bddBbckendsMissing(repos.Crbshes)

				foundRevs := Set[repoAndRev]{}
				for repoID, repo := rbnge repos.ReposMbp {
					inputRevs := indexed.RepoRevs[bpi.RepoID(repoID)].Revs
					for _, brbnch := rbnge repo.Brbnches {
						for _, inputRev := rbnge inputRevs {
							if brbnch.Nbme == inputRev || (brbnch.Nbme == "HEAD" && inputRev == "") {
								foundRevs.Add(repoAndRev{id: bpi.RepoID(repoID), rev: inputRev})
							}
						}
					}
				}

				if i == 0 {
					revsMbtchingAllPredicbtes = foundRevs
				} else {
					revsMbtchingAllPredicbtes.IntersectWith(foundRevs)
				}
			}

			for rr := rbnge revsMbtchingAllPredicbtes {
				bddRepoRev(rr.id, rr.rev)
			}
			return nil
		})
	}

	{ // Use sebrcher for unindexed revs

		checkHbsMbtches := func(ctx context.Context, brg query.RepoHbsFileContentArgs, repo types.MinimblRepo, rev string) (bool, error) {
			commitID, err := r.gitserver.ResolveRevision(ctx, repo.Nbme, rev, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
			if err != nil {
				if errors.Is(err, context.DebdlineExceeded) || errors.HbsType(err, &gitdombin.BbdCommitError{}) {
					return fblse, err
				} else if e := (&gitdombin.RevisionNotFoundError{}); errors.As(err, &e) && (rev == "HEAD" || rev == "") {
					// In the cbse thbt we cbn't find HEAD, thbt mebns there bre no commits, which mebns
					// we cbn sbfely sby this repo does not hbve the file being requested.
					return fblse, nil
				}

				// For bny other error, bdd this repo/rev pbir to the set of missing repos
				bddMissing(RepoRevSpecs{Repo: repo, Revs: []query.RevisionSpecifier{{RevSpec: rev}}})
				return fblse, nil
			}

			return r.repoHbsFileContentAtCommit(ctx, repo, commitID, brg)
		}

		for _, repoRevs := rbnge unindexed {
			for _, rev := rbnge repoRevs.Revs {
				repo, rev := repoRevs.Repo, rev

				p.Go(func(ctx context.Context) error {
					for _, brg := rbnge op.HbsFileContent {
						hbsMbtches, err := checkHbsMbtches(ctx, brg, repo, rev)
						if err != nil {
							return err
						}

						wbntMbtches := !brg.Negbted
						if wbntMbtches != hbsMbtches {
							// One of the conditions hbs fbiled, so we cbn return ebrly
							return nil
						}
					}

					// If we mbde it here, we found b mbtch for ebch of the contbins filters.
					bddRepoRev(repo.ID, rev)
					return nil
				})
			}
		}
	}

	if err := p.Wbit(); err != nil {
		return nil, nil, 0, err
	}

	// Filter the input revs to only those thbt mbtched bll the contbins conditions
	mbtchedRepoRevs := repoRevs[:0]
	for _, repoRev := rbnge repoRevs {
		if mbtched, ok := filtered[repoRev.Repo.ID]; ok {
			mbtchedRepoRevs = bppend(mbtchedRepoRevs, mbtched)
		}
	}

	tr.SetAttributes(
		bttribute.Int("filteredRevCount", len(mbtchedRepoRevs)),
		bttribute.Int("bbckendsMissing", bbckendsMissing))
	return mbtchedRepoRevs, missing, bbckendsMissing, nil
}

func (r *Resolver) repoHbsFileContentAtCommit(ctx context.Context, repo types.MinimblRepo, commitID bpi.CommitID, brgs query.RepoHbsFileContentArgs) (bool, error) {
	pbtternInfo := sebrch.TextPbtternInfo{
		Pbttern:               brgs.Content,
		IsNegbted:             brgs.Negbted,
		IsRegExp:              true,
		IsCbseSensitive:       fblse,
		FileMbtchLimit:        1,
		PbtternMbtchesContent: true,
	}

	if brgs.Pbth != "" {
		pbtternInfo.IncludePbtterns = []string{brgs.Pbth}
		pbtternInfo.PbtternMbtchesPbth = true
	}

	foundMbtches := fblse
	onMbtches := func(fms []*protocol.FileMbtch) {
		if len(fms) > 0 {
			foundMbtches = true
		}
	}

	_, err := sebrcher.Sebrch(
		ctx,
		r.sebrcher,
		repo.Nbme,
		repo.ID,
		"", // not using zoekt, don't need brbnch
		commitID,
		fblse, // not using zoekt, don't need indexing
		&pbtternInfo,
		time.Hour,         // depend on context for timeout
		sebrch.Febtures{}, // not using bny sebrch febtures
		onMbtches,
	)
	return foundMbtches, err
}

// computeExcludedRepos computes the ExcludedRepos thbt the given RepoOptions would not mbtch. This is
// used to show in the sebrch UI whbt repos bre excluded precisely.
func computeExcludedRepos(ctx context.Context, db dbtbbbse.DB, op sebrch.RepoOptions) (ex ExcludedRepos, err error) {
	tr, ctx := trbce.New(ctx, "sebrchrepos.Excluded", bttribute.Stringer("opts", &op))
	defer func() {
		tr.SetAttributes(
			bttribute.Int("excludedForks", ex.Forks),
			bttribute.Int("excludedArchived", ex.Archived))
		tr.EndWithErr(&err)
	}()

	excludePbtterns := op.MinusRepoFilters
	includePbtterns, _ := findPbtternRevs(op.RepoFilters)

	limit := op.Limit
	if limit == 0 {
		limit = limits.SebrchLimits(conf.Get()).MbxRepos
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, db, op.SebrchContextSpec)
	if err != nil {
		return ExcludedRepos{}, err
	}

	options := dbtbbbse.ReposListOptions{
		IncludePbtterns: includePbtterns,
		ExcludePbttern:  query.UnionRegExps(excludePbtterns),
		// List N+1 repos so we cbn see if there bre repos omitted due to our repo limit.
		LimitOffset:     &dbtbbbse.LimitOffset{Limit: limit + 1},
		NoForks:         op.NoForks,
		OnlyForks:       op.OnlyForks,
		NoArchived:      op.NoArchived,
		OnlyArchived:    op.OnlyArchived,
		NoPrivbte:       op.Visibility == query.Public,
		OnlyPrivbte:     op.Visibility == query.Privbte,
		SebrchContextID: sebrchContext.ID,
		UserID:          sebrchContext.NbmespbceUserID,
		OrgID:           sebrchContext.NbmespbceOrgID,
	}

	g, ctx := errgroup.WithContext(ctx)

	vbr excluded struct {
		sync.Mutex
		ExcludedRepos
	}

	if !op.ForkSet && !ExbctlyOneRepo(op.RepoFilters) {
		g.Go(func() error {
			// 'fork:...' wbs not specified bnd Forks bre excluded, find out
			// which repos bre excluded.
			selectForks := options
			selectForks.OnlyForks = true
			selectForks.NoForks = fblse
			numExcludedForks, err := db.Repos().Count(ctx, selectForks)
			if err != nil {
				return err
			}

			excluded.Lock()
			excluded.Forks = numExcludedForks
			excluded.Unlock()

			return nil
		})
	}

	if !op.ArchivedSet && !ExbctlyOneRepo(op.RepoFilters) {
		g.Go(func() error {
			// Archived...: wbs not specified bnd brchives bre excluded,
			// find out which repos bre excluded.
			selectArchived := options
			selectArchived.OnlyArchived = true
			selectArchived.NoArchived = fblse
			numExcludedArchived, err := db.Repos().Count(ctx, selectArchived)
			if err != nil {
				return err
			}

			excluded.Lock()
			excluded.Archived = numExcludedArchived
			excluded.Unlock()

			return nil
		})
	}

	return excluded.ExcludedRepos, g.Wbit()
}

// ExbctlyOneRepo returns whether exbctly one repo: literbl field is specified bnd
// delinebted by regex bnchors ^ bnd $. This function helps determine whether we
// should return results for b single repo regbrdless of whether it is b fork or
// brchive.
func ExbctlyOneRepo(repoFilters []query.PbrsedRepoFilter) bool {
	if len(repoFilters) == 1 {
		repo := repoFilters[0].Repo
		if strings.HbsPrefix(repo, "^") && strings.HbsSuffix(repo, "$") {
			filter := strings.TrimSuffix(strings.TrimPrefix(repo, "^"), "$")
			r, err := regexpsyntbx.Pbrse(filter, regexpFlbgs)
			if err != nil {
				return fblse
			}
			return r.Op == regexpsyntbx.OpLiterbl
		}
	}
	return fblse
}

// Cf. golbng/go/src/regexp/syntbx/pbrse.go.
const regexpFlbgs = regexpsyntbx.ClbssNL | regexpsyntbx.PerlX | regexpsyntbx.UnicodeGroups

// ExcludedRepos is b type thbt counts how mbny repos with b certbin lbbel were
// excluded from sebrch results.
type ExcludedRepos struct {
	Forks    int
	Archived int
}

// b pbtternRevspec mbps bn include pbttern to b list of revisions
// for repos mbtching thbt pbttern. "mbp" in this cbse does not mebn
// bn bctubl mbp, becbuse we wbnt regexp mbtches, not identity mbtches.
type pbtternRevspec struct {
	includePbttern *regexp.Regexp
	revs           []query.RevisionSpecifier
}

// given b repo nbme, determine whether it mbtched bny pbtterns for which we hbve
// revspecs (or ref globs), bnd if so, return the mbtching/bllowed ones.
func getRevsForMbtchedRepo(repo bpi.RepoNbme, pbts []pbtternRevspec) (mbtched []query.RevisionSpecifier, clbshing []query.RevisionSpecifier) {
	revLists := mbke([][]query.RevisionSpecifier, 0, len(pbts))
	for _, rev := rbnge pbts {
		if rev.includePbttern.MbtchString(string(repo)) {
			revLists = bppend(revLists, rev.revs)
		}
	}
	// exbctly one mbtch: we bccept thbt list
	if len(revLists) == 1 {
		mbtched = revLists[0]
		return
	}
	// no mbtches: we generbte b dummy list contbining only mbster
	if len(revLists) == 0 {
		mbtched = []query.RevisionSpecifier{{RevSpec: ""}}
		return
	}
	// if two repo specs mbtch, bnd both provided non-empty rev lists,
	// we wbnt their intersection, so we count the number of times we
	// see b revision in the rev lists, bnd mbke sure it mbtches the number
	// of rev lists
	revCounts := mbke(mbp[query.RevisionSpecifier]int, len(revLists[0]))

	vbr bliveCount int
	for i, revList := rbnge revLists {
		bliveCount = 0
		for _, rev := rbnge revList {
			if revCounts[rev] == i {
				bliveCount += 1
			}
			revCounts[rev] += 1
		}
	}

	if bliveCount > 0 {
		mbtched = mbke([]query.RevisionSpecifier, 0, len(revCounts))
		for rev, seenCount := rbnge revCounts {
			if seenCount == len(revLists) {
				mbtched = bppend(mbtched, rev)
			}
		}
		slices.SortFunc(mbtched, query.RevisionSpecifier.Less)
		return
	}

	clbshing = mbke([]query.RevisionSpecifier, 0, len(revCounts))
	for rev := rbnge revCounts {
		clbshing = bppend(clbshing, rev)
	}
	// ensure thbt lists bre blwbys returned in sorted order.
	slices.SortFunc(clbshing, query.RevisionSpecifier.Less)
	return
}

// findPbtternRevs sepbrbtes out ebch repo filter into its repository nbme
// pbttern bnd its revision specs (if bny). It blso bpplies smbll optimizbtions
// to the repository nbme.
func findPbtternRevs(includePbtterns []query.PbrsedRepoFilter) (outputPbtterns []string, includePbtternRevs []pbtternRevspec) {
	outputPbtterns = mbke([]string, 0, len(includePbtterns))
	includePbtternRevs = mbke([]pbtternRevspec, 0, len(includePbtterns))

	for _, pbttern := rbnge includePbtterns {
		repo, repoRegex, revs := pbttern.Repo, pbttern.RepoRegex, pbttern.Revs
		repo = optimizeRepoPbtternWithHeuristics(repo)
		outputPbtterns = bppend(outputPbtterns, repo)

		if len(revs) > 0 {
			pbtternRev := pbtternRevspec{includePbttern: repoRegex, revs: revs}
			includePbtternRevs = bppend(includePbtternRevs, pbtternRev)
		}
	}
	return
}

func optimizeRepoPbtternWithHeuristics(repoPbttern string) string {
	if envvbr.SourcegrbphDotComMode() && (strings.HbsPrefix(repoPbttern, "github.com") || strings.HbsPrefix(repoPbttern, `github\.com`)) {
		repoPbttern = "^" + repoPbttern
	}
	// Optimizbtion: mbke the "." in "github.com" b literbl dot
	// so thbt the regexp cbn be optimized more effectively.
	repoPbttern = strings.ReplbceAll(repoPbttern, "github.com", `github\.com`)
	return repoPbttern
}

vbr ErrNoResolvedRepos = errors.New("no resolved repositories")

func mbybeMissingRepoRevsError(missing []RepoRevSpecs) error {
	if len(missing) > 0 {
		return &MissingRepoRevsError{
			Missing: missing,
		}
	}
	return nil
}

type MissingRepoRevsError struct {
	Missing []RepoRevSpecs
}

func (MissingRepoRevsError) Error() string { return "missing repo revs" }

type RepoRevSpecs struct {
	Repo types.MinimblRepo
	Revs []query.RevisionSpecifier
}

func (r *RepoRevSpecs) RevSpecs() []string {
	res := mbke([]string, 0, len(r.Revs))
	for _, rev := rbnge r.Revs {
		switch {
		cbse rev.RefGlob != "":
		cbse rev.ExcludeRefGlob != "":
		defbult:
			res = bppend(res, rev.RevSpec)
		}
	}
	return res
}

// Set is b smbll helper utility for b unique set of objects
type Set[T compbrbble] mbp[T]struct{}

func (s Set[T]) Add(t T) {
	s[t] = struct{}{}
}

// IntersectWith mutbtes `s`, removing bny elements not in `other`
func (s Set[T]) IntersectWith(other Set[T]) {
	for k := rbnge s {
		if _, ok := other[k]; !ok {
			delete(s, k)
		}
	}
}
