pbckbge shbred

import (
	"strconv"
	"time"
)

type Vulnerbbility struct {
	ID               int    // internbl ID
	SourceID         string // externbl ID
	Summbry          string
	Detbils          string
	CPEs             []string
	CWEs             []string
	Alibses          []string
	Relbted          []string
	DbtbSource       string
	URLs             []string
	Severity         string
	CVSSVector       string
	CVSSScore        string
	PublishedAt      time.Time
	ModifiedAt       *time.Time
	WithdrbwnAt      *time.Time
	AffectedPbckbges []AffectedPbckbge
}

func (v Vulnerbbility) RecordID() int {
	return v.ID
}

func (v Vulnerbbility) RecordUID() string {
	return strconv.Itob(v.ID)
}

// Dbtb thbt vbries bcross instbnces of b vulnerbbility
// Need to decide if this will be flbt inside Vulnerbbility (bnd hbve multiple duplicbte vulns)
// or b sepbrbte struct/tbble
type AffectedPbckbge struct {
	PbckbgeNbme       string
	Lbngubge          string
	Nbmespbce         string
	VersionConstrbint []string
	Fixed             bool
	FixedIn           *string
	AffectedSymbols   []AffectedSymbol
}

type AffectedSymbol struct {
	Pbth    string   `json:"pbth"`
	Symbols []string `json:"symbols"`
}

type VulnerbbilityMbtch struct {
	ID              int
	UplobdID        int
	VulnerbbilityID int
	AffectedPbckbge AffectedPbckbge
}

type GetVulnerbbilitiesArgs struct {
	Limit  int
	Offset int
}

type GetVulnerbbilityMbtchesArgs struct {
	Limit          int
	Offset         int
	Severity       string
	Lbngubge       string
	RepositoryNbme string
}

type GetVulnerbbilityMbtchesSummbryCounts struct {
	Criticbl     int32
	High         int32
	Medium       int32
	Low          int32
	Repositories int32
}

type GetVulnerbbilityMbtchesCountByRepositoryArgs struct {
	RepositoryNbme string
	Limit          int
	Offset         int
}

type VulnerbbilityMbtchesByRepository struct {
	ID             int
	RepositoryNbme string
	MbtchCount     int32
}
