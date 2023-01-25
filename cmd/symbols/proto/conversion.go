package proto

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (x *SearchRequest) FromInternal(p *search.SymbolsParameters) {
	if x == nil || p == nil {
		return
	}

	*x = SearchRequest{
		Repo:     string(p.Repo),
		CommitId: string(p.CommitID),

		Query:           p.Query,
		IsRegExp:        p.IsRegExp,
		IsCaseSensitive: p.IsCaseSensitive,
		IncludePatterns: p.IncludePatterns,
		ExcludePattern:  p.ExcludePattern,

		First:   int32(p.First),
		Timeout: durationpb.New(p.Timeout),
	}
}

func (x *SearchRequest) ToInternal() search.SymbolsParameters {
	if x == nil {
		return search.SymbolsParameters{}
	}

	return search.SymbolsParameters{
		Repo:            api.RepoName(x.GetRepo()), // TODO@ggilmore: This api.RepoName is just a go type alias - is it worth creating a new message type just for this?
		CommitID:        api.CommitID(x.GetCommitId()),
		Query:           x.GetQuery(),
		IsRegExp:        x.GetIsRegExp(),
		IsCaseSensitive: x.GetIsCaseSensitive(),
		IncludePatterns: x.GetIncludePatterns(),
		ExcludePattern:  x.GetExcludePattern(),
		First:           int(x.GetFirst()),
		Timeout:         x.GetTimeout().AsDuration(),
	}
}

func (x *SymbolsResponse) FromInternal(r *search.SymbolsResponse) {
	if x == nil || r == nil {
		return
	}

	var symbols []*SymbolsResponse_Symbol
	for _, s := range r.Symbols {
		var ps SymbolsResponse_Symbol
		ps.FromInternal(&s)

		symbols = append(symbols, &ps)
	}

	var err *string
	if r.Err != "" {
		err = &r.Err
	}

	*x = SymbolsResponse{
		Symbols: symbols,
		Error:   err,
	}
}

func (x *SymbolsResponse) ToInternal() search.SymbolsResponse {
	if x == nil {
		return search.SymbolsResponse{}
	}

	var symbols []result.Symbol

	for _, s := range x.GetSymbols() {
		symbols = append(symbols, s.ToInternal())
	}

	return search.SymbolsResponse{
		Symbols: symbols,
		Err:     x.GetError(),
	}
}

func (x *SymbolsResponse_Symbol) FromInternal(s *result.Symbol) {
	if x == nil || s == nil {
		return
	}

	*x = SymbolsResponse_Symbol{
		Name: s.Name,
		Path: s.Path,

		Line:      int32(s.Line),
		Character: int32(s.Character),

		Kind:     s.Kind,
		Language: s.Language,

		Parent:     s.Parent,
		ParentKind: s.ParentKind,

		Signature:   s.Signature,
		FileLimited: s.FileLimited,
	}
}

func (x *SymbolsResponse_Symbol) ToInternal() result.Symbol {
	if x == nil {
		return result.Symbol{}
	}

	return result.Symbol{
		Name: x.GetName(),
		Path: x.GetPath(),

		Line:      int(x.GetLine()),
		Character: int(x.GetCharacter()),

		Kind:     x.GetKind(),
		Language: x.GetLanguage(),

		Parent:     x.GetParent(),
		ParentKind: x.GetParentKind(),

		Signature:   x.GetSignature(),
		FileLimited: x.GetFileLimited(),
	}
}

func (x *SymbolInfoResponse) FromInternal(s *types.SymbolInfo) {
	if x == nil {
		return
	}

	if s == nil {
		*x = SymbolInfoResponse{}
		return
	}

	var rcp RepoCommitPath
	rcp.FromInternal(&s.Definition.RepoCommitPath)

	var maybeRange *Range
	if s.Definition.Range != nil {
		maybeRange = &Range{}
		maybeRange.FromInternal(s.Definition.Range)
	}

	*x = SymbolInfoResponse{
		Result: &SymbolInfoResponse_DefinitionResult{
			Definition: &SymbolInfoResponse_Definition{
				RepoCommitPath: &rcp,
				Range:          maybeRange,
			},
			Hover: s.Hover,
		},
	}
}

func (x *SymbolInfoResponse) ToInternal() types.SymbolInfo {
	if x == nil {
		return types.SymbolInfo{}
	}

	maybeResult := x.GetResult()
	if maybeResult == nil {
		return types.SymbolInfo{}
	}

	var definition types.RepoCommitPathMaybeRange

	protoDefinition := maybeResult.GetDefinition()

	definition.RepoCommitPath = protoDefinition.GetRepoCommitPath().ToInternal()
	if protoDefinition.GetRange() != nil {
		defRange := protoDefinition.GetRange().ToInternal()
		definition.Range = &defRange
	}

	return types.SymbolInfo{
		Definition: definition,
		Hover:      maybeResult.Hover, // don't use GetHover() because it returns a string, not a pointer to a string
	}
}

func (x *LocalCodeIntelResponse) FromInternal(p *types.LocalCodeIntelPayload) {
	if x == nil || p == nil {
		return
	}

	var symbols []*LocalCodeIntelResponse_Symbol

	for _, s := range p.Symbols {
		var symbol LocalCodeIntelResponse_Symbol
		symbol.FromInternal(&s)

		symbols = append(symbols, &symbol)
	}

	*x = LocalCodeIntelResponse{
		Symbols: symbols,
	}
}

func (x *LocalCodeIntelResponse) ToInternal() types.LocalCodeIntelPayload {
	if x == nil {
		return types.LocalCodeIntelPayload{}
	}

	var symbols []types.Symbol

	for _, s := range x.GetSymbols() {
		symbols = append(symbols, s.ToInternal())
	}

	return types.LocalCodeIntelPayload{
		Symbols: symbols,
	}
}

func (x *LocalCodeIntelResponse_Symbol) FromInternal(s *types.Symbol) {
	if x == nil || s == nil {
		return
	}

	var refs []*Range

	for _, r := range s.Refs {
		protoRef := &Range{}
		protoRef.FromInternal(&r)

		refs = append(refs, protoRef)
	}

	var def Range
	def.FromInternal(&s.Def)

	*x = LocalCodeIntelResponse_Symbol{
		Name:  s.Name,
		Hover: s.Hover,
		Def:   &def,
		Refs:  refs,
	}
}

func (x *LocalCodeIntelResponse_Symbol) ToInternal() types.Symbol {
	if x == nil {
		return types.Symbol{}
	}

	def := x.GetDef().ToInternal()

	var refs []types.Range

	for _, ref := range x.GetRefs() {
		refs = append(refs, ref.ToInternal())
	}

	return types.Symbol{
		Name:  x.GetName(),
		Hover: x.GetHover(),
		Def:   def,
		Refs:  refs,
	}
}

func (x *RepoCommitPath) FromInternal(r *types.RepoCommitPath) {
	if x == nil || r == nil {
		return
	}

	*x = RepoCommitPath{
		Repo:   r.Repo,
		Commit: r.Commit,
		Path:   r.Path,
	}
}

func (x *RepoCommitPath) ToInternal() types.RepoCommitPath {
	if x == nil {
		return types.RepoCommitPath{}
	}

	return types.RepoCommitPath{
		Repo:   x.GetRepo(),
		Commit: x.GetCommit(),
		Path:   x.GetPath(),
	}
}

func (x *Range) FromInternal(r *types.Range) {
	if x == nil || r == nil {
		return
	}

	*x = Range{
		Row:    int32(r.Row),
		Column: int32(r.Column),
		Length: int32(r.Length),
	}
}

func (x *Range) ToInternal() types.Range {
	if x == nil {
		return types.Range{}
	}

	return types.Range{
		Row:    int(x.GetRow()),
		Column: int(x.GetColumn()),
		Length: int(x.GetLength()),
	}
}

func (x *Point) FromInternal(p *types.Point) {
	if x == nil || p == nil {
		return
	}

	*x = Point{
		Row:    int32(p.Row),
		Column: int32(p.Column),
	}
}

func (x *Point) ToInternal() types.Point {
	if x == nil {
		return types.Point{}
	}

	return types.Point{
		Row:    int(x.GetRow()),
		Column: int(x.GetColumn()),
	}
}
