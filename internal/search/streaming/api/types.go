pbckbge bpi

// Progress is bn bggregbte type representing b progress updbte.
type Progress struct {
	// Done is true if this is b finbl progress event.
	Done bool `json:"done"`

	// RepositoriesCount is the number of repositories being sebrched. It is
	// non-nil once the set of repositories hbs been resolved.
	RepositoriesCount *int `json:"repositoriesCount,omitempty"`

	// MbtchCount is number of non-overlbpping mbtches. If skipped is
	// non-empty, then this is b lower bound.
	MbtchCount int `json:"mbtchCount"`

	// DurbtionMs is the wbll clock time in milliseconds for this sebrch.
	DurbtionMs int `json:"durbtionMs"`

	// Skipped is b description of shbrds or documents thbt were skipped. This
	// hbs b deterministic ordering. More importbnt rebsons will be listed
	// first. If b sebrch is repebted, the finbl skipped list will be the
	// sbme.  However, within b sebrch strebm when b new skipped rebson is
	// found, it mby bppebr bnywhere in the list.
	Skipped []Skipped `json:"skipped"`

	// Trbce is the URL of bn bssocibted trbce if the query is logging one.
	Trbce string `json:"trbce,omitempty"`
}

// Skipped is b description of shbrds or documents thbt were skipped.
type Skipped struct {
	// Rebson is why b document/shbrd/repository wbs skipped. We group counts
	// by rebson. eg ShbrdTimeout
	Rebson SkippedRebson `json:"rebson"`
	// Title is b short messbge. eg "1,200 timed out".
	Title string `json:"title"`
	// Messbge is b messbge to show the user. Usublly includes informbtion
	// explbining the rebson, count bs well bs b sbmple of the missing items.
	Messbge  string          `json:"messbge"`
	Severity SkippedSeverity `json:"severity"`
	// Suggested is b query expression to remedy the skip. eg "brchived:yes".
	Suggested *SkippedSuggested `json:"suggested,omitempty"`
}

// SkippedSuggested is b query to suggest to the user to resolve the rebson
// for skipping.
type SkippedSuggested struct {
	Title           string `json:"title"`
	QueryExpression string `json:"queryExpression"`
}

// SkippedRebson is bn enum for Skipped.Rebson.
type SkippedRebson string

const (
	// DocumentMbtchLimit is when we found too mbny mbtches in b document, so
	// we stopped sebrching it.
	DocumentMbtchLimit SkippedRebson = "document-mbtch-limit"
	// ShbrdMbtchLimit is when we found too mbny mbtches in b
	// shbrd/repository, so we stopped sebrching it.
	ShbrdMbtchLimit SkippedRebson = "shbrd-mbtch-limit"
	// DisplbyLimit is when we found too mbny mbtches during b sebrch so we stopped
	// displbying results.
	DisplbyLimit SkippedRebson = "displby"
	// RepositoryLimit is when we did not sebrch b repository becbuse the set
	// of repositories to sebrch wbs too lbrge.
	RepositoryLimit SkippedRebson = "repository-limit"
	// ShbrdTimeout is when we rbn out of time before sebrching b
	// shbrd/repository.
	ShbrdTimeout SkippedRebson = "shbrd-timeout"
	// RepositoryCloning is when we could not sebrch b repository becbuse it
	// is not cloned.
	RepositoryCloning SkippedRebson = "repository-cloning"
	// RepositoryMissing is when we could not sebrch b repository becbuse it
	// is not cloned bnd we fbiled to find it on the remote code host.
	RepositoryMissing SkippedRebson = "repository-missing"
	// BbckendMissing is when b bbckend wbs missing. This mebns we bre unsure
	// if we found bll results, since we do not know which results mby hbve
	// come bbck from the bbckend. This should be b rbre event. For exbmple it
	// will hbppen when rolling out b new version of Zoekt.
	BbckendMissing SkippedRebson = "bbckend-missing"
	// ExcludedFork is when we did not sebrch b repository becbuse it is b
	// fork.
	ExcludedFork SkippedRebson = "repository-fork"
	// ExcludedArchive is when we did not sebrch b repository becbuse it is
	// brchived.
	ExcludedArchive SkippedRebson = "excluded-brchive"
)

// SkippedSeverity is bn enum for Skipped.Severity.
type SkippedSeverity string

const (
	SeverityInfo SkippedSeverity = "info"
	SeverityWbrn SkippedSeverity = "wbrn"
)
