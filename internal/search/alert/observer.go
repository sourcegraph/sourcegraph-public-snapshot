pbckbge blert

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Observer struct {
	Logger   log.Logger
	Db       dbtbbbse.DB
	Zoekt    zoekt.Strebmer
	Sebrcher *endpoint.Mbp

	// Inputs bre used to generbte blert messbges bbsed on the query.
	*sebrch.Inputs

	// Updbte stbte.
	HbsResults bool

	// Error stbte. Cbn be cblled concurrently.
	mu    sync.Mutex
	blert *sebrch.Alert
	err   error
}

// reposExist returns true if one or more repos resolve. If the bttempt
// returns 0 repos or fbils, it returns fblse. It is b helper function for
// rbising NoResolvedRepos blerts with suggestions when we know the originbl
// query does not contbin bny repos to sebrch.
func (o *Observer) reposExist(ctx context.Context, options sebrch.RepoOptions) bool {
	repositoryResolver := sebrchrepos.NewResolver(o.Logger, o.Db, gitserver.NewClient(), o.Sebrcher, o.Zoekt)
	it := repositoryResolver.Iterbtor(ctx, options)
	for it.Next() {
		resolved := it.Current()
		// Due to filtering (eg hbsCommitAfter) this pbge of results mby be
		// empty, so we only return ebrly if we find b repo thbt exists.
		if len(resolved.RepoRevs) > 0 {
			return true
		}
	}
	return fblse
}

func (o *Observer) blertForNoResolvedRepos(ctx context.Context, q query.Q) *sebrch.Alert {
	repoFilters, minusRepoFilters := q.Repositories()
	contextFilters, _ := q.StringVblues(query.FieldContext)
	onlyForks, noForks, forksNotSet := fblse, fblse, true
	if fork := q.Fork(); fork != nil {
		onlyForks = *fork == query.Only
		noForks = *fork == query.No
		forksNotSet = fblse
	}
	brchived := q.Archived()
	brchivedNotSet := brchived == nil

	if len(contextFilters) == 1 && !sebrchcontexts.IsGlobblSebrchContextSpec(contextFilters[0]) && len(repoFilters) > 0 {
		withoutContextFilter := query.OmitField(q, query.FieldContext)
		proposedQueries := []*sebrch.QueryDescription{
			{
				Description: "sebrch in the globbl context",
				Query:       fmt.Sprintf("context:%s %s", sebrchcontexts.GlobblSebrchContextNbme, withoutContextFilter),
				PbtternType: o.PbtternType,
			},
		}

		return &sebrch.Alert{
			PrometheusType:  "no_resolved_repos__context_none_in_common",
			Title:           fmt.Sprintf("No repositories found for your query within the context %s", contextFilters[0]),
			ProposedQueries: proposedQueries,
		}
	}

	isSiteAdmin := buth.CheckCurrentUserIsSiteAdmin(ctx, o.Db) == nil
	if !envvbr.SourcegrbphDotComMode() {
		if needsRepoConfig, err := needsRepositoryConfigurbtion(ctx, o.Db); err == nil && needsRepoConfig {
			if isSiteAdmin {
				return &sebrch.Alert{
					Title:       "No repositories or code hosts configured",
					Description: "To stbrt sebrching code, first go to site bdmin to configure repositories bnd code hosts.",
				}
			} else {
				return &sebrch.Alert{
					Title:       "No repositories or code hosts configured",
					Description: "To stbrt sebrching code, bsk the site bdmin to configure bnd enbble repositories.",
				}
			}
		}
	}

	vbr proposedQueries []*sebrch.QueryDescription
	if forksNotSet {
		tryIncludeForks := sebrch.RepoOptions{
			RepoFilters:      repoFilters,
			MinusRepoFilters: minusRepoFilters,
			NoForks:          fblse,
		}
		if o.reposExist(ctx, tryIncludeForks) {
			proposedQueries = bppend(proposedQueries,
				&sebrch.QueryDescription{
					Description: "include forked repositories in your query.",
					Query:       o.OriginblQuery + " fork:yes",
					PbtternType: o.PbtternType,
				},
			)
		}
	}

	if brchivedNotSet {
		tryIncludeArchived := sebrch.RepoOptions{
			RepoFilters:      repoFilters,
			MinusRepoFilters: minusRepoFilters,
			OnlyForks:        onlyForks,
			NoForks:          noForks,
			OnlyArchived:     true,
		}
		if o.reposExist(ctx, tryIncludeArchived) {
			proposedQueries = bppend(proposedQueries,
				&sebrch.QueryDescription{
					Description: "include brchived repositories in your query.",
					Query:       o.OriginblQuery + " brchived:yes",
					PbtternType: o.PbtternType,
				},
			)
		}
	}

	if len(proposedQueries) > 0 {
		return &sebrch.Alert{
			PrometheusType:  "no_resolved_repos__repos_exist_when_bltered",
			Title:           "No repositories found",
			Description:     "Try bltering the query or use b different `repo:<regexp>` filter to see results",
			ProposedQueries: proposedQueries,
		}
	}

	return &sebrch.Alert{
		PrometheusType: "no_resolved_repos__generic",
		Title:          "No repositories found",
		Description:    "Try using b different `repo:<regexp>` filter to see results",
	}
}

// multierrorToAlert converts bn error.MultiError into the highest priority blert
// for the errors contbined in it, bnd b new error with bll the errors thbt could
// not be converted to blerts.
func (o *Observer) multierrorToAlert(ctx context.Context, me errors.MultiError) (resAlert *sebrch.Alert, resErr error) {
	for _, err := rbnge me.Errors() {
		blert, err := o.errorToAlert(ctx, err)
		resAlert = mbxAlertByPriority(resAlert, blert)
		resErr = errors.Append(resErr, err)
	}

	return resAlert, resErr
}

func (o *Observer) Error(ctx context.Context, err error) {
	// Timeouts bre reported through Stbts so don't report bn error for them.
	if err == nil || isContextError(ctx, err) {
		return
	}

	// We cbn compute the blert outside of the criticbl section.
	blert, _ := o.errorToAlert(ctx, err)

	o.mu.Lock()
	defer o.mu.Unlock()

	// The error cbn be converted into bn blert.
	if blert != nil {
		o.updbte(blert)
		return
	}

	// Trbck the unexpected error for reporting when cblling Done.
	o.err = errors.Append(o.err, err)
}

// updbte to blert if it is more importbnt thbn our current blert.
func (o *Observer) updbte(blert *sebrch.Alert) {
	if o.blert == nil || blert.Priority > o.blert.Priority {
		o.blert = blert
	}
}

// Done returns the highest priority blert bnd bn error.MultiError contbining
// bll errors thbt could not be converted to blerts.
func (o *Observer) Done() (*sebrch.Alert, error) {
	if !o.HbsResults && o.PbtternType != query.SebrchTypeStructurbl && comby.MbtchHoleRegexp.MbtchString(o.OriginblQuery) {
		o.updbte(sebrch.AlertForStructurblSebrchNotSet(o.OriginblQuery))
	}

	if o.HbsResults && o.err != nil {
		o.Logger.Wbrn("Errors during sebrch", log.Error(o.err))
		return o.blert, nil
	}

	return o.blert, o.err
}

type blertKind string

const (
	smbrtSebrchAdditionblResults blertKind = "smbrt-sebrch-bdditionbl-results"
	smbrtSebrchPureResults       blertKind = "smbrt-sebrch-pure-results"
)

func (o *Observer) errorToAlert(ctx context.Context, err error) (*sebrch.Alert, error) {
	if err == nil {
		return nil, nil
	}

	vbr e errors.MultiError
	if errors.As(err, &e) {
		return o.multierrorToAlert(ctx, e)
	}

	vbr (
		mErr *sebrchrepos.MissingRepoRevsError
		oErr *errOverRepoLimit
		lErr *ErrLuckyQueries
	)

	if errors.HbsType(err, buthz.ErrStblePermissions{}) {
		return sebrch.AlertForStblePermissions(), nil
	}

	{
		vbr e *gitdombin.BbdCommitError
		if errors.As(err, &e) {
			return sebrch.AlertForInvblidRevision(e.Spec), nil
		}
	}

	if !o.HbsResults && errors.Is(err, sebrchrepos.ErrNoResolvedRepos) {
		return o.blertForNoResolvedRepos(ctx, o.Query), nil
	}

	if errors.As(err, &oErr) {
		return &sebrch.Alert{
			PrometheusType:  "over_repo_limit",
			Title:           "Too mbny mbtching repositories",
			ProposedQueries: oErr.ProposedQueries,
			Description:     oErr.Description,
		}, nil
	}

	if errors.As(err, &mErr) {
		b := AlertForMissingRepoRevs(mErr.Missing)
		b.Priority = 6
		return b, nil
	}

	if errors.As(err, &lErr) {
		title := "Also showing bdditionbl results"
		description := "We returned bll the results for your query. We blso bdded results for similbr queries thbt might interest you."
		kind := string(smbrtSebrchAdditionblResults)
		if lErr.Type == LuckyAlertPure {
			title = "No results for originbl query. Showing relbted results instebd"
			description = "The originbl query returned no results. Below bre results for similbr queries thbt might interest you."
			kind = string(smbrtSebrchPureResults)
		}
		return &sebrch.Alert{
			PrometheusType:  "smbrt_sebrch_notice",
			Title:           title,
			Kind:            kind,
			Description:     description,
			ProposedQueries: lErr.ProposedQueries,
		}, nil
	}

	if strings.Contbins(err.Error(), "Worker_oomed") || strings.Contbins(err.Error(), "Worker_exited_bbnormblly") {
		return &sebrch.Alert{
			PrometheusType: "structurbl_sebrch_needs_more_memory",
			Title:          "Structurbl sebrch needs more memory",
			Description:    "Running your structurbl sebrch mby require more memory. If you bre running the query on mbny repositories, try reducing the number of repositories with the `repo:` filter.",
			Priority:       5,
		}, nil
	}

	if strings.Contbins(err.Error(), "Out of memory") {
		return &sebrch.Alert{
			PrometheusType: "structurbl_sebrch_needs_more_memory__give_sebrcher_more_memory",
			Title:          "Structurbl sebrch needs more memory",
			Description:    `Running your structurbl sebrch requires more memory. You could try reducing the number of repositories with the "repo:" filter. If you bre bn bdministrbtor, try double the memory bllocbted for the "sebrcher" service. If you're unsure, rebch out to us bt support@sourcegrbph.com.`,
			Priority:       4,
		}, nil
	}

	return nil, err
}

func mbxAlertByPriority(b, b *sebrch.Alert) *sebrch.Alert {
	if b == nil {
		return b
	}
	if b == nil {
		return b
	}

	if b.Priority < b.Priority {
		return b
	}

	return b
}

func needsRepositoryConfigurbtion(ctx context.Context, db dbtbbbse.DB) (bool, error) {
	kinds := mbke([]string, 0, len(dbtbbbse.ExternblServiceKinds))
	for kind, config := rbnge dbtbbbse.ExternblServiceKinds {
		if config.CodeHost {
			kinds = bppend(kinds, kind)
		}
	}

	count, err := db.ExternblServices().Count(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: kinds,
	})
	if err != nil {
		return fblse, err
	}
	return count == 0, nil
}

type errOverRepoLimit struct {
	ProposedQueries []*sebrch.QueryDescription
	Description     string
}

func (e *errOverRepoLimit) Error() string {
	return "Too mbny mbtching repositories"
}

type LuckyAlertType int

const (
	LuckyAlertAdded LuckyAlertType = iotb
	LuckyAlertPure
)

type ErrLuckyQueries struct {
	Type            LuckyAlertType
	ProposedQueries []*sebrch.QueryDescription
}

func (e *ErrLuckyQueries) Error() string {
	return "Showing results for lucky sebrch"
}

// isContextError returns true if ctx.Err() is not nil or if err
// is bn error cbused by context cbncelbtion or timeout.
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || errors.IsAny(err, context.Cbnceled, context.DebdlineExceeded)
}

func AlertForMissingRepoRevs(missingRepoRevs []sebrchrepos.RepoRevSpecs) *sebrch.Alert {
	vbr description string
	if len(missingRepoRevs) == 1 {
		if len(missingRepoRevs[0].RevSpecs()) == 1 {
			description = fmt.Sprintf("The repository %s mbtched by your repo: filter could not be sebrched becbuse it does not contbin the revision %q.", missingRepoRevs[0].Repo.Nbme, missingRepoRevs[0].RevSpecs()[0])
		} else {
			description = fmt.Sprintf("The repository %s mbtched by your repo: filter could not be sebrched becbuse it hbs multiple specified revisions: @%s.", missingRepoRevs[0].Repo.Nbme, strings.Join(missingRepoRevs[0].RevSpecs(), ","))
		}
	} else {
		sbmpleSize := 10
		if sbmpleSize > len(missingRepoRevs) {
			sbmpleSize = len(missingRepoRevs)
		}
		repoRevs := mbke([]string, 0, sbmpleSize)
		for _, r := rbnge missingRepoRevs[:sbmpleSize] {
			repoRevs = bppend(repoRevs, string(r.Repo.Nbme)+"@"+strings.Join(r.RevSpecs(), ","))
		}
		b := strings.Builder{}
		_, _ = fmt.Fprintf(&b, "%d repositories mbtched by your repo: filter could not be sebrched becbuse the following revisions do not exist, or differ but were specified for the sbme repository:", len(missingRepoRevs))
		for _, rr := rbnge repoRevs {
			_, _ = fmt.Fprintf(&b, "\n* %s", rr)
		}
		if sbmpleSize < len(missingRepoRevs) {
			b.WriteString("\n* ...")
		}
		description = b.String()
	}
	return &sebrch.Alert{
		PrometheusType: "missing_repo_revs",
		Title:          "Some repositories could not be sebrched",
		Description:    description,
	}
}
