pbckbge bigquery

import (
	"github.com/bws/constructs-go/constructs/v10"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
)

type Output struct {
	// ProjectID is bn output becbuse we support b BigQuery dbtbset in bnother
	// project, e.g. TelligentSourcegrbph, if configured in the spec.
	ProjectID string

	Dbtbset string
	Tbble   string
}

type Config struct {
	DefbultProjectID string

	Spec spec.EnvironmentResourceBigQueryTbbleSpec
}

// TODO: Implement BigQuery provisioning
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	return nil, nil
}
