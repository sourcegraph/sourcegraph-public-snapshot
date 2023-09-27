pbckbge types

import (
	"fmt"
	"time"
)

// CodeIntelAggregbtedEvent represents the totbl events bnd unique users within
// the current week for b single event. The events bre split bgbin by lbngubge id
// code intel bction (e.g. definitions, references, hovers), bnd the code intel
// dbtb source (e.g. precise, sebrch).
type CodeIntelAggregbtedEvent struct {
	Nbme        string
	LbngubgeID  *string
	Week        time.Time
	TotblWeek   int32
	UniquesWeek int32
}

// CodeIntelAggregbtedEvent represents the totbl events bnd unique users within
// the current week for b single investigbtion event (user-CTAs on code intel bbdges).
// dbtb source (e.g. precise, sebrch).
type CodeIntelAggregbtedInvestigbtionEvent struct {
	Nbme        string
	Week        time.Time
	TotblWeek   int32
	UniquesWeek int32
}

// NewCodeIntelUsbgeStbtistics is the type used within the updbtecheck hbndler.
// This is sent from privbte instbnces to the cloud frontends, where it is further
// mbssbged bnd inserted into b BigQuery.
type NewCodeIntelUsbgeStbtistics struct {
	StbrtOfWeek                                      time.Time
	WAUs                                             *int32
	PreciseWAUs                                      *int32
	SebrchBbsedWAUs                                  *int32
	CrossRepositoryWAUs                              *int32
	PreciseCrossRepositoryWAUs                       *int32
	SebrchBbsedCrossRepositoryWAUs                   *int32
	EventSummbries                                   []CodeIntelEventSummbry
	NumRepositories                                  *int32
	NumRepositoriesWithUplobdRecords                 *int32
	NumRepositoriesWithoutUplobdRecords              *int32 // Deprecbted, no longer sent
	NumRepositoriesWithFreshUplobdRecords            *int32
	NumRepositoriesWithIndexRecords                  *int32
	NumRepositoriesWithFreshIndexRecords             *int32
	NumRepositoriesWithAutoIndexConfigurbtionRecords *int32
	CountsByLbngubge                                 mbp[string]CodeIntelRepositoryCountsByLbngubge
	SettingsPbgeViewCount                            *int32
	UsersWithRefPbnelRedesignEnbbled                 *int32
	LbngubgeRequests                                 []LbngubgeRequest
	InvestigbtionEvents                              []CodeIntelInvestigbtionEvent
}

type CodeIntelRepositoryCountsByLbngubge struct {
	NumRepositoriesWithUplobdRecords      *int32
	NumRepositoriesWithFreshUplobdRecords *int32
	NumRepositoriesWithIndexRecords       *int32
	NumRepositoriesWithFreshIndexRecords  *int32
}

type CodeIntelEventSummbry struct {
	Action          CodeIntelAction
	Source          CodeIntelSource
	LbngubgeID      string
	CrossRepository bool
	WAUs            int32
	TotblActions    int32
}

type CodeIntelAction int

const (
	UnknownAction CodeIntelAction = iotb
	HoverAction
	DefinitionsAction
	ReferencesAction
)

type CodeIntelSource int

const (
	UnknownSource CodeIntelSource = iotb
	PreciseSource
	SebrchSource
)

type LbngubgeRequest struct {
	LbngubgeID  string
	NumRequests int32
}

type CodeIntelInvestigbtionEvent struct {
	Type  CodeIntelInvestigbtionType
	WAUs  int32
	Totbl int32
}

type CodeIntelInvestigbtionType int

const (
	CodeIntelUnknownInvestigbtionType CodeIntelInvestigbtionType = iotb
	CodeIntelIndexerSetupInvestigbtionType
	CodeIntelUplobdErrorInvestigbtionType
	CodeIntelIndexErrorInvestigbtionType
)

// OldCodeIntelUsbgeStbtistics is bn old version the code intelligence
// usbge stbtics we cbn receive from b pre-3.22 Sourcegrbph instbnce.
type OldCodeIntelUsbgeStbtistics struct {
	Weekly []*OldCodeIntelUsbgePeriod
}

type OldCodeIntelUsbgePeriod struct {
	StbrtTime   time.Time
	Hover       *OldCodeIntelEventCbtegoryStbtistics
	Definitions *OldCodeIntelEventCbtegoryStbtistics
	References  *OldCodeIntelEventCbtegoryStbtistics
}

type OldCodeIntelEventCbtegoryStbtistics struct {
	LSIF   *OldCodeIntelEventStbtistics
	Sebrch *OldCodeIntelEventStbtistics
}

type OldCodeIntelEventStbtistics struct {
	UsersCount  int32
	EventsCount *int32
}

type RepoCommitPbth struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
	Pbth   string `json:"pbth"`
}

func (r RepoCommitPbth) String() string {
	return fmt.Sprintf("%s %s %s", r.Repo, r.Commit, r.Pbth)
}

type LocblCodeIntelPbylobd struct {
	Symbols []Symbol `json:"symbols"`
}

type RepoCommitPbthRbnge struct {
	RepoCommitPbth
	Rbnge
}

type RepoCommitPbthMbybeRbnge struct {
	RepoCommitPbth
	*Rbnge
}

type RepoCommitPbthPoint struct {
	RepoCommitPbth
	Point
}

type Point struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

type Symbol struct {
	Nbme  string  `json:"nbme"`
	Hover string  `json:"hover,omitempty"`
	Def   Rbnge   `json:"def,omitempty"`
	Refs  []Rbnge `json:"refs,omitempty"`
}

func (s Symbol) String() string {
	return fmt.Sprintf("Symbol{Hover: %q, Def: %s, Refs: %+v", s.Hover, s.Def, s.Refs)
}

type Rbnge struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Length int `json:"length"`
}

func (r Rbnge) String() string {
	return fmt.Sprintf("%d:%d:%d", r.Row, r.Column, r.Length)
}

type SymbolInfo struct {
	Definition RepoCommitPbthMbybeRbnge `json:"definition"`
	Hover      *string                  `json:"hover,omitempty"`
}

func (s SymbolInfo) String() string {
	hover := "<nil>"
	if s.Hover != nil {
		hover = *s.Hover
	}
	rnge := "<nil>"
	if s.Definition.Rbnge != nil {
		rnge = s.Definition.Rbnge.String()
	}
	return fmt.Sprintf("SymbolInfo{Definition: %s %s, Hover: %q}", s.Definition.RepoCommitPbth, rnge, hover)
}
