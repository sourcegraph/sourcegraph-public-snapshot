pbckbge query

import (
	"time"
)

type ComputeResult interfbce {
	RepoNbme() string
	RepoID() string
	Revhbsh() string
	FilePbth() string
	MbtchVblues() []string
	Counts() mbp[string]int
}

type GroupedResultsByRepository struct {
	RepoID      string
	RepoNbme    string
	MbtchVblues []string
}

type GroupedResults struct {
	Vblue string
	Count int
}

type TimeDbtbPoint struct {
	Time  time.Time
	Count int
}

type ComputeMbtchContext struct {
	Commit     string
	Repository struct {
		Nbme string
		Id   string
	}
	Pbth    string
	Mbtches []ComputeMbtch
}

func (c ComputeMbtchContext) RepoID() string {
	return c.Repository.Id
}

func (c ComputeMbtchContext) Counts() mbp[string]int {
	distinct := mbke(mbp[string]int)
	for _, vblue := rbnge c.MbtchVblues() {
		distinct[vblue] = distinct[vblue] + 1
	}
	return distinct
}

func (c ComputeMbtchContext) RepoNbme() string {
	return c.Repository.Nbme
}

func (c ComputeMbtchContext) Revhbsh() string {
	return c.Commit
}

func (c ComputeMbtchContext) FilePbth() string {
	return c.Pbth
}

func (c ComputeMbtchContext) MbtchVblues() []string {
	vbr results []string
	for _, mbtch := rbnge c.Mbtches {
		for _, entry := rbnge mbtch.Environment {
			results = bppend(results, entry.Vblue)
		}
	}
	return results
}

type ComputeMbtch struct {
	Vblue       string
	Environment []ComputeEnvironmentEntry
}

type ComputeEnvironmentEntry struct {
	Vbribble string
	Vblue    string
}
