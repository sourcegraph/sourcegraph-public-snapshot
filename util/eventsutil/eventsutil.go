package eventsutil

import (
	"strings"

	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

// parses a URL/URI in string form and returns the org, if there is one
func returnOrganization(link string) string {
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "www.")
	org, _, _ := githubutil.SplitGitHubRepoURI(link)
	return org
}
