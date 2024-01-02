package monitoring

import (
	"fmt"
	"path"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ObservableOwner denotes a team that owns an Observable. The current teams are described in
// the handbook: https://handbook.sourcegraph.com/departments/engineering/
type ObservableOwner struct {
	// opsgenieTeam is the team's name on OpsGenie and is used for routing alerts.
	opsgenieTeam string
	// human-friendly name for this team
	teamName string
	// path relative to handbookBaseURL for this team's page
	handbookSlug string
	// optional - defaults to /departments/engineering/teams
	handbookBasePath string
}

var (
	// ObservableOwnerInfraOrg represents the shared infra-org rotation which
	// currently manages Sourcegraph.com.
	ObservableOwnerInfraOrg = registerObservableOwner(ObservableOwner{
		opsgenieTeam:     "infra-support",
		handbookBasePath: "/departments/engineering",
		handbookSlug:     "infrastructure",
		teamName:         "Infrastructure Org",
	})

	ObservableOwnerSearch = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "search",
		handbookSlug: "search/product",
		teamName:     "Search",
	})
	ObservableOwnerSearchCore = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "search-core",
		handbookSlug: "search/core",
		teamName:     "Search Core",
	})
	ObservableOwnerBatches = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "batch-changes",
		handbookSlug: "batch-changes",
		teamName:     "Batch Changes",
	})
	ObservableOwnerCodeIntel = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "code-intel",
		handbookSlug: "code-intelligence",
		teamName:     "Code intelligence",
	})
	ObservableOwnerSource = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "source",
		handbookSlug: "source",
		teamName:     "Source",
	})
	ObservableOwnerCodeInsights = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "code-insights",
		handbookSlug: "code-insights",
		teamName:     "Code Insights",
	})
	ObservableOwnerDataAnalytics = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "data-analytics",
		handbookSlug: "data-analytics",
		teamName:     "Data & Analytics",
	})
	ObservableOwnerCody = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "cody",
		handbookSlug: "cody",
		teamName:     "Cody",
	})
	ObservableOwnerOwn = registerObservableOwner(ObservableOwner{
		opsgenieTeam: "own",
		teamName:     "own",
		handbookSlug: "own",
	})
)

// identifer must be all lowercase, and optionally  hyphenated.
//
// Some examples of valid identifiers:
// foo
// foo-bar
// foo-bar-baz
//
// Some examples of invalid identifiers:
// Foo
// FOO
// Foo-Bar
// foo_bar
var opsgenieTeamPattern = regexp.MustCompile("^([a-z]+)(-[a-z]+)*?$")

// validate does a simple offline validation that this owner is not a zero value
// and that the opsgenie team name matches the expected pattern.
func (o ObservableOwner) validate() error {
	var zero ObservableOwner
	if o == zero {
		return errors.New("Owner must be set")
	}

	if !opsgenieTeamPattern.Match([]byte(o.opsgenieTeam)) {
		return errors.Errorf(`Owner.opsgenieteam has invalid format: "%v"`, []byte(o.opsgenieTeam))
	}

	return nil
}

// toMarkdown returns a Markdown string that also links to the owner's team page in the handbook.
func (o ObservableOwner) toMarkdown() string {
	return fmt.Sprintf("[Sourcegraph %s team](%s)",
		o.teamName, o.getHandbookPageURL())
}

// getHandbookPageURL links to the owner's team page in the handbook.
func (o ObservableOwner) getHandbookPageURL() string {
	basePath := "/departments/engineering/teams"
	if o.handbookBasePath != "" {
		basePath = o.handbookBasePath
	}
	return "https://" + path.Join("handbook.sourcegraph.com", basePath, o.handbookSlug)
}

var allKnownOwners = make(map[string]ObservableOwner)

func registerObservableOwner(o ObservableOwner) ObservableOwner {
	if err := o.validate(); err != nil {
		panic(err)
	}
	if _, exists := allKnownOwners[o.teamName]; exists {
		panic(errors.Newf("duplicate ObservableOwner %+v", o))
	}
	allKnownOwners[o.teamName] = o
	return o
}
