package privateaccessperimeter

import (
	"fmt"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/accesscontextmanagerserviceperimeter"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/datagoogleproject"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	// Project is the service project.
	Project datagoogleproject.DataGoogleProject

	Service       spec.ServiceSpec
	EnvironmentID string

	Spec spec.EnvironmentPrivateAccessPerimeterSpec
}

type Output struct{}

// Manually retrieved from Console, no data source is available yet:
// https://github.com/hashicorp/terraform-provider-google/issues/8999
const orgAccessPolicy = "267168805930"

// New sets up GCP VPC Service Controls Perimeter around this project's Cloud
// Run APIs, allowlisting ingress based on the provided spec.
//
// This should only be created once, hence why it does not have accept
// a resourceid.ID
func New(scope constructs.Construct, config Config) (*Output, error) {
	id := resourceid.New("privateaccessperimeter") // top-level because this resource is a singleton

	// For each ingress policy, we allow ingress to everything in the perimeter.
	ingressToSpec := accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressTo{
		Resources: pointers.Ptr([]*string{pointers.Ptr("*")}),
		Operations: accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressToOperations{
			ServiceName: pointers.Ptr("*"),
		},
	}

	// Set up policy allowing allowlisted project IDs
	allowedProjectsIngressPolicy := accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPolicies{
		IngressFrom: &accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFrom{
			Sources: newAllowlistedIngressProjectSources(
				scope,
				id.Group("allowlisted_projects"),
				config.Spec.AllowlistedProjects),
			IdentityType: pointers.Ptr("ANY_IDENTITY"),
		},
		IngressTo: &ingressToSpec,
	}

	// Set up policy allowing allowlisted identities.
	allowedIdentities := append([]string{
		"serviceAccount:terraform-cloud@sourcegraph-ci.iam.gserviceaccount.com",
		"TODO",
	}, config.Spec.AllowlistedIdentities...)
	// TODO
	allowedIdentitiesIngressPolicy := accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPolicies{
		IngressFrom: &accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFrom{
			Identities: pointers.Ptr(pointers.Slice(allowedIdentities)),
			Sources: []accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFromSources{
				{AccessLevel: pointers.Ptr("*")},
			},
		},
		IngressTo: &ingressToSpec,
	}

	// Create our VPC SC perimeter.
	_ = accesscontextmanagerserviceperimeter.NewAccessContextManagerServicePerimeter(
		scope,
		id.TerraformID("vpc_sc_perimeter"),
		&accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterConfig{
			Parent: pointers.Stringf("accessPolicies/%s", orgAccessPolicy),
			Name: pointers.Stringf("accessPolicies/%s/servicePerimeters/%s_%s",
				orgAccessPolicy, strings.ReplaceAll(config.Service.ID, "-", "_"), strings.ReplaceAll(config.EnvironmentID, "-", "_")),

			Title: pointers.Stringf("Cloud Run Internal Access for %s - %s",
				config.Service.GetName(), config.EnvironmentID),

			PerimeterType: pointers.Ptr("PERIMETER_TYPE_REGULAR"),

			Status: &accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterStatus{
				// Only manage Cloud Run API access, which includes Cloud Run
				// internal URL access.
				RestrictedServices: pointers.Ptr(pointers.Slice([]string{
					"run.googleapis.com",
				})),

				// Place only the target service's project in perimeter, so we can allowlist
				// cross-project internal traffic to this perimeter
				Resources: pointers.Ptr(pointers.Slice([]string{
					fmt.Sprintf("projects/%v", config.Project.Number()),
				})),

				// Assign the ingress policies we have created.
				IngressPolicies: []accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPolicies{
					allowedProjectsIngressPolicy,
					allowedIdentitiesIngressPolicy,
				},
			},
		})

	return &Output{}, nil
}

func newAllowlistedIngressProjectSources(scope constructs.Construct, id resourceid.ID, allowlistedProjects []string) []accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFromSources {
	var allowedIngressProjects []accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFromSources
	for idx, allowedProject := range allowlistedProjects {
		id := id.Group("%d", idx)

		// Allowlist only accepts 'projects/{numeric_id}' format - we need to
		// fetch this from a data source.
		project := datagoogleproject.NewDataGoogleProject(scope, id.TerraformID("project"), &datagoogleproject.DataGoogleProjectConfig{
			ProjectId: &allowedProject,
		})
		allowedIngressProjects = append(allowedIngressProjects,
			accesscontextmanagerserviceperimeter.AccessContextManagerServicePerimeterSpecIngressPoliciesIngressFromSources{
				Resource: pointers.Stringf("projects/%v", project.Number()),
			})
	}
	return allowedIngressProjects
}
