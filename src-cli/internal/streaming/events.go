package streaming

import (
	"bytes"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EventMatch is an interface which only the top level match event types
// implement. Use this for your results slice rather than interface{}.
type EventMatch interface {
	// private marker method so only top level event match types are allowed.
	eventMatch()
}

// EventContentMatch is a subset of zoekt.FileMatch for our Event API.
type EventContentMatch struct {
	// Type is always ContentMatchType. Included here for marshalling.
	Type MatchType `json:"type"`

	Path         string       `json:"path"`
	Repository   string       `json:"repository"`
	Branches     []string     `json:"branches,omitempty"`
	Commit       string       `json:"commit,omitempty"`
	ChunkMatches []ChunkMatch `json:"chunkMatches"`
}

type ChunkMatch struct {
	Content      string   `json:"content"`
	ContentStart Location `json:"contentStart"`
	Ranges       []Range  `json:"ranges"`
}

type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

func (e *EventContentMatch) eventMatch() {}

// EventPathMatch is a subset of zoekt.FileMatch for our Event API.
type EventPathMatch struct {
	// Type is always PathMatchType. Included here for marshalling.
	Type MatchType `json:"type"`

	Path       string   `json:"path"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
	Commit     string   `json:"commit,omitempty"`
}

func (e *EventPathMatch) eventMatch() {}

// EventLineMatch is a subset of zoekt.LineMatch for our Event API.
type EventLineMatch struct {
	Line             string     `json:"line"`
	LineNumber       int32      `json:"lineNumber"`
	OffsetAndLengths [][2]int32 `json:"offsetAndLengths"`
}

// EventRepoMatch is a subset of zoekt.FileMatch for our Event API.
type EventRepoMatch struct {
	// Type is always RepoMatchType. Included here for marshalling.
	Type MatchType `json:"type"`

	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
}

func (e *EventRepoMatch) eventMatch() {}

// EventSymbolMatch is EventFileMatch but with Symbols instead of LineMatches
type EventSymbolMatch struct {
	// Type is always SymbolMatchType. Included here for marshalling.
	Type MatchType `json:"type"`

	Path       string   `json:"path"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
	Commit     string   `json:"commit,omitempty"`

	Symbols []Symbol `json:"symbols"`
}

func (e *EventSymbolMatch) eventMatch() {}

type Symbol struct {
	URL           string `json:"url"`
	Name          string `json:"name"`
	ContainerName string `json:"containerName"`
	Kind          string `json:"kind"`
}

// EventCommitMatch is the generic results interface from GQL. There is a lot
// of potential data that may be useful here, and some thought needs to be put
// into what is actually useful in a commit result / or if we should have a
// "type" for that.
type EventCommitMatch struct {
	// Type is always CommitMatchType. Included here for marshalling.
	Type MatchType `json:"type"`

	Icon    string `json:"icon"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	Detail  string `json:"detail"`
	Content string `json:"content"`
	// [line, character, length]
	Ranges [][3]int32 `json:"ranges"`
}

func (e *EventCommitMatch) eventMatch() {}

// EventFilter is a suggestion for a search filter. Currently has a 1-1
// correspondance with the SearchFilter graphql type.
type EventFilter struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Count    int    `json:"count"`
	LimitHit bool   `json:"limitHit"`
	Kind     string `json:"kind"`
}

// EventAlert is GQL.SearchAlert. It replaces when sent to match existing
// behaviour.
type EventAlert struct {
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	ProposedQueries []ProposedQuery `json:"proposedQueries"`
}

// ProposedQuery is a suggested query to run when we emit an alert.
type ProposedQuery struct {
	Description string `json:"description,omitempty"`
	Query       string `json:"query"`
}

// EventError emulates a JavaScript error with a message property
// as is returned when the search encounters an error.
type EventError struct {
	Message string `json:"message"`
}

type MatchType int

const (
	ContentMatchType MatchType = iota
	RepoMatchType
	SymbolMatchType
	CommitMatchType
	PathMatchType
)

func (t MatchType) MarshalJSON() ([]byte, error) {
	switch t {
	case ContentMatchType:
		return []byte(`"content"`), nil
	case RepoMatchType:
		return []byte(`"repo"`), nil
	case SymbolMatchType:
		return []byte(`"symbol"`), nil
	case CommitMatchType:
		return []byte(`"commit"`), nil
	case PathMatchType:
		return []byte(`"path"`), nil
	default:
		return nil, errors.Newf("unknown MatchType: %d", t)
	}

}

func (t *MatchType) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte(`"content"`)) {
		*t = ContentMatchType
	} else if bytes.Equal(b, []byte(`"repo"`)) {
		*t = RepoMatchType
	} else if bytes.Equal(b, []byte(`"symbol"`)) {
		*t = SymbolMatchType
	} else if bytes.Equal(b, []byte(`"commit"`)) {
		*t = CommitMatchType
	} else if bytes.Equal(b, []byte(`"path"`)) {
		*t = PathMatchType
	} else {
		return errors.Newf("unknown MatchType: %s", b)
	}
	return nil
}
