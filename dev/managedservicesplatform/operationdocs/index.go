package operationdocs

import (
	"path/filepath"
	"slices"
	"sort"

	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

// HandbookDirectory designates where in sourcegraph/handbook operation docs should go.
//
// Place under top-level 'engineering/managed-services' since it's too much work
// to find the appropriate team-specific content tree, and they change frequently.
const HandbookDirectory = "content/departments/engineering/managed-services"

// ServiceHandbookPath designates where in sourcegraph/handbook the contents
// of operationdocs.Render should go.
func ServiceHandbookPath(service string) string {
	return filepath.Join(HandbookDirectory, service+".md")
}

// ServiceHandbookPath designates where in sourcegraph/handbook the contents
// of operationdocs.RenderIndexPage should go.
func IndexPathHandbookPath() string {
	return filepath.Join(HandbookDirectory, "index.md")
}

// Relative paths to pages we want to link to in handbook mode, expecting that
// the index page and service-specific pages be housed in HandbookPath.
const (
	relativePathToMSPPage          = "../teams/core-services/managed-services/platform.md"
	relativePathToCoreServicesPage = "../teams/core-services/index.md"
)

// RenderIndexPage renders an index page for use at HandbookPath, assuming that
// operationdocs.Render contents are stored
func RenderIndexPage(services []*spec.Spec, opts Options) string {
	md := markdown.NewBuilder()

	md.Headingf(1, "Managed Services infrastructure")

	opts.AddDocumentComment(md)

	generalGuidanceLink, generalGuidance := markdown.HeadingLinkf("General guidance")
	md.Paragraphf(`These pages contain generated operational guidance for the infrastructure of %s services.
This includes information about each service, configured environments, Entitle requests, common tasks, monitoring, etc.
In addition to service-specific guidance, %s is also available.`,
		markdown.Link("Managed Services Platform (MSP)", relativePathToMSPPage),
		generalGuidanceLink)

	md.Paragraphf(`MSP is owned by %s, but individual teams are responsible for the services they operate on the platform.`,
		markdown.Link("Core Services", relativePathToCoreServicesPage))

	md.Paragraphf("Services are defined in %s, though service source code may live elsewhere.",
		markdown.Link(markdown.Code("sourcegraph/managed-services"), "https://github.com/sourcegraph/managed-services"))

	md.Admonitionf(markdown.AdmonitionNote,
		"This page may be out of date if a service or environment was recently added or updated - reach out to #discuss-core-services for help updating these pages, or use %s to view the generated documentation in your terminal.",
		markdown.Code("sg msp operations $SERVICE_ID"))

	owners, byOwner := collectByOwner(services)
	for _, o := range owners {
		md.Headingf(2, o)
		md.Paragraphf("Managed Services Platform services owned by %s:", markdown.Code(o))
		md.List(mapTo(byOwner[o], func(s *spec.Spec) string {
			// TODO: See Service.Description docstring
			// title := fmt.Sprintf("%s - %s", s.Service.GetName(), s.Service.Description)
			return markdown.Linkf(s.Service.GetName(), "./%s.md", s.Service.ID)
		}))
	}

	md.Headingf(2, generalGuidance)

	md.Headingf(3, "Infrastructure access")
	md.Paragraphf(`For MSP service environments other than %[1]s, access needs to be requested through Entitle.
Test environments are placed in the "Engineering Projects" GCP folder, which should have access granted to engineers by default.

Entitle access to a production MSP project is generally provisioned through the %[2]s and %[3]s custom GCP roles, which provide read-only and editing access respectively.
Convenience links for requesting these roles are available in the per-service operation pages above, based on each environment.

You can also choose to request access to an individual project in Entitle by following these steps:

- Go to [app.entitle.io/request](https://app.entitle.io/request) and select **Specific Permission**
- Fill out the following:
  - Integration: **GCP Production Projects**
  - Resource types: **Project**
  - Resource: name of MSP project you are interested in
  - Role: %[2]s (or %[3]s if you need additional privileges - use with care!)
  - Duration: choose your own adventure!

The custom roles used for MSP infrastructure access are [configured in %[5]s](https://github.com/sourcegraph/infrastructure/blob/main/gcp/custom-roles/msp.tf).`,
		markdown.Code("category: test"),                // %[1]s
		markdown.Code("mspServiceReader"),              // %[2]s
		markdown.Code("mspServiceEditor"),              // %[3]s
		markdown.Code("gcp/org/customer-roles/msp.tf"), // %[4]s
		markdown.Code("sourcegraph/infrastructure"),    // %[5]s
	)

	md.Headingf(3, "Terraform Cloud access")
	md.Paragraphf(`Terraform Cloud (TFC) workspaces for MSP [can be found using the %s workspace tag](https://app.terraform.io/app/sourcegraph/workspaces?tag=msp).

To gain access to MSP project TFC workspaces, [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) and _then_ [request membership to the %s TFC team via Entitle](https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjM2MDAiLCJqdXN0aWZpY2F0aW9uIjoiRU5URVIgSlVTVElGSUNBVElPTiBIRVJFIiwicm9sZUlkcyI6W3siaWQiOiJiMzg3MzJjYy04OTUyLTQ2Y2QtYmIxZS1lZjI2ODUwNzIyNmIiLCJ0aHJvdWdoIjoiYjM4NzMyY2MtODk1Mi00NmNkLWJiMWUtZWYyNjg1MDcyMjZiIiwidHlwZSI6InJvbGUifV19).
This TFC team has access to all MSP workspaces, and is [configured here](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/infrastructure/-/blob/terraform-cloud/terraform.tfvars?L44:1-48:4).

Note that you **must [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) before making your Entitle request**.
If you make your Entitle request, then log in, you will be removed from any team memberships granted through Entitle by Terraform Cloud's SSO implementation.

For more details, also see [creating and configuring services](https://github.com/sourcegraph/managed-services#operations).`,
		markdown.Code("msp"),
		markdown.Code("Managed Services Platform Operators"))

	return md.String()
}

func collectByOwner(services []*spec.Spec) ([]string, map[string]specSet) {
	m := make(map[string]specSet)
	for _, s := range services {
		for _, o := range s.Service.Owners {
			m[o] = append(m[o], s)
		}
	}

	owners := maps.Keys(m)
	slices.Sort(owners)
	for _, o := range owners {
		sort.Sort(m[o])
	}
	return owners, m
}

type specSet []*spec.Spec

func (s specSet) Len() int { return len(s) }
func (s specSet) Less(i, j int) bool {
	return s[i].Service.ID < s[j].Service.ID
}
func (s specSet) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
