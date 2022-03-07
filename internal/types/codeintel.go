package types

import (
	"fmt"
	"time"
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

type LocalCodeIntelPayload struct {
	Symbols []Symbol `json:"symbols"`
}

type RepoCommitPathRange struct {
	RepoCommitPath
	Range
}

type Symbol struct {
	Hover *string `json:"hover,omitempty"`
	Def   *Range  `json:"def,omitempty"`
	Refs  []Range `json:"refs,omitempty"`
}

func (s Symbol) String() string {
	hover := "<nil>"
	if s.Hover != nil {
		hover = *s.Hover
	}
	def := "<nil>"
	if s.Def != nil {
		def = s.Def.String()
	}
	return fmt.Sprintf("Symbol{Hover: %q, Def: %s, Refs: %+v", hover, def, s.Refs)
}

type Range struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Length int `json:"length"`
}

func (r Range) String() string {
	return fmt.Sprintf("%d:%d:%d", r.Row, r.Column, r.Length)
}

type SymbolInfo struct {
	Definition RepoCommitPathRange `json:"definition"`
	Hover      *string             `json:"hover,omitempty"`
}
