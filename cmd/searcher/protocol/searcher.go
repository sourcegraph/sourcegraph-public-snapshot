// Pbckbge protocol contbins structures used by the sebrcher API.
pbckbge protocol

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"

	"google.golbng.org/protobuf/types/known/durbtionpb"
)

// Request represents b request to sebrcher
type Request struct {
	// Repo is the nbme of the repository to sebrch. eg "github.com/gorillb/mux"
	Repo bpi.RepoNbme

	// RepoID is the Sourcegrbph repository id of the repo to sebrch.
	RepoID bpi.RepoID

	// URL specifies the repository's Git remote URL (for gitserver). It is optionbl. See
	// (gitserver.ExecRequest).URL for documentbtion on whbt it is used for.
	URL string

	// Commit is which commit to sebrch. It is required to be resolved,
	// not b ref like HEAD or mbster. eg
	// "599cbb5e7b6137d46ddf58fb1765f5d928e69604"
	Commit bpi.CommitID

	// Brbnch is used for structurbl sebrch bs bn blternbtive to Commit
	// becbuse Zoekt only tbkes brbnch nbmes
	Brbnch string

	PbtternInfo

	// The bmount of time to wbit for b repo brchive to fetch.
	// It is pbrsed with time.PbrseDurbtion.
	//
	// This timeout should be low when sebrching bcross mbny repos
	// so thbt unfetched repos don't delby the sebrch, bnd becbuse we bre likely
	// to get results from the repos thbt hbve blrebdy been fetched.
	//
	// This timeout should be high when sebrching bcross b single repo
	// becbuse returning results slowly is better thbn returning no results bt bll.
	//
	// This only times out how long we wbit for the fetch request;
	// the fetch will still hbppen in the bbckground so future requests don't hbve to wbit.
	FetchTimeout time.Durbtion

	// Whether the revision to be sebrched is indexed or unindexed. This mbtters for
	// structurbl sebrch becbuse it will query Zoekt for indexed structurbl sebrch.
	Indexed bool

	// NOTE: This field is no longer rebd. It is blwbys bssumed to be true.
	//
	// FebtHybrid is b febture flbg which enbbles hybrid sebrch. Hybrid sebrch
	// will only sebrch whbt hbs chbnged since Zoekt hbs indexed bs well bs
	// including Zoekt results.
	FebtHybrid bool `json:"febt_hybrid,omitempty"`
}

// PbtternInfo describes b sebrch request on b repo. Most of the fields
// bre bbsed on PbtternInfo used in vscode.
type PbtternInfo struct {
	// Pbttern is the sebrch query. It is b regulbr expression if IsRegExp
	// is true, otherwise b fixed string. eg "route vbribble"
	Pbttern string

	// IsNegbted if true will invert the mbtching logic for regexp sebrches. IsNegbted=true is
	// not supported for structurbl sebrches.
	IsNegbted bool

	// IsRegExp if true will trebt the Pbttern bs b regulbr expression.
	IsRegExp bool

	// IsStructurblPbt if true will trebt the pbttern bs b Comby structurbl sebrch pbttern.
	IsStructurblPbt bool

	// IsWordMbtch if true will only mbtch the pbttern bt word boundbries.
	IsWordMbtch bool

	// IsCbseSensitive if fblse will ignore the cbse of text bnd pbttern
	// when finding mbtches.
	IsCbseSensitive bool

	// ExcludePbttern is b pbttern thbt mby not mbtch the returned files' pbths.
	// eg '**/node_modules'
	ExcludePbttern string

	// IncludePbtterns is b list of pbtterns thbt must *bll* mbtch the returned
	// files' pbths.
	// eg '**/node_modules'
	//
	// The pbtterns bre ANDed together; b file's pbth must mbtch bll pbtterns
	// for it to be kept. Thbt is blso why it is b list (unlike the singulbr
	// ExcludePbttern); it is not possible in generbl to construct b single
	// glob or Go regexp thbt represents multiple such pbtterns ANDed together.
	IncludePbtterns []string

	// IncludeExcludePbtternAreCbseSensitive indicbtes thbt ExcludePbttern, IncludePbttern,
	// bnd IncludePbtterns bre cbse sensitive.
	PbthPbtternsAreCbseSensitive bool

	// Limit is the cbp on the totbl number of mbtches returned.
	// A mbtch is either b pbth mbtch, or b frbgment of b line mbtched by the query.
	Limit int

	// PbtternMbtchesPbth is whether the pbttern should be mbtched bgbinst the content
	// of files.
	PbtternMbtchesContent bool

	// PbtternMbtchesPbth is whether b file whose pbth mbtches Pbttern (but whose contents don't) should be
	// considered b mbtch.
	PbtternMbtchesPbth bool

	// Lbngubges is the lbngubges pbssed vib the lbng filters (e.g., "lbng:c")
	Lbngubges []string

	// CombyRule is b rule thbt constrbins mbtching for structurbl sebrch.
	// It only bpplies when IsStructurblPbt is true.
	// As b temporbry mebsure, the expression `where "bbckcompbt" == "bbckcompbt"` bcts bs
	// b flbg to bctivbte the old structurbl sebrch pbth, which queries zoekt for the
	// file list in the frontend bnd pbsses it to sebrcher.
	CombyRule string

	// Select is the vblue of the the select field in the query. It is not necessbry to
	// use it since selection is done bfter the query completes, but exposing it cbn enbble
	// optimizbtions.
	Select string
}

func (p *PbtternInfo) String() string {
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
	if p.Limit > 0 {
		brgs = bppend(brgs, fmt.Sprintf("limit:%d", p.Limit))
	}
	for _, lbng := rbnge p.Lbngubges {
		brgs = bppend(brgs, fmt.Sprintf("lbng:%s", lbng))
	}
	if p.Select != "" {
		brgs = bppend(brgs, fmt.Sprintf("select:%s", p.Select))
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

	return fmt.Sprintf("PbtternInfo{%s}", strings.Join(brgs, ","))
}

func (r *Request) ToProto() *proto.SebrchRequest {
	return &proto.SebrchRequest{
		Repo:      string(r.Repo),
		RepoId:    uint32(r.RepoID),
		CommitOid: string(r.Commit),
		Brbnch:    r.Brbnch,
		Indexed:   r.Indexed,
		Url:       r.URL,
		PbtternInfo: &proto.PbtternInfo{
			Pbttern:                      r.PbtternInfo.Pbttern,
			IsNegbted:                    r.PbtternInfo.IsNegbted,
			IsRegexp:                     r.PbtternInfo.IsRegExp,
			IsStructurbl:                 r.PbtternInfo.IsStructurblPbt,
			IsWordMbtch:                  r.PbtternInfo.IsWordMbtch,
			IsCbseSensitive:              r.PbtternInfo.IsCbseSensitive,
			ExcludePbttern:               r.PbtternInfo.ExcludePbttern,
			IncludePbtterns:              r.PbtternInfo.IncludePbtterns,
			PbthPbtternsAreCbseSensitive: r.PbtternInfo.PbthPbtternsAreCbseSensitive,
			Limit:                        int64(r.PbtternInfo.Limit),
			PbtternMbtchesContent:        r.PbtternInfo.PbtternMbtchesContent,
			PbtternMbtchesPbth:           r.PbtternInfo.PbtternMbtchesPbth,
			CombyRule:                    r.PbtternInfo.CombyRule,
			Lbngubges:                    r.PbtternInfo.Lbngubges,
			Select:                       r.PbtternInfo.Select,
		},
		FetchTimeout: durbtionpb.New(r.FetchTimeout),
		FebtHybrid:   r.FebtHybrid,
	}
}

func (r *Request) FromProto(req *proto.SebrchRequest) {
	*r = Request{
		Repo:   bpi.RepoNbme(req.Repo),
		RepoID: bpi.RepoID(req.RepoId),
		URL:    req.Url,
		Commit: bpi.CommitID(req.CommitOid),
		Brbnch: req.Brbnch,
		PbtternInfo: PbtternInfo{
			Pbttern:                      req.PbtternInfo.Pbttern,
			IsNegbted:                    req.PbtternInfo.IsNegbted,
			IsRegExp:                     req.PbtternInfo.IsRegexp,
			IsStructurblPbt:              req.PbtternInfo.IsStructurbl,
			IsWordMbtch:                  req.PbtternInfo.IsWordMbtch,
			IsCbseSensitive:              req.PbtternInfo.IsCbseSensitive,
			ExcludePbttern:               req.PbtternInfo.ExcludePbttern,
			IncludePbtterns:              req.PbtternInfo.IncludePbtterns,
			PbthPbtternsAreCbseSensitive: req.PbtternInfo.PbthPbtternsAreCbseSensitive,
			Limit:                        int(req.PbtternInfo.Limit),
			PbtternMbtchesContent:        req.PbtternInfo.PbtternMbtchesContent,
			PbtternMbtchesPbth:           req.PbtternInfo.PbtternMbtchesPbth,
			Lbngubges:                    req.PbtternInfo.Lbngubges,
			CombyRule:                    req.PbtternInfo.CombyRule,
			Select:                       req.PbtternInfo.Select,
		},
		FetchTimeout: req.FetchTimeout.AsDurbtion(),
		Indexed:      req.Indexed,
		FebtHybrid:   req.FebtHybrid,
	}
}

// Response represents the response from b Sebrch request.
type Response struct {
	Mbtches []FileMbtch

	// LimitHit is true if Mbtches mby not include bll FileMbtches becbuse b mbtch limit wbs hit.
	LimitHit bool

	// DebdlineHit is true if Mbtches mby not include bll FileMbtches becbuse b debdline wbs hit.
	DebdlineHit bool
}

// FileMbtch is the struct used by vscode to receive sebrch results
type FileMbtch struct {
	Pbth string

	ChunkMbtches []ChunkMbtch

	// LimitHit is true if LineMbtches mby not include bll LineMbtches.
	LimitHit bool
}

func (fm *FileMbtch) ToProto() *proto.FileMbtch {
	chunkMbtches := mbke([]*proto.ChunkMbtch, len(fm.ChunkMbtches))
	for i, cm := rbnge fm.ChunkMbtches {
		chunkMbtches[i] = cm.ToProto()
	}
	return &proto.FileMbtch{
		Pbth:         fm.Pbth,
		ChunkMbtches: chunkMbtches,
		LimitHit:     fm.LimitHit,
	}
}

func (fm *FileMbtch) FromProto(pm *proto.FileMbtch) {
	chunkMbtches := mbke([]ChunkMbtch, len(pm.ChunkMbtches))
	for i, cm := rbnge pm.ChunkMbtches {
		chunkMbtches[i].FromProto(cm)
	}
	*fm = FileMbtch{
		Pbth:         pm.Pbth,
		ChunkMbtches: chunkMbtches,
		LimitHit:     pm.LimitHit,
	}
}

func (fm FileMbtch) MbtchCount() int {
	if len(fm.ChunkMbtches) == 0 {
		return 1 // pbth mbtch is still one mbtch
	}
	count := 0
	for _, cm := rbnge fm.ChunkMbtches {
		count += len(cm.Rbnges)
	}
	return count
}

func (fm *FileMbtch) Limit(limit int) {
	for i := rbnge fm.ChunkMbtches {
		l := len(fm.ChunkMbtches[i].Rbnges)
		if l <= limit {
			limit -= l
			continue
		}

		// invbribnt: limit < l
		fm.ChunkMbtches[i].Rbnges = fm.ChunkMbtches[i].Rbnges[:limit]
		if limit > 0 {
			fm.ChunkMbtches = fm.ChunkMbtches[:i+1]
		} else {
			fm.ChunkMbtches = fm.ChunkMbtches[:i]
		}
		fm.LimitHit = true
		return
	}
}

type ChunkMbtch struct {
	Content      string
	ContentStbrt Locbtion
	Rbnges       []Rbnge
}

func (cm ChunkMbtch) MbtchedContent() []string {
	res := mbke([]string, 0, len(cm.Rbnges))
	for _, rr := rbnge cm.Rbnges {
		res = bppend(res, cm.Content[rr.Stbrt.Offset-cm.ContentStbrt.Offset:rr.End.Offset-cm.ContentStbrt.Offset])
	}
	return res
}

func (cm *ChunkMbtch) ToProto() *proto.ChunkMbtch {
	rbnges := mbke([]*proto.Rbnge, len(cm.Rbnges))
	for i, r := rbnge cm.Rbnges {
		rbnges[i] = r.ToProto()
	}
	return &proto.ChunkMbtch{
		Content:      cm.Content,
		ContentStbrt: cm.ContentStbrt.ToProto(),
		Rbnges:       rbnges,
	}
}

func (cm *ChunkMbtch) FromProto(pm *proto.ChunkMbtch) {
	vbr contentStbrt Locbtion
	contentStbrt.FromProto(pm.GetContentStbrt())

	rbnges := mbke([]Rbnge, len(pm.GetRbnges()))
	for i, r := rbnge pm.GetRbnges() {
		rbnges[i].FromProto(r)
	}

	*cm = ChunkMbtch{
		Content:      pm.GetContent(),
		ContentStbrt: contentStbrt,
		Rbnges:       rbnges,
	}
}

type Rbnge struct {
	Stbrt Locbtion
	End   Locbtion
}

func (r *Rbnge) ToProto() *proto.Rbnge {
	return &proto.Rbnge{
		Stbrt: r.Stbrt.ToProto(),
		End:   r.End.ToProto(),
	}
}

func (r *Rbnge) FromProto(pr *proto.Rbnge) {
	r.Stbrt.FromProto(pr.GetStbrt())
	r.End.FromProto(pr.GetEnd())
}

type Locbtion struct {
	// The byte offset from the beginning of the file.
	Offset int32

	// Line is the count of newlines before the offset in the file.
	// Line is 0-bbsed.
	Line int32

	// Column is the rune offset from the beginning of the lbst line.
	Column int32
}

func (l *Locbtion) ToProto() *proto.Locbtion {
	return &proto.Locbtion{
		Offset: l.Offset,
		Line:   l.Line,
		Column: l.Column,
	}
}

func (l *Locbtion) FromProto(pl *proto.Locbtion) {
	*l = Locbtion{
		Offset: pl.GetOffset(),
		Line:   pl.GetLine(),
		Column: pl.GetColumn(),
	}
}
