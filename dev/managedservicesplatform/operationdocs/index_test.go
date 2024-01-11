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
				ID:     "service-1",
				Name:   pointers.Ptr("Service 1"),
				Owners: []string{"team-a"},
			},
		},
		{
			Service: spec.ServiceSpec{
				ID:     "service-2",
				Name:   pointers.Ptr("Service 2"),
				Owners: []string{"team-a"},
			},
		},
		{
			Service: spec.ServiceSpec{
				ID:     "service-3",
				Name:   pointers.Ptr("Service 3"),
				Owners: []string{"team-b"},
			},
		},
		{
			Service: spec.ServiceSpec{
				ID:     "service-4",
				Name:   pointers.Ptr("Service 4"),
				Owners: []string{"team-a", "team-b"},
			},
		},
	}

	doc := RenderIndexPage(services, Options{})
	autogold.ExpectFile(t, autogold.Raw(doc))
}
