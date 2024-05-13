package operationdocs

import (
	"context"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/notionreposync/renderer/renderertest"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const (
	// Use a real service and environment to help validate links actually work
	// in our golden tests (TestRender).
	// https://handbook.sourcegraph.com/departments/engineering/managed-services/msp-testbed#test
	testServiceID            = "msp-testbed"
	testServiceEnvironment   = "test"
	testProjectID            = "msp-testbed-test-77589aae45d0"
	robertServiceEnvironment = "robert"
	robertProjectID          = "msp-testbed-robert-7be9"
)

func TestRender(t *testing.T) {
	for _, tc := range []struct {
		name   string
		spec   spec.Spec
		alerts map[string]terraform.AlertPolicy
		opts   Options
	}{{
		name: "basic",
		spec: spec.Spec{
			Service: spec.ServiceSpec{
				ID:          testServiceID,
				Description: "Test service for MSP",
				Name:        pointers.Ptr("MSP Testbed"),
			},
			Build: spec.BuildSpec{
				Image: "us.gcr.io/sourcegraph-dev/msp-example",
				Source: spec.BuildSourceSpec{
					Repo: "github.com/sourcegraph/sourcegraph",
					Dir:  "cmd/msp-example",
				},
			},
			Environments: []spec.EnvironmentSpec{{
				ID:        testServiceEnvironment,
				ProjectID: testProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Deploy: spec.EnvironmentDeploySpec{
					Type: "rollout",
				},
			}},
			Rollout: &spec.RolloutSpec{
				Stages: []spec.RolloutStageSpec{{EnvironmentID: testServiceEnvironment}},
			},
		},
		alerts: map[string]terraform.AlertPolicy{
			"monitoring-common-cpu": {
				DisplayName: "High Container CPU Utilization",
				Documentation: terraform.Documentation{
					Content: "High CPU Usage - it may be neccessary to reduce load or increase CPU allocation",
				},
				Severity: "WARNING",
			},
			"monitoring-common-memory": {
				DisplayName: "High Container Memory Utilization",
				Documentation: terraform.Documentation{
					Content: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
				},
				Severity: "WARNING",
			},
		},
		opts: Options{},
	}, {
		name: "resources",
		spec: spec.Spec{
			Service: spec.ServiceSpec{
				ID: "msp-testbed",

				Description: "Test service for MSP",
			},
			Build: spec.BuildSpec{
				Image: "us.gcr.io/sourcegraph-dev/msp-example",
				Source: spec.BuildSourceSpec{
					Repo: "github.com/sourcegraph/sourcegraph",
					Dir:  "cmd/msp-example",
				},
			},
			Environments: []spec.EnvironmentSpec{{
				ID:        testServiceEnvironment,
				ProjectID: testProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Resources: &spec.EnvironmentResourcesSpec{
					Redis: &spec.EnvironmentResourceRedisSpec{},
					PostgreSQL: &spec.EnvironmentResourcePostgreSQLSpec{
						Databases: []string{"foo", "bar"},
					},
					BigQueryDataset: &spec.EnvironmentResourceBigQueryDatasetSpec{
						Tables: []string{"bar", "baz"},
					},
				},
				Deploy: spec.EnvironmentDeploySpec{
					Type: "subscription",
				},
			}},
		},
		alerts: map[string]terraform.AlertPolicy{
			"monitoring-common-cpu": {
				DisplayName: "High Container CPU Utilization",
				Documentation: terraform.Documentation{
					Content: "High CPU Usage - it may be neccessary to reduce load or increase CPU allocation",
				},
				Severity: "WARNING",
			},
			"monitoring-common-memory": {
				DisplayName: "High Container Memory Utilization",
				Documentation: terraform.Documentation{
					Content: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
				},
				Severity: "WARNING",
			},
		},
		opts: Options{},
	}, {
		name: "with README",
		spec: spec.Spec{
			Service: spec.ServiceSpec{
				ID:          testServiceID,
				Description: "Test service for MSP",
				Name:        pointers.Ptr("MSP Testbed"),
			},
			Build: spec.BuildSpec{
				Image: "us.gcr.io/sourcegraph-dev/msp-example",
				Source: spec.BuildSourceSpec{
					Repo: "github.com/sourcegraph/sourcegraph",
					Dir:  "cmd/msp-example",
				},
			},
			Environments: []spec.EnvironmentSpec{{
				ID:        testServiceEnvironment,
				ProjectID: testProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Deploy: spec.EnvironmentDeploySpec{
					Type: "manual",
				},
			}},
			README: []byte(`This service does X, Y, Z. Refer to [here](sourcegraph.com) for more information.

## Additional operations

Some additional operations!`),
		},
		alerts: map[string]terraform.AlertPolicy{
			"monitoring-common-cpu": {
				DisplayName: "High Container CPU Utilization",
				Documentation: terraform.Documentation{
					Content: "High CPU Usage - it may be neccessary to reduce load or increase CPU allocation",
				},
				Severity: "WARNING",
			},
			"monitoring-common-memory": {
				DisplayName: "High Container Memory Utilization",
				Documentation: terraform.Documentation{
					Content: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
				},
				Severity: "WARNING",
			},
		},
		opts: Options{},
	}, {
		name: "multi env rollout",
		spec: spec.Spec{
			Service: spec.ServiceSpec{
				ID:          testServiceID,
				Description: "Test service for MSP",
				Name:        pointers.Ptr("MSP Testbed"),
			},
			Build: spec.BuildSpec{
				Image: "us.gcr.io/sourcegraph-dev/msp-example",
				Source: spec.BuildSourceSpec{
					Repo: "github.com/sourcegraph/sourcegraph",
					Dir:  "cmd/msp-example",
				},
			},
			Environments: []spec.EnvironmentSpec{{
				ID:        testServiceEnvironment,
				ProjectID: testProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Deploy: spec.EnvironmentDeploySpec{
					Type: "rollout",
				},
			}, {
				ID:        robertServiceEnvironment,
				ProjectID: robertProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Deploy: spec.EnvironmentDeploySpec{
					Type: "rollout",
				},
			}},
			Rollout: &spec.RolloutSpec{
				Stages: []spec.RolloutStageSpec{{EnvironmentID: testServiceEnvironment}, {EnvironmentID: robertServiceEnvironment}},
			},
		},
		alerts: map[string]terraform.AlertPolicy{
			"monitoring-common-cpu": {
				DisplayName: "High Container CPU Utilization",
				Documentation: terraform.Documentation{
					Content: "High CPU Usage - it may be neccessary to reduce load or increase CPU allocation",
				},
				Severity: "WARNING",
			},
			"monitoring-common-memory": {
				DisplayName: "High Container Memory Utilization",
				Documentation: terraform.Documentation{
					Content: "High Memory Usage - it may be neccessary to reduce load or increase memory allocation",
				},
				Severity: "WARNING",
			},
		},
		opts: Options{},
	}, {
		name: "with managed-services revision",
		spec: spec.Spec{
			Service: spec.ServiceSpec{
				ID:          testServiceID,
				Description: "Test service for MSP",
				Name:        pointers.Ptr("MSP Testbed"),
			},
			Build: spec.BuildSpec{
				Image: "us.gcr.io/sourcegraph-dev/msp-example",
				Source: spec.BuildSourceSpec{
					Repo: "github.com/sourcegraph/sourcegraph",
				},
			},
			Environments: []spec.EnvironmentSpec{{
				ID:        testServiceEnvironment,
				ProjectID: testProjectID,
				Category:  spec.EnvironmentCategoryTest,
				Deploy: spec.EnvironmentDeploySpec{
					Type: "rollout",
				},
			}},
			Rollout: &spec.RolloutSpec{
				Stages: []spec.RolloutStageSpec{{EnvironmentID: testServiceEnvironment}},
			},
		},
		alerts: map[string]terraform.AlertPolicy{},
		opts:   Options{ManagedServicesRevision: "a857d23cdc4184a045e4022285d38bed4acddac9"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := Render(tc.spec, tc.alerts, tc.opts)
			require.NoError(t, err)
			autogold.ExpectFile(t, autogold.Raw(doc))

			t.Run("renderable by Notion converter", func(t *testing.T) {
				blocks := renderertest.MockBlockUpdater{}
				assert.NoError(t, NewNotionConverter(context.Background(), &blocks).
					ProcessMarkdown([]byte(doc)))
				assert.NotEmpty(t, blocks.GetAddedBlocks())
			})
		})
	}
}
