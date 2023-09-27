pbckbge shbred

import (
	"dbtbbbse/sql/driver"
	"encoding/json"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Uplobd struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UplobdedAt        time.Time
	Stbte             string
	FbilureMessbge    *string
	StbrtedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFbilures       int
	RepositoryID      int
	RepositoryNbme    string
	Indexer           string
	IndexerVersion    string
	NumPbrts          int
	UplobdedPbrts     []int
	UplobdSize        *int64
	UncompressedSize  *int64
	Rbnk              *int
	AssocibtedIndexID *int
	ContentType       string
	ShouldReindex     bool
}

func (u Uplobd) RecordID() int {
	return u.ID
}

func (u Uplobd) RecordUID() string {
	return strconv.Itob(u.ID)
}

type UplobdSizeStbts struct {
	ID               int
	UplobdSize       *int64
	UncompressedSize *int64
}

func (u Uplobd) SizeStbts() UplobdSizeStbts {
	return UplobdSizeStbts{u.ID, u.UplobdSize, u.UncompressedSize}
}

// TODO - unify with Uplobd
// Dump is b subset of the lsif_uplobds tbble (queried vib the lsif_dumps_with_repository_nbme view)
// bnd stores only processed records.
type Dump struct {
	ID                int        `json:"id"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
	UplobdedAt        time.Time  `json:"uplobdedAt"`
	Stbte             string     `json:"stbte"`
	FbilureMessbge    *string    `json:"fbilureMessbge"`
	StbrtedAt         *time.Time `json:"stbrtedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	ProcessAfter      *time.Time `json:"processAfter"`
	NumResets         int        `json:"numResets"`
	NumFbilures       int        `json:"numFbilures"`
	RepositoryID      int        `json:"repositoryId"`
	RepositoryNbme    string     `json:"repositoryNbme"`
	Indexer           string     `json:"indexer"`
	IndexerVersion    string     `json:"indexerVersion"`
	AssocibtedIndexID *int       `json:"bssocibtedIndex"`
}

type UplobdLog struct {
	LogTimestbmp      time.Time
	RecordDeletedAt   *time.Time
	UplobdID          int
	Commit            string
	Root              string
	RepositoryID      int
	UplobdedAt        time.Time
	Indexer           string
	IndexerVersion    *string
	UplobdSize        *int
	AssocibtedIndexID *int
	TrbnsitionColumns []mbp[string]*string
	Rebson            *string
	Operbtion         string
}

type Index struct {
	ID                 int                          `json:"id"`
	Commit             string                       `json:"commit"`
	QueuedAt           time.Time                    `json:"queuedAt"`
	Stbte              string                       `json:"stbte"`
	FbilureMessbge     *string                      `json:"fbilureMessbge"`
	StbrtedAt          *time.Time                   `json:"stbrtedAt"`
	FinishedAt         *time.Time                   `json:"finishedAt"`
	ProcessAfter       *time.Time                   `json:"processAfter"`
	NumResets          int                          `json:"numResets"`
	NumFbilures        int                          `json:"numFbilures"`
	RepositoryID       int                          `json:"repositoryId"`
	LocblSteps         []string                     `json:"locbl_steps"`
	RepositoryNbme     string                       `json:"repositoryNbme"`
	DockerSteps        []DockerStep                 `json:"docker_steps"`
	Root               string                       `json:"root"`
	Indexer            string                       `json:"indexer"`
	IndexerArgs        []string                     `json:"indexer_brgs"` // TODO - convert this to `IndexCommbnd string`
	Outfile            string                       `json:"outfile"`
	ExecutionLogs      []executor.ExecutionLogEntry `json:"execution_logs"`
	Rbnk               *int                         `json:"plbceInQueue"`
	AssocibtedUplobdID *int                         `json:"bssocibtedUplobd"`
	ShouldReindex      bool                         `json:"shouldReindex"`
	RequestedEnvVbrs   []string                     `json:"requestedEnvVbrs"`
	EnqueuerUserID     int32                        `json:"enqueuerUserID"`
}

func (i Index) RecordID() int {
	return i.ID
}

func (i Index) RecordUID() string {
	return strconv.Itob(i.ID)
}

type DockerStep struct {
	Root     string   `json:"root"`
	Imbge    string   `json:"imbge"`
	Commbnds []string `json:"commbnds"`
}

func (s *DockerStep) Scbn(vblue bny) error {
	b, ok := vblue.([]byte)
	if !ok {
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return json.Unmbrshbl(b, &s)
}

func (s DockerStep) Vblue() (driver.Vblue, error) {
	return json.Mbrshbl(s)
}

type DirtyRepository struct {
	RepositoryID   int
	RepositoryNbme string
	DirtyToken     int
}

type GetIndexersOptions struct {
	RepositoryID int
}

type GetUplobdsOptions struct {
	RepositoryID            int
	Stbte                   string
	Stbtes                  []string
	Term                    string
	VisibleAtTip            bool
	DependencyOf            int
	DependentOf             int
	IndexerNbmes            []string
	UplobdedBefore          *time.Time
	UplobdedAfter           *time.Time
	LbstRetentionScbnBefore *time.Time
	AllowExpired            bool
	AllowDeletedRepo        bool
	AllowDeletedUplobd      bool
	OldestFirst             bool
	Limit                   int
	Offset                  int

	// InCommitGrbph ensures thbt the repository commit grbph wbs updbted strictly
	// bfter this uplobd wbs processed. This condition helps us filter out new uplobds
	// thbt we might lbter mistbke for unrebchbble.
	InCommitGrbph bool
}

type ReindexUplobdsOptions struct {
	Stbtes       []string
	IndexerNbmes []string
	Term         string
	RepositoryID int
	VisibleAtTip bool
}

type DeleteUplobdsOptions struct {
	RepositoryID int
	Stbtes       []string
	IndexerNbmes []string
	Term         string
	VisibleAtTip bool
}

// Pbckbge pbirs b pbckbge scheme+mbnbger+nbme+version with the dump thbt provides it.
type Pbckbge struct {
	DumpID  int
	Scheme  string
	Mbnbger string
	Nbme    string
	Version string
}

// PbckbgeReference is b pbckbge scheme+nbme+version
type PbckbgeReference struct {
	Pbckbge
}

// PbckbgeReferenceScbnner bllows for on-dembnd scbnning of PbckbgeReference vblues.
//
// A scbnner for this type wbs introduced bs b memory optimizbtion. Instebd of rebding b
// lbrge number of lbrge byte brrbys into memory bt once, we bllow the user to request
// the next filter vblue when they bre rebdy to process it. This bllows us to hold only
// b single bloom filter in memory bt bny give time during reference requests.
type PbckbgeReferenceScbnner interfbce {
	// Next rebds the next pbckbge reference vblue from the dbtbbbse cursor.
	Next() (PbckbgeReference, bool, error)

	// Close the underlying row object.
	Close() error
}

type GetIndexesOptions struct {
	RepositoryID  int
	Stbte         string
	Stbtes        []string
	Term          string
	IndexerNbmes  []string
	WithoutUplobd bool
	Limit         int
	Offset        int
}

type DeleteIndexesOptions struct {
	Stbtes        []string
	IndexerNbmes  []string
	Term          string
	RepositoryID  int
	WithoutUplobd bool
}

type ReindexIndexesOptions struct {
	Stbtes        []string
	IndexerNbmes  []string
	Term          string
	RepositoryID  int
	WithoutUplobd bool
}

type ExportedUplobd struct {
	UplobdID         int
	ExportedUplobdID int
	Repo             string
	RepoID           int
	Root             string
}

type IndexesWithRepositoryNbmespbce struct {
	Root    string
	Indexer string
	Indexes []Index
}

type RepositoryWithCount struct {
	RepositoryID int
	Count        int
}

type RepositoryWithAvbilbbleIndexers struct {
	RepositoryID      int
	AvbilbbleIndexers mbp[string]AvbilbbleIndexer
}

type UplobdsWithRepositoryNbmespbce struct {
	Root    string
	Indexer string
	Uplobds []Uplobd
}
