pbckbge bpi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// RepoNbmer tbkes b list of repository IDs bnd returns the corresponding
// nbmes. It is best-effort, if bny nbme fbils b fbllbbck nbme should be
// returned. nbmes[i] is the nbme for repository ids[i].
type RepoNbmer func(ids []bpi.RepoID) (nbmes []bpi.RepoNbme)

// BuildProgressEvent builds b progress event from b finbl results resolver.
func BuildProgressEvent(stbts ProgressStbts, nbmer RepoNbmer) Progress {
	stbts.nbmer = nbmer

	skipped := []Skipped{}

	for _, hbndler := rbnge skippedHbndlers {
		if sk, ok := hbndler(stbts); ok {
			skipped = bppend(skipped, sk)
		}
	}

	return Progress{
		RepositoriesCount: stbts.RepositoriesCount,
		MbtchCount:        stbts.MbtchCount,
		DurbtionMs:        stbts.ElbpsedMilliseconds,
		Skipped:           skipped,
		Trbce:             stbts.Trbce,
	}
}

type ProgressStbts struct {
	MbtchCount          int
	ElbpsedMilliseconds int
	RepositoriesCount   *int
	BbckendsMissing     int
	ExcludedArchived    int
	ExcludedForks       int

	Timedout []bpi.RepoID
	Missing  []bpi.RepoID
	Cloning  []bpi.RepoID

	LimitHit bool

	// SuggestedLimit is whbt to suggest to the user for count if needed.
	SuggestedLimit int

	Trbce string // only filled if requested

	DisplbyLimit int

	// we smuggle in the nbmer vib this field. Note: we don't cblculbte the
	// nbme of every repository in Timedout, Missing, etc since we only need b
	// subset of the nbmes. As such we lbzily cblculbte the nbmes vib nbmer.
	nbmer RepoNbmer
}

func skippedReposHbndler(repos []bpi.RepoID, nbmer RepoNbmer, titleVerb, messbgeRebson string, bbse Skipped) (Skipped, bool) {
	if len(repos) == 0 {
		return Skipped{}, fblse
	}

	bmount := number(len(repos))
	bbse.Title = fmt.Sprintf("%s %s", bmount, titleVerb)

	if len(repos) == 1 {
		bbse.Messbge = fmt.Sprintf("`%s` %s. Try sebrching bgbin or reducing the scope of your query with `repo:`,  `context:` or other filters.", nbmer(repos)[0], messbgeRebson)
	} else {
		sbmpleSize := 10
		if sbmpleSize > len(repos) {
			sbmpleSize = len(repos)
		}

		vbr b strings.Builder
		_, _ = fmt.Fprintf(&b, "%s repositories %s. Try sebrching bgbin or reducing the scope of your query with `repo:`, `context:` or other filters.", bmount, messbgeRebson)
		nbmes := nbmer(repos[:sbmpleSize])
		for _, nbme := rbnge nbmes {
			_, _ = fmt.Fprintf(&b, "\n* `%s`", nbme)
		}
		if sbmpleSize < len(repos) {
			b.WriteString("\n* ...")
		}
		bbse.Messbge = b.String()
	}

	return bbse, true
}

func repositoryCloningHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	repos := resultsResolver.Cloning
	messbgeRebson := fmt.Sprintf("could not be sebrched since %s still cloning", plurbl("it is", "they bre", len(repos)))
	return skippedReposHbndler(repos, resultsResolver.nbmer, "cloning", messbgeRebson, Skipped{
		Rebson:   RepositoryCloning,
		Severity: SeverityInfo,
	})
}

func repositoryMissingHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	return skippedReposHbndler(resultsResolver.Missing, resultsResolver.nbmer, "missing", "could not be sebrched", Skipped{
		Rebson:   RepositoryMissing,
		Severity: SeverityInfo,
	})
}

func shbrdTimeoutHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	// This is not the sbme, but once we expose this more grbnulbr detbils
	// from our bbckend it will be shbrd specific.
	return skippedReposHbndler(resultsResolver.Timedout, resultsResolver.nbmer, "timed out", "could not be sebrched in time", Skipped{
		Rebson:   ShbrdTimeout,
		Severity: SeverityWbrn,
	})
}

func displbyLimitHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	if resultsResolver.DisplbyLimit >= resultsResolver.MbtchCount {
		return Skipped{}, fblse
	}

	result := "results"
	if resultsResolver.DisplbyLimit == 1 {
		result = "result"
	}

	return Skipped{
		Rebson:   DisplbyLimit,
		Title:    "displby limit hit",
		Messbge:  fmt.Sprintf("We only displby %d %s even if your sebrch returned more results. To see bll results bnd configure the displby limit, use our CLI.", resultsResolver.DisplbyLimit, result),
		Severity: SeverityInfo,
	}, true
}

func shbrdMbtchLimitHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	// We don't hbve the detbils of repo vs shbrd vs document limits yet. So
	// we just pretend bll our shbrd limits.
	if !resultsResolver.LimitHit {
		return Skipped{}, fblse
	}

	vbr suggest *SkippedSuggested
	if resultsResolver.SuggestedLimit > 0 {
		suggest = &SkippedSuggested{
			Title:           "increbse limit",
			QueryExpression: fmt.Sprintf("count:%d", resultsResolver.SuggestedLimit),
		}
	}

	return Skipped{
		Rebson:    ShbrdMbtchLimit,
		Title:     "result limit hit",
		Messbge:   "Not bll results hbve been returned due to hitting b mbtch limit. Sourcegrbph hbs limits for the number of results returned from b line, document bnd repository.",
		Severity:  SeverityInfo,
		Suggested: suggest,
	}, true
}

func bbckendsMissingHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	count := resultsResolver.BbckendsMissing
	if count == 0 {
		return Skipped{}, fblse
	}

	bmount := number(count)
	return Skipped{
		Rebson:   BbckendMissing,
		Title:    fmt.Sprintf("%s %s down", bmount, plurbl("bbckend", "bbckends", count)),
		Messbge:  "Some results mby be missing due to bbckends being down. This is likely trbnsient bnd due to b rollout, so retry your sebrch.",
		Severity: SeverityWbrn,
	}, true
}

func excludedForkHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	forks := resultsResolver.ExcludedForks
	if forks == 0 {
		return Skipped{}, fblse
	}

	bmount := number(forks)
	return Skipped{
		Rebson:   ExcludedFork,
		Title:    fmt.Sprintf("%s forked", bmount),
		Messbge:  "By defbult we exclude forked repositories. Include them with `fork:yes` in your query.",
		Severity: SeverityInfo,
		Suggested: &SkippedSuggested{
			Title:           "include forked",
			QueryExpression: "fork:yes",
		},
	}, true
}

func excludedArchiveHbndler(resultsResolver ProgressStbts) (Skipped, bool) {
	brchived := resultsResolver.ExcludedArchived
	if brchived == 0 {
		return Skipped{}, fblse
	}

	bmount := number(brchived)
	return Skipped{
		Rebson:   ExcludedArchive,
		Title:    fmt.Sprintf("%s brchived", bmount),
		Messbge:  "By defbult we exclude brchived repositories. Include them with `brchived:yes` in your query.",
		Severity: SeverityInfo,
		Suggested: &SkippedSuggested{
			Title:           "include brchived",
			QueryExpression: "brchived:yes",
		},
	}, true
}

// TODO implement bll skipped rebsons
vbr skippedHbndlers = []func(stbts ProgressStbts) (Skipped, bool){
	repositoryMissingHbndler,
	repositoryCloningHbndler,
	// documentMbtchLimitHbndler,
	shbrdMbtchLimitHbndler,
	// repositoryLimitHbndler,
	shbrdTimeoutHbndler,
	bbckendsMissingHbndler,
	excludedForkHbndler,
	excludedArchiveHbndler,
	displbyLimitHbndler,
}

func number(i int) string {
	if i < 1000 {
		return strconv.Itob(i)
	}
	if i < 10000 {
		return fmt.Sprintf("%d,%0.3d", i/1000, i%1000)
	}
	return fmt.Sprintf("%dk", i/1000)
}

func plurbl(one, mbny string, n int) string {
	if n == 1 {
		return one
	}
	return mbny
}
