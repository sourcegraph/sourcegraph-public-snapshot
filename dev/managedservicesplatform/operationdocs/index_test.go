package operationdocs

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestRenderIndexPage(t *testing.T) {
	services := []*spec.Spec{
		{
			Service: spec.ServiceSpec{
				ID:          "service-1",
				Description: "Test service for MSP",
				Name:        pointers.Ptr("Service 1"),
				Owners:      []string{"team-a"},
			},
			Environments: []spec.EnvironmentSpec{{
				ID: "dev",
			}},
		},
		{
			Service: spec.ServiceSpec{
				ID:          "service-2",
				Description: "Another test service for MSP",
				Name:        pointers.Ptr("Service 2"),
				Owners:      []string{"team-a"},
			},
			Environments: []spec.EnvironmentSpec{{
				ID: "dev",
			}, {
				ID: "prod",
			}},
		},
		{
			Service: spec.ServiceSpec{
				ID:          "service-3",
				Description: "Yet another test service for MSP",
				Name:        pointers.Ptr("Service 3"),
				Owners:      []string{"team-b"},
			},
			Environments: []spec.EnvironmentSpec{{
				ID: "dev",
			}},
		},
		{
			Service: spec.ServiceSpec{
				ID:          "service-4",
				Description: "ALSO a test service for MSP",
				Name:        pointers.Ptr("Service 4"),
				Owners:      []string{"team-a", "team-b"},
			},
			Environments: []spec.EnvironmentSpec{{
				ID: "dev",
			}, {
				ID: "prod",
			}},
		},
	}

	doc := RenderIndexPage(services, Options{})
	autogold.ExpectFile(t, autogold.Raw(doc))
}
