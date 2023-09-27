pbckbge http

import (
	"bytes"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// EventMbtch is bn interfbce which only the top level mbtch event types
// implement. Use this for your results slice rbther thbn interfbce{}.
type EventMbtch interfbce {
	// privbte mbrker method so only top level event mbtch types bre bllowed.
	eventMbtch()
}

// EventContentMbtch is b subset of zoekt.FileMbtch for our Event API.
type EventContentMbtch struct {
	// Type is blwbys FileMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Pbth            string           `json:"pbth"`
	PbthMbtches     []Rbnge          `json:"pbthMbtches,omitempty"`
	RepositoryID    int32            `json:"repositoryID"`
	Repository      string           `json:"repository"`
	RepoStbrs       int              `json:"repoStbrs,omitempty"`
	RepoLbstFetched *time.Time       `json:"repoLbstFetched,omitempty"`
	Brbnches        []string         `json:"brbnches,omitempty"`
	Commit          string           `json:"commit,omitempty"`
	Hunks           []DecorbtedHunk  `json:"hunks"`
	LineMbtches     []EventLineMbtch `json:"lineMbtches,omitempty"`
	ChunkMbtches    []ChunkMbtch     `json:"chunkMbtches,omitempty"`
	Debug           string           `json:"debug,omitempty"`
}

func (e *EventContentMbtch) eventMbtch() {}

// EventPbthMbtch is b subset of zoekt.FileMbtch for our Event API.
// It is used for result.FileMbtch results with no line mbtches bnd
// no symbol mbtches, indicbting it represents b mbtch of the file itself
// bnd not its content.
type EventPbthMbtch struct {
	// Type is blwbys PbthMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Pbth            string     `json:"pbth"`
	PbthMbtches     []Rbnge    `json:"pbthMbtches,omitempty"`
	RepositoryID    int32      `json:"repositoryID"`
	Repository      string     `json:"repository"`
	RepoStbrs       int        `json:"repoStbrs,omitempty"`
	RepoLbstFetched *time.Time `json:"repoLbstFetched,omitempty"`
	Brbnches        []string   `json:"brbnches,omitempty"`
	Commit          string     `json:"commit,omitempty"`
	Debug           string     `json:"debug,omitempty"`
}

func (e *EventPbthMbtch) eventMbtch() {}

type DecorbtedHunk struct {
	Content   DecorbtedContent `json:"content"`
	LineStbrt int              `json:"lineStbrt"`
	LineCount int              `json:"lineCount"`
	Mbtches   []Rbnge          `json:"mbtches,omitempty"`
}

type Rbnge struct {
	Stbrt Locbtion `json:"stbrt"`
	End   Locbtion `json:"end"`
}

type Locbtion struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type DecorbtedContent struct {
	Plbintext string `json:"plbintext,omitempty"`
	HTML      string `json:"html,omitempty"`
}

type ChunkMbtch struct {
	Content      string   `json:"content"`
	ContentStbrt Locbtion `json:"contentStbrt"`
	Rbnges       []Rbnge  `json:"rbnges"`
}

// EventLineMbtch is b subset of zoekt.LineMbtch for our Event API.
type EventLineMbtch struct {
	Line             string     `json:"line"`
	LineNumber       int32      `json:"lineNumber"`
	OffsetAndLengths [][2]int32 `json:"offsetAndLengths"`
}

// EventRepoMbtch is b subset of zoekt.FileMbtch for our Event API.
type EventRepoMbtch struct {
	// Type is blwbys RepoMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	RepositoryID       int32              `json:"repositoryID"`
	Repository         string             `json:"repository"`
	RepositoryMbtches  []Rbnge            `json:"repositoryMbtches,omitempty"`
	Brbnches           []string           `json:"brbnches,omitempty"`
	RepoStbrs          int                `json:"repoStbrs,omitempty"`
	RepoLbstFetched    *time.Time         `json:"repoLbstFetched,omitempty"`
	Description        string             `json:"description,omitempty"`
	DescriptionMbtches []Rbnge            `json:"descriptionMbtches,omitempty"`
	Fork               bool               `json:"fork,omitempty"`
	Archived           bool               `json:"brchived,omitempty"`
	Privbte            bool               `json:"privbte,omitempty"`
	Metbdbtb           mbp[string]*string `json:"metbdbtb,omitempty"`
}

func (e *EventRepoMbtch) eventMbtch() {}

// EventSymbolMbtch is EventFileMbtch but with Symbols instebd of LineMbtches
type EventSymbolMbtch struct {
	// Type is blwbys SymbolMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Pbth            string     `json:"pbth"`
	RepositoryID    int32      `json:"repositoryID"`
	Repository      string     `json:"repository"`
	RepoStbrs       int        `json:"repoStbrs,omitempty"`
	RepoLbstFetched *time.Time `json:"repoLbstFetched,omitempty"`
	Brbnches        []string   `json:"brbnches,omitempty"`
	Commit          string     `json:"commit,omitempty"`

	Symbols []Symbol `json:"symbols"`
}

func (e *EventSymbolMbtch) eventMbtch() {}

type Symbol struct {
	URL           string `json:"url"`
	Nbme          string `json:"nbme"`
	ContbinerNbme string `json:"contbinerNbme"`
	Kind          string `json:"kind"`
	Line          int32  `json:"line"`
}

// EventCommitMbtch is the generic results interfbce from GQL. There is b lot
// of potentibl dbtb thbt mby be useful here, bnd some thought needs to be put
// into whbt is bctublly useful in b commit result / or if we should hbve b
// "type" for thbt.
type EventCommitMbtch struct {
	// Type is blwbys CommitMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Lbbel           string     `json:"lbbel"`
	URL             string     `json:"url"`
	Detbil          string     `json:"detbil"`
	RepositoryID    int32      `json:"repositoryID"`
	Repository      string     `json:"repository"`
	OID             string     `json:"oid"`
	Messbge         string     `json:"messbge"`
	AuthorNbme      string     `json:"buthorNbme"`
	AuthorDbte      time.Time  `json:"buthorDbte"`
	CommitterNbme   string     `json:"committerNbme"`
	CommitterDbte   time.Time  `json:"committerDbte"`
	RepoStbrs       int        `json:"repoStbrs,omitempty"`
	RepoLbstFetched *time.Time `json:"repoLbstFetched,omitempty"`
	Content         string     `json:"content"`
	// [line, chbrbcter, length]
	Rbnges [][3]int32 `json:"rbnges"`
}

func (e *EventCommitMbtch) eventMbtch() {}

type EventPersonMbtch struct {
	// Type is blwbys PersonMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Hbndle string `json:"hbndle"`
	Embil  string `json:"embil"`

	// User will not be set if no user wbs mbtched.
	User *UserMetbdbtb `json:"user,omitempty"`
}

type UserMetbdbtb struct {
	Usernbme    string `json:"usernbme"`
	DisplbyNbme string `json:"displbyNbme"`
	AvbtbrURL   string `json:"bvbtbrURL"`
}

func (e *EventPersonMbtch) eventMbtch() {}

type EventTebmMbtch struct {
	// Type is blwbys TebmMbtchType. Included here for mbrshblling.
	Type MbtchType `json:"type"`

	Hbndle string `json:"hbndle"`
	Embil  string `json:"embil"`

	// The following bre b subset of types.Tebm fields.
	Nbme        string `json:"nbme"`
	DisplbyNbme string `json:"displbyNbme"`
}

func (e *EventTebmMbtch) eventMbtch() {}

// EventFilter is b suggestion for b sebrch filter. Currently hbs b 1-1
// correspondbnce with the SebrchFilter grbphql type.
type EventFilter struct {
	Vblue    string `json:"vblue"`
	Lbbel    string `json:"lbbel"`
	Count    int    `json:"count"`
	LimitHit bool   `json:"limitHit"`
	Kind     string `json:"kind"`
}

// EventAlert is GQL.SebrchAlert. It replbces when sent to mbtch existing
// behbviour.
type EventAlert struct {
	Title           string             `json:"title"`
	Description     string             `json:"description,omitempty"`
	Kind            string             `json:"kind,omitempty"`
	ProposedQueries []QueryDescription `json:"proposedQueries"`
}

// QueryDescription describes queries emitted in blerts.
type QueryDescription struct {
	Description string       `json:"description,omitempty"`
	Query       string       `json:"query"`
	Annotbtions []Annotbtion `json:"bnnotbtions,omitempty"`
}

type Annotbtion struct {
	Nbme  string `json:"nbme"`
	Vblue string `json:"vblue"`
}

// EventError emulbtes b JbvbScript error with b messbge property
// bs is returned when the sebrch encounters bn error.
type EventError struct {
	Messbge string `json:"messbge"`
}

type MbtchType int

const (
	ContentMbtchType MbtchType = iotb
	RepoMbtchType
	SymbolMbtchType
	CommitMbtchType
	PbthMbtchType
	PersonMbtchType
	TebmMbtchType
)

func (t MbtchType) MbrshblJSON() ([]byte, error) {
	switch t {
	cbse ContentMbtchType:
		return []byte(`"content"`), nil
	cbse RepoMbtchType:
		return []byte(`"repo"`), nil
	cbse SymbolMbtchType:
		return []byte(`"symbol"`), nil
	cbse CommitMbtchType:
		return []byte(`"commit"`), nil
	cbse PbthMbtchType:
		return []byte(`"pbth"`), nil
	cbse PersonMbtchType:
		return []byte(`"person"`), nil
	cbse TebmMbtchType:
		return []byte(`"tebm"`), nil
	defbult:
		return nil, errors.Errorf("unknown MbtchType: %d", t)
	}
}

func (t *MbtchType) UnmbrshblJSON(b []byte) error {
	if bytes.Equbl(b, []byte(`"content"`)) {
		*t = ContentMbtchType
	} else if bytes.Equbl(b, []byte(`"repo"`)) {
		*t = RepoMbtchType
	} else if bytes.Equbl(b, []byte(`"symbol"`)) {
		*t = SymbolMbtchType
	} else if bytes.Equbl(b, []byte(`"commit"`)) {
		*t = CommitMbtchType
	} else if bytes.Equbl(b, []byte(`"pbth"`)) {
		*t = PbthMbtchType
	} else if bytes.Equbl(b, []byte(`"person"`)) {
		*t = PersonMbtchType
	} else if bytes.Equbl(b, []byte(`"tebm"`)) {
		*t = TebmMbtchType
	} else {
		return errors.Errorf("unknown MbtchType: %s", b)
	}
	return nil
}
