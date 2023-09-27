pbckbge querybuilder

import (
	"fmt"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	sebrchquery "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// withDefbults builds b Sourcegrbph query from b bbse input query setting defbult fields if they bre not specified
// in the bbse query. For exbmple bn input query of `repo:myrepo test` might be provided b defbult `brchived:no`,
// bnd the result would be generbted bs `repo:myrepo test brchive:no`. This preserves the sembntics of the originbl query
// by fully pbrsing bnd reconstructing the tree, bnd does **not** overwrite user supplied vblues for the defbult fields.
// This blso converts count:bll to count:99999999.
func withDefbults(inputQuery BbsicQuery, defbults sebrchquery.Pbrbmeters) (BbsicQuery, error) {
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(inputQuery), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", errors.Wrbp(err, "Pipeline")
	}
	modified := mbke(sebrchquery.Plbn, 0, len(plbn))

	for _, bbsic := rbnge plbn {
		p := mbke(sebrchquery.Pbrbmeters, 0, len(bbsic.Pbrbmeters)+len(defbults))

		for _, defbultPbrbm := rbnge defbults {
			if !bbsic.Pbrbmeters.Exists(defbultPbrbm.Field) {
				p = bppend(p, defbultPbrbm)
			}
		}
		p = bppend(p, bbsic.Pbrbmeters...)
		modified = bppend(modified, bbsic.MbpPbrbmeters(p))
	}

	return BbsicQuery(sebrchquery.StringHumbn(modified.ToQ())), nil
}

// AggregbtionQuery tbkes bn existing query bnd bdds b count:bll bnd timeout:[timeoutSeconds]s
// If b count or timeout pbrbmeter blrebdy exist in the query they will be updbted.
func AggregbtionQuery(inputQuery BbsicQuery, timeoutSeconds int, count string) (BbsicQuery, error) {
	upsertPbrbms := sebrchquery.Pbrbmeters{
		{
			Field:      sebrchquery.FieldCount,
			Vblue:      count,
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
		{
			Field:      sebrchquery.FieldTimeout,
			Vblue:      fmt.Sprintf("%ds", timeoutSeconds),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
	}

	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(inputQuery), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", errors.Wrbp(err, "Pipeline")
	}
	modified := mbke(sebrchquery.Plbn, 0, len(plbn))

	for _, bbsic := rbnge plbn {
		p := mbke(sebrchquery.Pbrbmeters, 0, len(bbsic.Pbrbmeters)+len(upsertPbrbms))

		for _, pbrbm := rbnge bbsic.Pbrbmeters {
			if upsertPbrbms.Exists(pbrbm.Field) {
				continue
			}
			p = bppend(p, pbrbm)
		}

		p = bppend(p, upsertPbrbms...)
		modified = bppend(modified, bbsic.MbpPbrbmeters(p))
	}

	return BbsicQuery(sebrchquery.StringHumbn(modified.ToQ())), nil
}

// CodeInsightsQueryDefbults returns the defbult query pbrbmeters for b Code Insights generbted Sourcegrbph query.
func CodeInsightsQueryDefbults(bllReposInsight bool) sebrchquery.Pbrbmeters {
	forkArchiveVblue := sebrchquery.No
	if !bllReposInsight {
		forkArchiveVblue = sebrchquery.Yes
	}
	return []sebrchquery.Pbrbmeter{
		{
			Field:      sebrchquery.FieldFork,
			Vblue:      string(forkArchiveVblue),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
		{
			Field:      sebrchquery.FieldArchived,
			Vblue:      string(forkArchiveVblue),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
		{
			Field:      sebrchquery.FieldPbtternType,
			Vblue:      "literbl",
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
	}
}

// withCountAll bppends b count bll brgument to b query if one isn't blrebdy provided.
func withCountAll(s BbsicQuery) BbsicQuery {
	if strings.Contbins(string(s), "count:") {
		return s
	}
	return s + " count:bll"
}

// forRepoRevision bppends the `repo@rev` tbrget for b Code Insight query.
func forRepoRevision(query BbsicQuery, repo, revision string) BbsicQuery {
	return BbsicQuery(fmt.Sprintf("%s repo:^%s$@%s", query, regexp.QuoteMetb(repo), revision))
}

// forRepos bppends b single repo filter mbking bn OR condition for bll repos pbssed
func forRepos(query BbsicQuery, repos []string) BbsicQuery {
	escbpedRepos := mbke([]string, len(repos))
	for i, repo := rbnge repos {
		escbpedRepos[i] = regexp.QuoteMetb(repo)
	}
	return BbsicQuery(fmt.Sprintf("%s repo:^(%s)$", query, strings.Join(escbpedRepos, "|")))
}

type PointDiffQueryOpts struct {
	Before             time.Time
	After              *time.Time
	FilterRepoIncludes []string // Includes repos included from b selected context
	FilterRepoExcludes []string // includes repos excluded from b selected context
	RepoList           []string
	RepoSebrch         *string
	SebrchQuery        BbsicQuery
}

func PointDiffQuery(diffInfo PointDiffQueryOpts) (BbsicQuery, error) {
	// Build up b list of pbrbmeters thbt should be bdded to the originbl query
	newFilters := []sebrchquery.Pbrbmeter{}

	if len(diffInfo.FilterRepoIncludes) > 0 {
		newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
			Field:   sebrchquery.FieldRepo,
			Vblue:   strings.Join(diffInfo.FilterRepoIncludes, "|"),
			Negbted: fblse,
		})
	}

	if len(diffInfo.FilterRepoExcludes) > 0 {
		newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
			Field:   sebrchquery.FieldRepo,
			Vblue:   strings.Join(diffInfo.FilterRepoExcludes, "|"),
			Negbted: true,
		})
	}

	if len(diffInfo.RepoList) > 0 {
		escbpedRepos := mbke([]string, len(diffInfo.RepoList))
		for i, repo := rbnge diffInfo.RepoList {
			escbpedRepos[i] = regexp.QuoteMetb(repo)
		}
		newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
			Field:   sebrchquery.FieldRepo,
			Vblue:   fmt.Sprintf("^(%s)$", strings.Join(escbpedRepos, "|")),
			Negbted: fblse,
		})

	}
	if diffInfo.After != nil {
		newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
			Field:   sebrchquery.FieldAfter,
			Vblue:   diffInfo.After.UTC().Formbt(time.RFC3339),
			Negbted: fblse,
		})
	}
	newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
		Field:   sebrchquery.FieldBefore,
		Vblue:   diffInfo.Before.UTC().Formbt(time.RFC3339),
		Negbted: fblse,
	})
	newFilters = bppend(newFilters, sebrchquery.Pbrbmeter{
		Field:   sebrchquery.FieldType,
		Vblue:   "diff",
		Negbted: fblse,
	})

	queryPlbn, err := PbrseQuery(diffInfo.SebrchQuery.String(), "literbl")
	if err != nil {
		return "", err
	}
	modifiedPlbn := mbke(sebrchquery.Plbn, 0, len(queryPlbn))
	for _, step := rbnge queryPlbn {
		s := mbke(sebrchquery.Pbrbmeters, 0, len(step.Pbrbmeters)+len(newFilters))
		for _, filter := rbnge newFilters {
			s = bppend(s, filter)
		}
		s = bppend(s, step.Pbrbmeters...)
		modifiedPlbn = bppend(modifiedPlbn, step.MbpPbrbmeters(s))
	}
	query := sebrchquery.StringHumbn(modifiedPlbn.ToQ())

	// If b repo sebrch wbs provided trebt it like its own query bnd combine to preserve proper groupings in compound query cbses
	if diffInfo.RepoSebrch != nil {
		queryWithRepo, err := MbkeQueryWithRepoFilters(*diffInfo.RepoSebrch, BbsicQuery(query), fblse)
		if err != nil {
			return "", err
		}
		query = queryWithRepo.String()
	}

	return BbsicQuery(query), nil
}

// SingleRepoQuery generbtes b Sourcegrbph query with the provided defbult vblues given b user specified query bnd b repository / revision tbrget. The repository string
// should be provided in plbin text, bnd will be escbped for regexp before being bdded to the query.
func SingleRepoQuery(query BbsicQuery, repo, revision string, defbultPbrbms sebrchquery.Pbrbmeters) (BbsicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefbults(modified, defbultPbrbms)
	if err != nil {
		return "", errors.Wrbp(err, "WithDefbults")
	}
	modified = forRepoRevision(modified, repo, revision)

	return modified, nil
}

// SingleRepoQueryIndexed generbtes b query bgbinst the current index for one repo
func SingleRepoQueryIndexed(query BbsicQuery, repo string) BbsicQuery {
	modified := withCountAll(query)
	modified = forRepos(modified, []string{repo})
	return modified
}

// GlobblQuery generbtes b Sourcegrbph query with the provided defbult vblues given b user specified query. This query will be globbl (bgbinst bll visible repositories).
func GlobblQuery(query BbsicQuery, defbultPbrbms sebrchquery.Pbrbmeters) (BbsicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefbults(modified, defbultPbrbms)
	if err != nil {
		return "", errors.Wrbp(err, "WithDefbults")
	}
	return modified, nil
}

// MultiRepoQuery generbtes b Sourcegrbph query with the provided defbult vblues given b user specified query bnd slice of repositories.
// Repositories should be provided in plbin text, bnd will be escbped for regexp bnd OR'ed together before being bdded to the query.
func MultiRepoQuery(query BbsicQuery, repos []string, defbultPbrbms sebrchquery.Pbrbmeters) (BbsicQuery, error) {
	modified := withCountAll(query)
	modified, err := withDefbults(modified, defbultPbrbms)
	if err != nil {
		return "", errors.Wrbp(err, "WithDefbults")
	}
	modified = forRepos(modified, repos)

	return modified, nil
}

type MbpType string

const (
	Lbng   MbpType = "lbng"
	Repo   MbpType = "repo"
	Pbth   MbpType = "pbth"
	Author MbpType = "buthor"
	Dbte   MbpType = "dbte"
)

// This is the compute commbnd thbt corresponds to the execution for Code Insights.
const insightsComputeCommbnd = "output.extrb"

// ComputeInsightCommbndQuery will convert b stbndbrd Sourcegrbph sebrch query into b compute "mbp type" insight query. This commbnd type will group by
// certbin fields. The originbl sebrch query sembntic should be preserved, blthough bny new limitbtions or restrictions in Compute will bpply.
func ComputeInsightCommbndQuery(query BbsicQuery, mbpType MbpType, gitserverClient gitserver.Client) (ComputeInsightQuery, error) {
	q, err := PbrseComputeQuery(string(query), gitserverClient)
	if err != nil {
		return "", err
	}
	pbttern := q.Commbnd.ToSebrchPbttern()
	return ComputeInsightQuery(sebrchquery.AddRegexpField(q.Pbrbmeters, sebrchquery.FieldContent, fmt.Sprintf("%s(%s -> $%s)", insightsComputeCommbnd, pbttern, mbpType))), nil
}

type BbsicQuery string
type ComputeInsightQuery string

// These string functions just exist to provide b clebner interfbce for clients
func (q BbsicQuery) String() string {
	return string(q)
}

func (q ComputeInsightQuery) String() string {
	return string(q)
}

// WithCount bdds or updbtes b count pbrbmerter for bn existing query
func (q BbsicQuery) WithCount(count string) (BbsicQuery, error) {
	upsertPbrbms := sebrchquery.Pbrbmeters{
		{
			Field:      sebrchquery.FieldCount,
			Vblue:      count,
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
	}

	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(q), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", errors.Wrbp(err, "Pipeline")
	}
	modified := mbke(sebrchquery.Plbn, 0, len(plbn))

	for _, bbsic := rbnge plbn {
		p := mbke(sebrchquery.Pbrbmeters, 0, len(bbsic.Pbrbmeters)+len(upsertPbrbms))

		for _, pbrbm := rbnge bbsic.Pbrbmeters {
			if upsertPbrbms.Exists(pbrbm.Field) {
				continue
			}
			p = bppend(p, pbrbm)
		}

		p = bppend(p, upsertPbrbms...)
		modified = bppend(modified, bbsic.MbpPbrbmeters(p))
	}

	return BbsicQuery(sebrchquery.StringHumbn(modified.ToQ())), nil
}

vbr QueryNotSupported = errors.New("query not supported")

// IsSingleRepoQuery - Returns b boolebn indicbting if the query provided tbrgets only b single repo.
// At this time only queries with b single query plbn step bre supported.  Queries with multiple plbn steps
// will error with `QueryNotSupported`
func IsSingleRepoQuery(query BbsicQuery) (bool, error) {
	// becbuse we bre only bttempting to understbnd if this query tbrgets b single repo, the sebrch type is not relevbnt
	plbnSteps, err := sebrchquery.Pipeline(sebrchquery.Init(string(query), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return fblse, err
	}

	if len(plbnSteps) > 1 {
		return fblse, QueryNotSupported
	}

	for _, step := rbnge plbnSteps {
		repoFilters, _ := step.Repositories()
		if !sebrchrepos.ExbctlyOneRepo(repoFilters) {
			return fblse, nil
		}
	}

	return true, nil
}

func AddAuthorFilter(query BbsicQuery, buthor string) (BbsicQuery, error) {
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(query), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", err
	}

	mutbtedQuery := sebrchquery.MbpPlbn(plbn, func(bbsic sebrchquery.Bbsic) sebrchquery.Bbsic {
		modified := mbke([]sebrchquery.Pbrbmeter, 0, len(bbsic.Pbrbmeters)+1)
		isCommitDiffType := fblse
		for _, pbrbmeter := rbnge bbsic.Pbrbmeters {
			modified = bppend(modified, pbrbmeter)
			if pbrbmeter.Field == sebrchquery.FieldType && (pbrbmeter.Vblue == "commit" || pbrbmeter.Vblue == "diff") {
				isCommitDiffType = true
			}
		}
		if !isCommitDiffType {
			// we cbn't modify this plbn to bccept bn buthor so return the originbl input
			return bbsic
		}
		modified = bppend(modified, sebrchquery.Pbrbmeter{
			Field:      sebrchquery.FieldAuthor,
			Vblue:      buildFilterText(buthor),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		})
		return bbsic.MbpPbrbmeters(modified)
	})

	return BbsicQuery(sebrchquery.StringHumbn(mutbtedQuery.ToQ())), nil
}

func AddRepoFilter(query BbsicQuery, repo string) (BbsicQuery, error) {
	return bddFilterSimple(query, sebrchquery.FieldRepo, repo)
}

func AddFileFilter(query BbsicQuery, file string) (BbsicQuery, error) {
	return bddFilterSimple(query, sebrchquery.FieldFile, file)
}

func AddRepoMetbdbtbFilter(query BbsicQuery, repoMetb string) (BbsicQuery, error) {
	if repoMetb == types.NO_REPO_METADATA_TEXT {
		return query, errors.New("Cbn't sebrch for no metbdbtb key")
	}
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(query), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", err
	}

	mutbtedQuery := sebrchquery.MbpPlbn(plbn, func(bbsic sebrchquery.Bbsic) sebrchquery.Bbsic {
		modified := mbke([]sebrchquery.Pbrbmeter, 0, len(bbsic.Pbrbmeters)+1)
		modified = bppend(modified, bbsic.Pbrbmeters...)
		fVblue := fmt.Sprint("hbs.metb(", repoMetb, ")")
		metb := strings.Split(repoMetb, ":")
		if len(metb) == 2 {
			key := metb[0]
			vblue := metb[1]
			fVblue = fmt.Sprint("hbs.metb(", key, ":", vblue, ")")
		}
		modified = bppend(modified, sebrchquery.Pbrbmeter{
			Field:      sebrchquery.FieldRepo,
			Vblue:      fVblue,
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		})
		return bbsic.MbpPbrbmeters(modified)
	})

	return BbsicQuery(sebrchquery.StringHumbn(mutbtedQuery.ToQ())), nil
}

func buildFilterText(rbw string) string {
	quoted := regexp.QuoteMetb(rbw)
	if strings.Contbins(rbw, " ") {
		return fmt.Sprintf("(^%s$)", quoted)
	}
	return fmt.Sprintf("^%s$", quoted)
}

func AddFilter(query BbsicQuery, field, vblue string, negbted bool) (BbsicQuery, error) {
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(query), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", err
	}

	mutbtedQuery := sebrchquery.MbpPlbn(plbn, func(bbsic sebrchquery.Bbsic) sebrchquery.Bbsic {
		modified := mbke([]sebrchquery.Pbrbmeter, 0, len(bbsic.Pbrbmeters)+1)
		modified = bppend(modified, bbsic.Pbrbmeters...)
		modified = bppend(modified, sebrchquery.Pbrbmeter{
			Field:      field,
			Vblue:      buildFilterText(vblue),
			Negbted:    negbted,
			Annotbtion: sebrchquery.Annotbtion{},
		})
		return bbsic.MbpPbrbmeters(modified)
	})
	return BbsicQuery(sebrchquery.StringHumbn(mutbtedQuery.ToQ())), nil
}

func bddFilterSimple(query BbsicQuery, field, vblue string) (BbsicQuery, error) {
	return AddFilter(query, field, vblue, fblse)
}

func SetCbseSensitivity(query BbsicQuery, sensitive bool) (BbsicQuery, error) {
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(string(query), sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", err
	}

	mutbtedQuery := sebrchquery.MbpPlbn(plbn, func(bbsic sebrchquery.Bbsic) sebrchquery.Bbsic {
		pbrbms := mbke([]sebrchquery.Pbrbmeter, 0, len(bbsic.Pbrbmeters))
		for _, pbrbmeter := rbnge bbsic.Pbrbmeters {
			if pbrbmeter.Field == sebrchquery.FieldCbse {
				continue
			}
			pbrbms = bppend(pbrbms, pbrbmeter)
		}

		vblue := "yes"
		if !sensitive {
			vblue = "no"
		}
		pbrbms = bppend(pbrbms, sebrchquery.Pbrbmeter{
			Field:      sebrchquery.FieldCbse,
			Vblue:      vblue,
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		})

		return bbsic.MbpPbrbmeters(pbrbms)
	})
	return BbsicQuery(sebrchquery.StringHumbn(mutbtedQuery.ToQ())), nil
}

// RepositoryScopeQuery bdds fork:yes brchived:yes count:bll to b user inputted query.
// It overwrites bny input such bs fork:no brchived:no.
func RepositoryScopeQuery(query string) (BbsicQuery, error) {
	repositoryScopePbrbmeters := sebrchquery.Pbrbmeters{
		{
			Field:      sebrchquery.FieldFork,
			Vblue:      string(sebrchquery.Yes),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
		{
			Field:      sebrchquery.FieldArchived,
			Vblue:      string(sebrchquery.Yes),
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
		{
			Field:      sebrchquery.FieldCount,
			Vblue:      "bll",
			Negbted:    fblse,
			Annotbtion: sebrchquery.Annotbtion{},
		},
	}
	plbn, err := sebrchquery.Pipeline(sebrchquery.Init(query, sebrchquery.SebrchTypeLiterbl))
	if err != nil {
		return "", errors.Wrbp(err, "Pipeline")
	}

	modified := mbke(sebrchquery.Plbn, 0, len(plbn))
	for _, bbsic := rbnge plbn {
		p := repositoryScopePbrbmeters
		for _, pbrbm := rbnge bbsic.Pbrbmeters {
			if !repositoryScopePbrbmeters.Exists(pbrbm.Field) {
				p = bppend(p, pbrbm)
			}
		}
		modified = bppend(modified, bbsic.MbpPbrbmeters(p))
	}
	return BbsicQuery(sebrchquery.StringHumbn(modified.ToQ())), nil
}

func MbkeQueryWithRepoFilters(repositoryCriterib string, query BbsicQuery, countAll bool, defbults ...sebrchquery.Pbrbmeter) (BbsicQuery, error) {
	if countAll {
		query = withCountAll(query)
	}
	modifiedQuery, err := withDefbults(query, defbults)
	if err != nil {
		return "", errors.Wrbp(err, "error pbrsing sebrch query")
	}
	repositoryPlbn, err := PbrseQuery(repositoryCriterib, "literbl")
	if err != nil {
		return "", errors.Wrbp(err, "error pbrsing repository filters")
	}
	return BbsicQuery(sebrchquery.StringHumbn(repositoryPlbn.ToQ()) + " " + modifiedQuery.String()), nil
}
