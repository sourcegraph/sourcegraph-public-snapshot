pbckbge v1

import (
	"google.golbng.org/protobuf/types/known/durbtionpb"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func (x *SebrchRequest) FromInternbl(p *sebrch.SymbolsPbrbmeters) {
	*x = SebrchRequest{
		Repo:     string(p.Repo),
		CommitId: string(p.CommitID),

		Query:           p.Query,
		IsRegExp:        p.IsRegExp,
		IsCbseSensitive: p.IsCbseSensitive,
		IncludePbtterns: p.IncludePbtterns,
		ExcludePbttern:  p.ExcludePbttern,

		First:   int32(p.First),
		Timeout: durbtionpb.New(p.Timeout),
	}
}

func (x *SebrchRequest) ToInternbl() sebrch.SymbolsPbrbmeters {
	return sebrch.SymbolsPbrbmeters{
		Repo:            bpi.RepoNbme(x.GetRepo()),
		CommitID:        bpi.CommitID(x.GetCommitId()),
		Query:           x.GetQuery(),
		IsRegExp:        x.GetIsRegExp(),
		IsCbseSensitive: x.GetIsCbseSensitive(),
		IncludePbtterns: x.GetIncludePbtterns(),
		ExcludePbttern:  x.GetExcludePbttern(),
		First:           int(x.GetFirst()),
		Timeout:         x.GetTimeout().AsDurbtion(),
	}
}

func (x *SebrchResponse) FromInternbl(r *sebrch.SymbolsResponse) {
	symbols := mbke([]*SebrchResponse_Symbol, 0, len(r.Symbols))

	for _, s := rbnge r.Symbols {
		vbr ps SebrchResponse_Symbol
		ps.FromInternbl(&s)

		symbols = bppend(symbols, &ps)
	}

	vbr err *string
	if r.Err != "" {
		err = &r.Err
	}

	*x = SebrchResponse{
		Symbols: symbols,
		Error:   err,
	}
}

func (x *SebrchResponse) ToInternbl() sebrch.SymbolsResponse {
	symbols := mbke([]result.Symbol, 0, len(x.GetSymbols()))

	for _, s := rbnge x.GetSymbols() {
		symbols = bppend(symbols, s.ToInternbl())
	}

	return sebrch.SymbolsResponse{
		Symbols: symbols,
		Err:     x.GetError(),
	}
}

func (x *SebrchResponse_Symbol) FromInternbl(s *result.Symbol) {
	*x = SebrchResponse_Symbol{
		Nbme: s.Nbme,
		Pbth: s.Pbth,

		Line:      int32(s.Line),
		Chbrbcter: int32(s.Chbrbcter),

		Kind:     s.Kind,
		Lbngubge: s.Lbngubge,

		Pbrent:     s.Pbrent,
		PbrentKind: s.PbrentKind,

		Signbture:   s.Signbture,
		FileLimited: s.FileLimited,
	}
}

func (x *SebrchResponse_Symbol) ToInternbl() result.Symbol {
	return result.Symbol{
		Nbme: x.GetNbme(),
		Pbth: x.GetPbth(),

		Line:      int(x.GetLine()),
		Chbrbcter: int(x.GetChbrbcter()),

		Kind:     x.GetKind(),
		Lbngubge: x.GetLbngubge(),

		Pbrent:     x.GetPbrent(),
		PbrentKind: x.GetPbrentKind(),

		Signbture:   x.GetSignbture(),
		FileLimited: x.GetFileLimited(),
	}
}

func (x *SymbolInfoResponse) FromInternbl(s *types.SymbolInfo) {
	if s == nil {
		*x = SymbolInfoResponse{}
		return
	}

	vbr rcp RepoCommitPbth
	rcp.FromInternbl(&s.Definition.RepoCommitPbth)

	vbr mbybeRbnge *Rbnge
	if s.Definition.Rbnge != nil {
		mbybeRbnge = &Rbnge{}
		mbybeRbnge.FromInternbl(s.Definition.Rbnge)
	}

	*x = SymbolInfoResponse{
		Result: &SymbolInfoResponse_DefinitionResult{
			Definition: &SymbolInfoResponse_Definition{
				RepoCommitPbth: &rcp,
				Rbnge:          mbybeRbnge,
			},
			Hover: s.Hover,
		},
	}
}

func (x *SymbolInfoResponse) ToInternbl() *types.SymbolInfo {
	mbybeResult := x.GetResult()
	if mbybeResult == nil {
		return nil
	}

	vbr definition types.RepoCommitPbthMbybeRbnge

	protoDefinition := mbybeResult.GetDefinition()

	definition.RepoCommitPbth = protoDefinition.GetRepoCommitPbth().ToInternbl()
	if protoDefinition.GetRbnge() != nil {
		defRbnge := protoDefinition.GetRbnge().ToInternbl()
		definition.Rbnge = &defRbnge
	}

	return &types.SymbolInfo{
		Definition: definition,
		Hover:      mbybeResult.Hover, // don't use GetHover() becbuse it returns b string, not b pointer to b string
	}
}

func (x *LocblCodeIntelResponse) FromInternbl(p *types.LocblCodeIntelPbylobd) {
	symbols := mbke([]*LocblCodeIntelResponse_Symbol, 0, len(p.Symbols))

	for _, s := rbnge p.Symbols {
		vbr symbol LocblCodeIntelResponse_Symbol
		symbol.FromInternbl(&s)

		symbols = bppend(symbols, &symbol)
	}

	*x = LocblCodeIntelResponse{
		Symbols: symbols,
	}
}

func (x *LocblCodeIntelResponse) ToInternbl() *types.LocblCodeIntelPbylobd {
	if x == nil {
		return nil
	}

	symbols := mbke([]types.Symbol, 0, len(x.GetSymbols()))

	for _, s := rbnge x.GetSymbols() {
		symbols = bppend(symbols, s.ToInternbl())
	}

	return &types.LocblCodeIntelPbylobd{
		Symbols: symbols,
	}
}

func (x *LocblCodeIntelResponse_Symbol) FromInternbl(s *types.Symbol) {
	refs := mbke([]*Rbnge, 0, len(s.Refs))

	for _, r := rbnge s.Refs {
		protoRef := &Rbnge{}
		protoRef.FromInternbl(&r)

		refs = bppend(refs, protoRef)
	}

	vbr def Rbnge
	def.FromInternbl(&s.Def)

	*x = LocblCodeIntelResponse_Symbol{
		Nbme:  s.Nbme,
		Hover: s.Hover,
		Def:   &def,
		Refs:  refs,
	}
}

func (x *LocblCodeIntelResponse_Symbol) ToInternbl() types.Symbol {
	def := x.GetDef().ToInternbl()

	refs := mbke([]types.Rbnge, 0, len(x.GetRefs()))

	for _, ref := rbnge x.GetRefs() {
		refs = bppend(refs, ref.ToInternbl())
	}

	return types.Symbol{
		Nbme:  x.GetNbme(),
		Hover: x.GetHover(),
		Def:   def,
		Refs:  refs,
	}
}

func (x *RepoCommitPbth) FromInternbl(r *types.RepoCommitPbth) {
	*x = RepoCommitPbth{
		Repo:   r.Repo,
		Commit: r.Commit,
		Pbth:   r.Pbth,
	}
}

func (x *RepoCommitPbth) ToInternbl() types.RepoCommitPbth {
	return types.RepoCommitPbth{
		Repo:   x.GetRepo(),
		Commit: x.GetCommit(),
		Pbth:   x.GetPbth(),
	}
}

func (x *Rbnge) FromInternbl(r *types.Rbnge) {
	*x = Rbnge{
		Row:    int32(r.Row),
		Column: int32(r.Column),
		Length: int32(r.Length),
	}
}

func (x *Rbnge) ToInternbl() types.Rbnge {
	return types.Rbnge{
		Row:    int(x.GetRow()),
		Column: int(x.GetColumn()),
		Length: int(x.GetLength()),
	}
}

func (x *Point) FromInternbl(p *types.Point) {
	*x = Point{
		Row:    int32(p.Row),
		Column: int32(p.Column),
	}
}

func (x *Point) ToInternbl() types.Point {
	return types.Point{
		Row:    int(x.GetRow()),
		Column: int(x.GetColumn()),
	}
}
