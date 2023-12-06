package bigquery

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerydataset"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerydatasetiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/bigquerytable"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/serviceaccount"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	// ProjectID is an output because we support a BigQuery dataset in another
	// project, e.g. TelligentSourcegraph, if configured in the spec.
	ProjectID string
	// DatasetID is the BigQuery dataset the service should target.
	DatasetID string
	// Tables are the tables that must be created before services can connect
	// to one of their configured tables.
	Tables []cdktf.ITerraformDependable
}

type Config struct {
	DefaultProjectID string
	ServiceID        string

	WorkloadServiceAccount *serviceaccount.Output

	Spec spec.EnvironmentResourceBigQueryDatasetSpec
}

// New creates a BigQuery dataset and all configured tables.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	var (
		datasetID = pointers.Deref(config.Spec.DatasetID, config.ServiceID)
		projectID = pointers.Deref(config.Spec.ProjectID, config.DefaultProjectID)
		location  = pointers.Deref(config.Spec.Location, "US")
		labels    = map[string]*string{
			"service": &config.ServiceID,
		}
	)

	dataset := bigquerydataset.NewBigqueryDataset(scope, id.TerraformID("dataset"), &bigquerydataset.BigqueryDatasetConfig{
		Project:  &projectID,
		Location: &location,

		DatasetId: &datasetID,
		Labels:    &labels,
	})

	// Grant the workload SA editor access to the entire dataset.
	editorRole := bigquerydatasetiammember.NewBigqueryDatasetIamMember(scope, id.TerraformID("workload_dataset_editor"), &bigquerydatasetiammember.BigqueryDatasetIamMemberConfig{
		Project:   &projectID,
		DatasetId: dataset.DatasetId(),

		Role:   pointers.Ptr("roles/bigquery.dataEditor"),
		Member: &config.WorkloadServiceAccount.Member,
	})

	// Provision each requested table.
	var tables []cdktf.ITerraformDependable
	for _, tableID := range config.Spec.Tables {
		tables = append(tables,
			bigquerytable.NewBigqueryTable(scope, id.Group(tableID).TerraformID("table"),
				&bigquerytable.BigqueryTableConfig{
					Project:   &projectID,
					DatasetId: dataset.DatasetId(),

					TableId: &tableID,
					Schema:  pointers.Ptr(string(config.Spec.GetSchema(tableID))),
					Labels:  &labels,

					// In order to write to the table, the workload SA must have editor
					// access, so we make table depend on the role grant.
					DependsOn: &[]cdktf.ITerraformDependable{editorRole},
				}))
	}

	return &Output{
		ProjectID: projectID,
		DatasetID: *dataset.DatasetId(),
		Tables:    tables,
	}, nil
}
