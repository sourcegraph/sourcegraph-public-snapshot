package bigquery

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
)

type Output struct {
	// ProjectID is an output because we support a BigQuery dataset in another
	// project, e.g. TelligentSourcegraph, if configured in the spec.
	ProjectID string

	Dataset string
	Table   string
}

type Config struct {
	DefaultProject project.Project

	Spec spec.EnvironmentResourceBigQueryTableSpec
}

func New(scope constructs.Construct, name string, config Config) (*Output, error) {
	return nil, nil // TODO
}
