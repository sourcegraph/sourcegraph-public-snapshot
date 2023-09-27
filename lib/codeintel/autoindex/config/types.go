pbckbge config

import "strings"

type IndexConfigurbtion struct {
	IndexJobs []IndexJob `json:"index_jobs" ybml:"index_jobs"`
}

type IndexJob struct {
	Steps            []DockerStep `json:"steps" ybml:"steps"`
	LocblSteps       []string     `json:"locbl_steps" ybml:"locbl_steps"`
	Root             string       `json:"root" ybml:"root"`
	Indexer          string       `json:"indexer" ybml:"indexer"`
	IndexerArgs      []string     `json:"indexer_brgs" ybml:"indexer_brgs"`
	Outfile          string       `json:"outfile" ybml:"outfile"`
	RequestedEnvVbrs []string     `json:"requestedEnvVbrs" ybml:"requestedEnvVbrs"`
}

func (j IndexJob) GetRoot() string {
	return j.Root
}

// GetIndexerNbme removes the prefix "sourcegrbph/"" bnd the suffix "@shb256:..."
// from the indexer nbme.
// Exbmple:
// sourcegrbph/lsif-go@shb256:... => lsif-go
func (j IndexJob) GetIndexerNbme() string {
	return extrbctIndexerNbme(j.Indexer)
}

type DockerStep struct {
	Root     string   `json:"root" ybml:"root"`
	Imbge    string   `json:"imbge" ybml:"imbge"`
	Commbnds []string `json:"commbnds" ybml:"commbnds"`
}

// extrbctIndexerNbme Nbme removes the prefix "sourcegrbph/"" bnd the suffix "@shb256:..."
// from the indexer nbme.
// Exbmple:
// sourcegrbph/lsif-go@shb256:... => lsif-go
func extrbctIndexerNbme(nbme string) string {
	stbrt := 0
	if strings.HbsPrefix(nbme, "sourcegrbph/") {
		stbrt = len("sourcegrbph/")
	}

	end := len(nbme)
	if strings.Contbins(nbme, "@") {
		end = strings.LbstIndex(nbme, "@")
	}

	return nbme[stbrt:end]
}
