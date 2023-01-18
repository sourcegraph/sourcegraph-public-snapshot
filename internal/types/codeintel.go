package types

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/proto"
)

// CodeIntelAggregatedEvent represents the total events and unique users within
// the current week for a single event. The events are split again by language id
// code intel action (e.g. defintions, references, hovers), and the code intel
// data source (e.g. precise, search).
type CodeIntelAggregatedEvent struct {
	Name        string
	LanguageID  *string
	Week        time.Time
	TotalWeek   int32
	UniquesWeek int32
}

// CodeIntelAggregatedEvent represents the total events and unique users within
// the current week for a single investigation event (user-CTAs on code intel badges).
// data source (e.g. precise, search).
type CodeIntelAggregatedInvestigationEvent struct {
	Name        string
	Week        time.Time
	TotalWeek   int32
	UniquesWeek int32
}

// NewCodeIntelUsageStatistics is the type used within the updatecheck handler.
// This is sent from private instances to the cloud frontends, where it is further
// massaged and inserted into a BigQuery.
type NewCodeIntelUsageStatistics struct {
	StartOfWeek                                      time.Time
	WAUs                                             *int32
	PreciseWAUs                                      *int32
	SearchBasedWAUs                                  *int32
	CrossRepositoryWAUs                              *int32
	PreciseCrossRepositoryWAUs                       *int32
	SearchBasedCrossRepositoryWAUs                   *int32
	EventSummaries                                   []CodeIntelEventSummary
	NumRepositories                                  *int32
	NumRepositoriesWithUploadRecords                 *int32
	NumRepositoriesWithoutUploadRecords              *int32 // Deprecated, no longer sent
	NumRepositoriesWithFreshUploadRecords            *int32
	NumRepositoriesWithIndexRecords                  *int32
	NumRepositoriesWithFreshIndexRecords             *int32
	NumRepositoriesWithAutoIndexConfigurationRecords *int32
	CountsByLanguage                                 map[string]CodeIntelRepositoryCountsByLanguage
	SettingsPageViewCount                            *int32
	UsersWithRefPanelRedesignEnabled                 *int32
	LanguageRequests                                 []LanguageRequest
	InvestigationEvents                              []CodeIntelInvestigationEvent
}

type CodeIntelRepositoryCountsByLanguage struct {
	NumRepositoriesWithUploadRecords      *int32
	NumRepositoriesWithFreshUploadRecords *int32
	NumRepositoriesWithIndexRecords       *int32
	NumRepositoriesWithFreshIndexRecords  *int32
}

type CodeIntelEventSummary struct {
	Action          CodeIntelAction
	Source          CodeIntelSource
	LanguageID      string
	CrossRepository bool
	WAUs            int32
	TotalActions    int32
}

type CodeIntelAction int

const (
	UnknownAction CodeIntelAction = iota
	HoverAction
	DefinitionsAction
	ReferencesAction
)

type CodeIntelSource int

const (
	UnknownSource CodeIntelSource = iota
	PreciseSource
	SearchSource
)

type LanguageRequest struct {
	LanguageID  string
	NumRequests int32
}

type CodeIntelInvestigationEvent struct {
	Type  CodeIntelInvestigationType
	WAUs  int32
	Total int32
}

type CodeIntelInvestigationType int

const (
	CodeIntelUnknownInvestigationType CodeIntelInvestigationType = iota
	CodeIntelIndexerSetupInvestigationType
	CodeIntelUploadErrorInvestigationType
	CodeIntelIndexErrorInvestigationType
)

// OldCodeIntelUsageStatistics is an old version the code intelligence
// usage statics we can receive from a pre-3.22 Sourcegraph instance.
type OldCodeIntelUsageStatistics struct {
	Weekly []*OldCodeIntelUsagePeriod
}

type OldCodeIntelUsagePeriod struct {
	StartTime   time.Time
	Hover       *OldCodeIntelEventCategoryStatistics
	Definitions *OldCodeIntelEventCategoryStatistics
	References  *OldCodeIntelEventCategoryStatistics
}

type OldCodeIntelEventCategoryStatistics struct {
	LSIF   *OldCodeIntelEventStatistics
	Search *OldCodeIntelEventStatistics
}

type OldCodeIntelEventStatistics struct {
	UsersCount  int32
	EventsCount *int32
}

type RepoCommitPath struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
	Path   string `json:"path"`
}

func (r RepoCommitPath) String() string {
	return fmt.Sprintf("%s %s %s", r.Repo, r.Commit, r.Path)
}

func (r *RepoCommitPath) ToProto() *proto.RepoCommitPath {
	return &proto.RepoCommitPath{
		Repo:   r.Repo,
		Commit: r.Commit,
		Path:   r.Path,
	}
}

func (r *RepoCommitPath) FromProto(p *proto.RepoCommitPath) {
	r.Repo = p.Repo
	r.Commit = p.Commit
	r.Path = p.Path
}

type LocalCodeIntelPayload struct {
	Symbols []Symbol `json:"symbols"`
}

func (p *LocalCodeIntelPayload) ToProto() *proto.LocalCodeIntelResponse {
	var symbols []*proto.LocalCodeIntelResponse_Symbol

	for _, s := range p.Symbols {
		symbols = append(symbols, s.ToProto())
	}

	return &proto.LocalCodeIntelResponse{
		Symbols: symbols,
	}
}

func (p *LocalCodeIntelPayload) FromProto(r *proto.LocalCodeIntelResponse) {
	var symbols []Symbol

	for _, s := range r.Symbols {
		var symbol Symbol
		symbol.FromProto(s)

		symbols = append(symbols, symbol)
	}

	p.Symbols = symbols
}

type RepoCommitPathRange struct {
	RepoCommitPath
	Range
}

type RepoCommitPathMaybeRange struct {
	RepoCommitPath
	*Range
}

type RepoCommitPathPoint struct {
	RepoCommitPath
	Point
}

type Point struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

func (p *Point) ToProto() *proto.Point {
	return &proto.Point{
		Row:    int32(p.Row),
		Column: int32(p.Column),
	}
}

func (p *Point) FromProto(pp *proto.Point) {
	p.Row = int(pp.Row)
	p.Column = int(pp.Column)
}

type Symbol struct {
	Name  string  `json:"name"`
	Hover string  `json:"hover,omitempty"`
	Def   Range   `json:"def,omitempty"`
	Refs  []Range `json:"refs,omitempty"`
}

func (s *Symbol) ToProto() *proto.LocalCodeIntelResponse_Symbol {
	var refs []*proto.Range

	for _, r := range s.Refs {
		refs = append(refs, r.ToProto())
	}

	return &proto.LocalCodeIntelResponse_Symbol{
		Name:  s.Name,
		Hover: s.Hover,
		Def:   s.Def.ToProto(),
		Refs:  refs,
	}
}

func (s *Symbol) FromProto(p *proto.LocalCodeIntelResponse_Symbol) {
	s.Name = p.Name
	s.Hover = p.Hover

	s.Def.FromProto(p.Def)

	var refs []Range

	for _, r := range p.Refs {
		var ref Range
		ref.FromProto(r)

		refs = append(refs, ref)
	}
}

func (s Symbol) String() string {
	return fmt.Sprintf("Symbol{Hover: %q, Def: %s, Refs: %+v", s.Hover, s.Def, s.Refs)
}

type Range struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Length int `json:"length"`
}

func (r *Range) ToProto() *proto.Range {
	return &proto.Range{
		Row:    int32(r.Row),
		Column: int32(r.Column),
		Length: int32(r.Length),
	}
}

func (r *Range) FromProto(p *proto.Range) {
	r.Row = int(p.Row)
	r.Column = int(p.Column)
	r.Length = int(p.Length)
}

func (r Range) String() string {
	return fmt.Sprintf("%d:%d:%d", r.Row, r.Column, r.Length)
}

type SymbolInfo struct {
	Definition RepoCommitPathMaybeRange `json:"definition"`
	Hover      *string                  `json:"hover,omitempty"`
}

func (si *SymbolInfo) ToProto() *proto.SymbolInfoResponse {
	var response proto.SymbolInfoResponse

	response.Definition.RepoCommitPath = si.Definition.RepoCommitPath.ToProto()
	if si.Definition.Range != nil {
		response.Definition.Range = si.Definition.Range.ToProto()
	}

	response.Hover = si.Hover

	return &response
}

func (si *SymbolInfo) FromProto(p *proto.SymbolInfoResponse) {
	si.Definition.RepoCommitPath.FromProto(p.Definition.RepoCommitPath)
	if p.Definition.Range != nil {
		si.Definition.Range.FromProto(p.Definition.Range)
	}

	si.Hover = p.Hover
}

func (s SymbolInfo) String() string {
	hover := "<nil>"
	if s.Hover != nil {
		hover = *s.Hover
	}
	rnge := "<nil>"
	if s.Definition.Range != nil {
		rnge = s.Definition.Range.String()
	}
	return fmt.Sprintf("SymbolInfo{Definition: %s %s, Hover: %q}", s.Definition.RepoCommitPath, rnge, hover)
}
