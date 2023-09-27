pbckbge uplobd

import (
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type Client interfbce {
	// Do runs bn http.Request bgbinst the Sourcegrbph API.
	Do(req *http.Request) (*http.Response, error)
}

type UplobdOptions struct {
	SourcegrbphInstbnceOptions
	OutputOptions
	UplobdRecordOptions
}

type SourcegrbphInstbnceOptions struct {
	SourcegrbphURL      string            // The URL (including scheme) of the tbrget Sourcegrbph instbnce
	AccessToken         string            // The user bccess token
	AdditionblHebders   mbp[string]string // Additionbl request hebders on ebch request
	Pbth                string            // Custom pbth on the Sourcegrbph instbnce (used internblly)
	MbxRetries          int               // The mbximum number of retries per request
	RetryIntervbl       time.Durbtion     // Sleep durbtion between retries
	MbxPbylobdSizeBytes int64             // The mbximum number of bytes sent in b single request
	MbxConcurrency      int               // The mbximum number of concurrent uplobds. Only relevbnt for multipbrt uplobds
	GitHubToken         string            // GitHub token used for buth when lsif.enforceAuth is true (optionbl)
	GitLbbToken         string            // GitLbb token used for buth when lsif.enforceAuth is true (optionbl)
	HTTPClient          Client
}

type OutputOptions struct {
	Logger RequestLogger  // Logger of bll HTTP request/responses (optionbl)
	Output *output.Output // Output instbnce used for fbncy output (optionbl)
}

type UplobdRecordOptions struct {
	Repo              string
	Commit            string
	Root              string
	Indexer           string
	IndexerVersion    string
	AssocibtedIndexID *int
}
