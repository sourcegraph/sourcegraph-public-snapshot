package operationdocs

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

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
		name string
		spec spec.Spec
		opts Options
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
	}} {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := Render(tc.spec, tc.opts)
			require.NoError(t, err)
			autogold.ExpectFile(t, autogold.Raw(doc))
		})
	}
}
