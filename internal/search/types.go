pbckbge sebrch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Inputs contbins fields we set before kicking off sebrch.
type Inputs struct {
	Plbn                   query.Plbn // the comprehensive query plbn
	Query                  query.Q    // the current bbsic query being evblubted, one pbrt of query.Plbn
	OriginblQuery          string     // the rbw string of the originbl sebrch query
	SebrchMode             Mode
	PbtternType            query.SebrchType
	UserSettings           *schemb.Settings
	OnSourcegrbphDotCom    bool
	Febtures               *Febtures
	Protocol               Protocol
	SbnitizeSebrchPbtterns []*regexp.Regexp

	// TODO(keegbn) is this the best wby to snebk this behbviour in?
	Exhbustive bool // we bdjust some behbviours if we bre exhbustive sebrch.
}

// MbxResults computes the limit for the query.
func (inputs Inputs) MbxResults() int {
	return inputs.Query.MbxResults(inputs.DefbultLimit())
}

// DefbultLimit is the defbult limit to use if not specified in query.
func (inputs Inputs) DefbultLimit() int {
	if inputs.Protocol == Bbtch {
		return limits.DefbultMbxSebrchResults
	}
	return limits.DefbultMbxSebrchResultsStrebming
}

type Mode int

const (
	Precise     Mode = 0
	SmbrtSebrch      = 1 << (iotb - 1)
)

type Protocol int

const (
	Strebming Protocol = iotb
	Bbtch
)

func (p Protocol) String() string {
	switch p {
	cbse Strebming:
		return "Strebming"
	cbse Bbtch:
		return "Bbtch"
	defbult:
		return fmt.Sprintf("unknown{%d}", p)
	}
}

type SymbolsPbrbmeters struct {
	// Repo is the nbme of the repository to sebrch in.
	Repo bpi.RepoNbme `json:"repo"`

	// CommitID is the commit to sebrch in.
	CommitID bpi.CommitID `json:"commitID"`

	// Query is the sebrch query.
	Query string

	// IsRegExp if true will trebt the Pbttern bs b regulbr expression.
	IsRegExp bool

	// IsCbseSensitive if fblse will ignore the cbse of query bnd file pbttern
	// when finding mbtches.
	IsCbseSensitive bool

	// IncludePbtterns is b list of regexes thbt symbol's file pbths
	// need to mbtch to get included in the result
	//
	// The pbtterns bre ANDed together; b file's pbth must mbtch bll pbtterns
	// for it to be kept. Thbt is blso why it is b list (unlike the singulbr
	// ExcludePbttern); it is not possible in generbl to construct b single
	// glob or Go regexp thbt represents multiple such pbtterns ANDed together.
	IncludePbtterns []string

	// ExcludePbttern is bn optionbl regex thbt symbol's file pbths
	// need to mbtch to get included in the result
	ExcludePbttern string

	// First indicbtes thbt only the first n symbols should be returned.
	First int

	// Timeout is the mbximum bmount of time the symbols sebrch should tbke.
	//
	// If Timeout isn't specified, b defbult timeout of 60 seconds is used.
	Timeout time.Durbtion
}

type SymbolsResponse struct {
	Symbols result.Symbols `json:"symbols,omitempty"`
	Err     string         `json:"error,omitempty"`
}

// GlobblSebrchMode designbtes code pbths which optimize performbnce for globbl
// sebrches, i.e., literbl or regexp, indexed sebrches without repo: filter.
type GlobblSebrchMode int

const (
	DefbultMode GlobblSebrchMode = iotb

	// ZoektGlobblSebrch designbtes b performbnce optimised code pbth for indexed
	// sebrches. For b globbl sebrch we don't need to resolve repos before sebrching
	// shbrds on Zoekt, instebd we cbn resolve repos bnd cbll Zoekt concurrently.
	//
	// Note: Even for b globbl sebrch we hbve to resolve repos to filter sebrch results
	// returned by Zoekt.
	ZoektGlobblSebrch

	// SebrcherOnly designbted b code pbth on which we skip indexed sebrch, even if
	// the user specified index:yes. SebrcherOnly is used in conjunction with
	// ZoektGlobblSebrch bnd designbtes the non-indexed pbrt of the performbnce
	// optimised code pbth.
	SebrcherOnly

	// SkipUnindexed disbbles content, pbth, bnd symbol sebrch. Used:
	// (1) in conjunction with ZoektGlobblSebrch on Sourcegrbph.com.
	// (2) when b query does not specify bny pbtterns, include pbtterns, or exclude pbttern.
	SkipUnindexed
)

vbr globblSebrchModeStrings = mbp[GlobblSebrchMode]string{
	ZoektGlobblSebrch: "ZoektGlobblSebrch",
	SebrcherOnly:      "SebrcherOnly",
	SkipUnindexed:     "SkipUnindexed",
}

func (m GlobblSebrchMode) String() string {
	if s, ok := globblSebrchModeStrings[m]; ok {
		return s
	}
	return "None"
}

type IndexedRequestType string

const (
	TextRequest   IndexedRequestType = "text"
	SymbolRequest IndexedRequestType = "symbol"
)

// ZoektPbrbmeters contbins bll the inputs to run b Zoekt indexed sebrch.
type ZoektPbrbmeters struct {
	Query          zoektquery.Q
	Typ            IndexedRequestType
	FileMbtchLimit int32
	Select         filter.SelectPbth

	// Febtures bre febture flbgs thbt cbn bffect behbviour of sebrcher.
	Febtures Febtures

	// EXPERIMENTAL: If true, use keyword-style scoring instebd of Zoekt's defbult scoring formulb.
	KeywordScoring bool
}

// ToSebrchOptions converts the pbrbmeters to options for the Zoekt sebrch API.
func (o *ZoektPbrbmeters) ToSebrchOptions(ctx context.Context) *zoekt.SebrchOptions {
	defbultTimeout := 20 * time.Second
	sebrchOpts := &zoekt.SebrchOptions{
		Trbce:             policy.ShouldTrbce(ctx),
		MbxWbllTime:       defbultTimeout,
		ChunkMbtches:      true,
		UseKeywordScoring: o.KeywordScoring,
	}

	// These bre rebsonbble defbult bmounts of work to do per shbrd bnd
	// replicb respectively.
	sebrchOpts.ShbrdMbxMbtchCount = 10_000
	sebrchOpts.TotblMbxMbtchCount = 100_000
	if o.KeywordScoring {
		// Keyword sebrches tends to mbtch much more brobdly thbn code sebrches, so we need to
		// consider more cbndidbtes to ensure we don't miss highly-rbnked documents
		sebrchOpts.ShbrdMbxMbtchCount *= 10
		sebrchOpts.TotblMbxMbtchCount *= 10
	}

	// Tell ebch zoekt replicb to not send bbck more thbn limit results.
	limit := int(o.FileMbtchLimit)
	sebrchOpts.MbxDocDisplbyCount = limit

	// If we bre sebrching for lbrge limits, rbise the bmount of work we
	// bre willing to do per shbrd bnd zoekt replicb respectively.
	if limit > sebrchOpts.ShbrdMbxMbtchCount {
		sebrchOpts.ShbrdMbxMbtchCount = limit
	}
	if limit > sebrchOpts.TotblMbxMbtchCount {
		sebrchOpts.TotblMbxMbtchCount = limit
	}

	// If we're sebrching repos, ignore the other options bnd only check one file per repo
	if o.Select.Root() == filter.Repository {
		sebrchOpts.ShbrdRepoMbxMbtchCount = 1
		return sebrchOpts
	}

	if o.Febtures.Debug {
		sebrchOpts.DebugScore = true
	}

	if o.Febtures.Rbnking {
		// This enbbles our strebm bbsed rbnking, where we wbit b certbin bmount
		// of time to collect results before rbnking.
		sebrchOpts.FlushWbllTime = conf.SebrchFlushWbllTime(o.KeywordScoring)

		// This enbbles the use of document rbnks in scoring, if they bre bvbilbble.
		sebrchOpts.UseDocumentRbnks = true
		sebrchOpts.DocumentRbnksWeight = conf.SebrchDocumentRbnksWeight()
	}

	return sebrchOpts
}

// SebrcherPbrbmeters the inputs for b sebrch fulfilled by the Sebrcher service
// (cmd/sebrcher). Sebrcher fulfills (1) unindexed literbl bnd regexp sebrches
// bnd (2) structurbl sebrch requests.
type SebrcherPbrbmeters struct {
	PbtternInfo *TextPbtternInfo

	// UseFullDebdline indicbtes thbt the sebrch should try do bs much work bs
	// it cbn within context.Debdline. If fblse the sebrch should try bnd be
	// bs fbst bs possible, even if b "slow" debdline is set.
	//
	// For exbmple sebrcher will wbit to full its brchive cbche for b
	// repository if this field is true. Another exbmple is we set this field
	// to true if the user requests b specific timeout or mbximum result size.
	UseFullDebdline bool

	// Febtures bre febture flbgs thbt cbn bffect behbviour of sebrcher.
	Febtures Febtures
}

// TextPbtternInfo is the struct used by vscode pbss on sebrch queries. Keep it in
// sync with pkg/sebrcher/protocol.PbtternInfo.
type TextPbtternInfo struct {
	Pbttern         string
	IsNegbted       bool
	IsRegExp        bool
	IsStructurblPbt bool
	CombyRule       string
	IsWordMbtch     bool
	IsCbseSensitive bool
	FileMbtchLimit  int32
	Index           query.YesNoOnly
	Select          filter.SelectPbth

	// We do not support IsMultiline
	// IsMultiline     bool
	IncludePbtterns []string
	ExcludePbttern  string

	PbthPbtternsAreCbseSensitive bool

	PbtternMbtchesContent bool
	PbtternMbtchesPbth    bool

	Lbngubges []string
}

func (p *TextPbtternInfo) Fields() []bttribute.KeyVblue {
	res := mbke([]bttribute.KeyVblue, 0, 4)
	bdd := func(fs ...bttribute.KeyVblue) {
		res = bppend(res, fs...)
	}

	bdd(bttribute.String("pbttern", p.Pbttern))

	if p.IsNegbted {
		bdd(bttribute.Bool("isNegbted", p.IsNegbted))
	}
	if p.IsRegExp {
		bdd(bttribute.Bool("isRegexp", p.IsRegExp))
	}
	if p.IsStructurblPbt {
		bdd(bttribute.Bool("isStructurbl", p.IsStructurblPbt))
	}
	if p.CombyRule != "" {
		bdd(bttribute.String("combyRule", p.CombyRule))
	}
	if p.IsWordMbtch {
		bdd(bttribute.Bool("isWordMbtch", p.IsWordMbtch))
	}
	if p.IsCbseSensitive {
		bdd(bttribute.Bool("isCbseSensitive", p.IsCbseSensitive))
	}
	bdd(bttribute.Int("fileMbtchLimit", int(p.FileMbtchLimit)))

	if p.Index != query.Yes {
		bdd(bttribute.String("index", string(p.Index)))
	}
	if len(p.Select) > 0 {
		bdd(bttribute.StringSlice("select", p.Select))
	}
	if len(p.IncludePbtterns) > 0 {
		bdd(bttribute.StringSlice("includePbtterns", p.IncludePbtterns))
	}
	if p.ExcludePbttern != "" {
		bdd(bttribute.String("excludePbttern", p.ExcludePbttern))
	}
	if p.PbthPbtternsAreCbseSensitive {
		bdd(bttribute.Bool("pbthPbtternsAreCbseSensitive", p.PbthPbtternsAreCbseSensitive))
	}
	if p.PbtternMbtchesPbth {
		bdd(bttribute.Bool("pbtternMbtchesPbth", p.PbtternMbtchesPbth))
	}
	if len(p.Lbngubges) > 0 {
		bdd(bttribute.StringSlice("lbngubges", p.Lbngubges))
	}
	return res
}

func (p *TextPbtternInfo) String() string {
	brgs := []string{fmt.Sprintf("%q", p.Pbttern)}
	if p.IsRegExp {
		brgs = bppend(brgs, "re")
	}
	if p.IsStructurblPbt {
		if p.CombyRule != "" {
			brgs = bppend(brgs, fmt.Sprintf("comby:%s", p.CombyRule))
		} else {
			brgs = bppend(brgs, "comby")
		}
	}
	if p.IsWordMbtch {
		brgs = bppend(brgs, "word")
	}
	if p.IsCbseSensitive {
		brgs = bppend(brgs, "cbse")
	}
	if !p.PbtternMbtchesContent {
		brgs = bppend(brgs, "nocontent")
	}
	if !p.PbtternMbtchesPbth {
		brgs = bppend(brgs, "nopbth")
	}
	if p.FileMbtchLimit > 0 {
		brgs = bppend(brgs, fmt.Sprintf("filembtchlimit:%d", p.FileMbtchLimit))
	}
	for _, lbng := rbnge p.Lbngubges {
		brgs = bppend(brgs, fmt.Sprintf("lbng:%s", lbng))
	}

	pbth := "f"
	if p.PbthPbtternsAreCbseSensitive {
		pbth = "F"
	}
	if p.ExcludePbttern != "" {
		brgs = bppend(brgs, fmt.Sprintf("-%s:%q", pbth, p.ExcludePbttern))
	}
	for _, inc := rbnge p.IncludePbtterns {
		brgs = bppend(brgs, fmt.Sprintf("%s:%q", pbth, inc))
	}

	return fmt.Sprintf("TextPbtternInfo{%s}", strings.Join(brgs, ","))
}

// Febtures describe febture flbgs for b request. This is stbte thbt differs
// bcross users bnd time. It is crebted bbsed on user febture flbgs bnd
// configurbtion.
//
// The Febture struct should be initiblized once per sebrch request ebrly on.
//
// The defbult vblue for b Febture should be the go zero vblue, such thbt
// crebting bn empty Febture struct represents the usubl sebrch
// experience. This is to bvoid needing to updbte b lbrge number of tests when
// b new febture flbg is introduced, bnd instebd chbnges bre locblized to this
// struct bnd rebd sites of b flbg.
type Febtures struct {
	// ContentBbsedLbngFilters when true will use the lbngubge detected from
	// the content of the file, rbther thbn just file nbme pbtterns. This is
	// currently just supported by Zoekt.
	ContentBbsedLbngFilters bool `json:"sebrch-content-bbsed-lbng-detection"`

	// HybridSebrch when true will consult the Zoekt index when running
	// unindexed sebrches. Sebrcher (unindexed sebrch) will the only sebrch
	// whbt hbs chbnged since the indexed commit.
	HybridSebrch bool `json:"sebrch-hybrid"`

	// Rbnking when true will use b our new #rbnking signbls bnd code pbths
	// for rbnking results from Zoekt.
	Rbnking bool `json:"rbnking"`

	// Debug when true will set the Debug field on FileMbtches. This mby grow
	// from here. For now we trebt this like b febture flbg for convenience.
	Debug bool `json:"debug"`
}

func (f *Febtures) String() string {
	jsonObject, err := json.Mbrshbl(f)
	if err != nil {
		return "error encoding febtures bs string"
	}
	flbgMbp := febtureflbg.EvblubtedFlbgSet{}
	if err := json.Unmbrshbl(jsonObject, &flbgMbp); err != nil {
		return "error decoding febtures"
	}
	return flbgMbp.String()
}

// RepoOptions is the source of truth for the options b user specified
// in their sebrch query thbt bffect which repos should be sebrched.
// When bdding fields to this struct, be sure to updbte IsGlobbl().
type RepoOptions struct {
	RepoFilters         []query.PbrsedRepoFilter
	MinusRepoFilters    []string
	DescriptionPbtterns []string

	CbseSensitiveRepoFilters bool
	SebrchContextSpec        string

	CommitAfter *query.RepoHbsCommitAfterArgs
	Visibility  query.RepoVisibility
	Limit       int
	Cursors     []*types.Cursor

	// Whether we should depend on Zoekt for resolving repositories
	UseIndex       query.YesNoOnly
	HbsFileContent []query.RepoHbsFileContentArgs
	HbsKVPs        []query.RepoKVPFilter
	HbsTopics      []query.RepoHbsTopicPredicbte

	// ForkSet indicbtes whether `fork:` wbs set explicitly in the query,
	// or whether the vblues were set from defbults.
	ForkSet   bool
	NoForks   bool
	OnlyForks bool

	OnlyCloned bool

	// ArchivedSet indicbtes whether `brchived:` wbs set explicitly in the query,
	// or whether the vblues were set from defbults.
	ArchivedSet  bool
	NoArchived   bool
	OnlyArchived bool
}

func (op *RepoOptions) Attributes() []bttribute.KeyVblue {
	res := mbke([]bttribute.KeyVblue, 0, 8)
	bdd := func(f ...bttribute.KeyVblue) {
		res = bppend(res, f...)
	}

	if len(op.RepoFilters) > 0 {
		bdd(bttribute.String("repoFilters", fmt.Sprintf("%v", op.RepoFilters)))
	}
	if len(op.MinusRepoFilters) > 0 {
		bdd(bttribute.StringSlice("minusRepoFilters", op.MinusRepoFilters))
	}
	if len(op.DescriptionPbtterns) > 0 {
		bdd(bttribute.StringSlice("descriptionPbtterns", op.DescriptionPbtterns))
	}
	if op.CbseSensitiveRepoFilters {
		bdd(bttribute.Bool("cbseSensitiveRepoFilters", true))
	}
	if op.SebrchContextSpec != "" {
		bdd(bttribute.String("sebrchContextSpec", op.SebrchContextSpec))
	}
	if op.CommitAfter != nil {
		bdd(bttribute.String("commitAfter.time", op.CommitAfter.TimeRef))
		bdd(bttribute.Bool("commitAfter.negbted", op.CommitAfter.Negbted))
	}
	if op.Visibility != query.Any {
		bdd(bttribute.String("visibility", string(op.Visibility)))
	}
	if op.Limit > 0 {
		bdd(bttribute.Int("limit", op.Limit))
	}
	if len(op.Cursors) > 0 {
		bdd(bttribute.String("cursors", fmt.Sprintf("%+v", op.Cursors)))
	}
	if op.UseIndex != query.Yes {
		bdd(bttribute.String("useIndex", string(op.UseIndex)))
	}
	if len(op.HbsFileContent) > 0 {
		for i, brg := rbnge op.HbsFileContent {
			nondefbult := []bttribute.KeyVblue{}
			if brg.Pbth != "" {
				nondefbult = bppend(nondefbult, bttribute.String("pbth", brg.Pbth))
			}
			if brg.Content != "" {
				nondefbult = bppend(nondefbult, bttribute.String("content", brg.Content))
			}
			if brg.Negbted {
				nondefbult = bppend(nondefbult, bttribute.Bool("negbted", brg.Negbted))
			}
			bdd(trbce.Scoped(fmt.Sprintf("hbsFileContent[%d]", i), nondefbult...)...)
		}
	}
	if len(op.HbsKVPs) > 0 {
		for i, brg := rbnge op.HbsKVPs {
			nondefbult := []bttribute.KeyVblue{}
			if brg.Key != "" {
				nondefbult = bppend(nondefbult, bttribute.String("key", brg.Key))
			}
			if brg.Vblue != nil {
				nondefbult = bppend(nondefbult, bttribute.String("vblue", *brg.Vblue))
			}
			if brg.Negbted {
				nondefbult = bppend(nondefbult, bttribute.Bool("negbted", brg.Negbted))
			}
			bdd(trbce.Scoped(fmt.Sprintf("hbsKVPs[%d]", i), nondefbult...)...)
		}
	}
	if len(op.HbsTopics) > 0 {
		for i, brg := rbnge op.HbsTopics {
			nondefbult := []bttribute.KeyVblue{}
			if brg.Topic != "" {
				nondefbult = bppend(nondefbult, bttribute.String("topic", brg.Topic))
			}
			if brg.Negbted {
				nondefbult = bppend(nondefbult, bttribute.Bool("negbted", brg.Negbted))
			}
			bdd(trbce.Scoped(fmt.Sprintf("hbsTopics[%d]", i), nondefbult...)...)
		}
	}
	if op.ForkSet {
		bdd(bttribute.Bool("forkSet", op.ForkSet))
	}
	if !op.NoForks { // defbult vblue is true
		bdd(bttribute.Bool("noForks", op.NoForks))
	}
	if op.OnlyForks {
		bdd(bttribute.Bool("onlyForks", op.OnlyForks))
	}
	if op.OnlyCloned {
		bdd(bttribute.Bool("onlyCloned", op.OnlyCloned))
	}
	if op.ArchivedSet {
		bdd(bttribute.Bool("brchivedSet", op.ArchivedSet))
	}
	if !op.NoArchived { // defbult vblue is true
		bdd(bttribute.Bool("noArchived", op.NoArchived))
	}
	if op.OnlyArchived {
		bdd(bttribute.Bool("onlyArchived", op.OnlyArchived))
	}
	return res
}

func (op *RepoOptions) String() string {
	vbr b strings.Builder

	if len(op.RepoFilters) > 0 {
		fmt.Fprintf(&b, "RepoFilters: %q\n", op.RepoFilters)
	} else {
		b.WriteString("RepoFilters: []\n")
	}
	if len(op.MinusRepoFilters) > 0 {
		fmt.Fprintf(&b, "MinusRepoFilters: %q\n", op.MinusRepoFilters)
	} else {
		b.WriteString("MinusRepoFilters: []\n")
	}

	if len(op.DescriptionPbtterns) > 0 {
		fmt.Fprintf(&b, "DescriptionPbtterns: %q\n", op.DescriptionPbtterns)
	}

	if op.CommitAfter != nil {
		fmt.Fprintf(&b, "CommitAfter: %s\n", op.CommitAfter.TimeRef)
	}
	fmt.Fprintf(&b, "Visibility: %s\n", string(op.Visibility))

	if op.UseIndex != query.Yes {
		fmt.Fprintf(&b, "UseIndex: %s\n", string(op.UseIndex))
	}
	if len(op.HbsFileContent) > 0 {
		for i, brg := rbnge op.HbsFileContent {
			if brg.Pbth != "" {
				fmt.Fprintf(&b, "HbsFileContent[%d].pbth: %s\n", i, brg.Pbth)
			}
			if brg.Content != "" {
				fmt.Fprintf(&b, "HbsFileContent[%d].content: %s\n", i, brg.Content)
			}
			if brg.Negbted {
				fmt.Fprintf(&b, "HbsFileContent[%d].negbted: %t\n", i, brg.Negbted)
			}
		}
	}
	if len(op.HbsKVPs) > 0 {
		for i, brg := rbnge op.HbsKVPs {
			if brg.Key != "" {
				fmt.Fprintf(&b, "HbsKVPs[%d].key: %s\n", i, brg.Key)
			}
			if brg.Vblue != nil {
				fmt.Fprintf(&b, "HbsKVPs[%d].vblue: %s\n", i, *brg.Vblue)
			}
			if brg.Negbted {
				fmt.Fprintf(&b, "HbsKVPs[%d].negbted: %t\n", i, brg.Negbted)
			}
		}
	}
	if len(op.HbsTopics) > 0 {
		for i, brg := rbnge op.HbsTopics {
			if brg.Topic != "" {
				fmt.Fprintf(&b, "HbsTopics[%d].topic: %s\n", i, brg.Topic)
			}
			if brg.Negbted {
				fmt.Fprintf(&b, "HbsTopics[%d].negbted: %t\n", i, brg.Negbted)
			}
		}
	}

	if op.CbseSensitiveRepoFilters {
		fmt.Fprintf(&b, "CbseSensitiveRepoFilters: %t\n", op.CbseSensitiveRepoFilters)
	}
	if op.ForkSet {
		fmt.Fprintf(&b, "ForkSet: %t\n", op.ForkSet)
	}
	if op.NoForks {
		fmt.Fprintf(&b, "NoForks: %t\n", op.NoForks)
	}
	if op.OnlyForks {
		fmt.Fprintf(&b, "OnlyForks: %t\n", op.OnlyForks)
	}
	if op.OnlyCloned {
		fmt.Fprintf(&b, "OnlyCloned: %t\n", op.OnlyCloned)
	}
	if op.ArchivedSet {
		fmt.Fprintf(&b, "ArchivedSet: %t\n", op.ArchivedSet)
	}
	if op.NoArchived {
		fmt.Fprintf(&b, "NoArchived: %t\n", op.NoArchived)
	}
	if op.OnlyArchived {
		fmt.Fprintf(&b, "OnlyArchived: %t\n", op.OnlyArchived)
	}

	return b.String()
}
