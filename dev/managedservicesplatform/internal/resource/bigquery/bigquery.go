package bigquery

import (
	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

type Output struct {
	// ProjectID is an output because we support a BigQuery dataset in another
	// project, e.g. TelligentSourcegraph, if configured in the spec.
	ProjectID string

	Dataset string
	Table   string
}

type Config struct {
	DefaultProjectID string

	Spec spec.EnvironmentResourceBigQueryTableSpec
}

// TODO: Implement BigQuery provisioning
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	return nil, nil
}
