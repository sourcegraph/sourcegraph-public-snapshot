// Command tracking-issue uses the GitHub API to maintain open tracking issues.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/machinebox/graphql"
	"golang.org/x/oauth2"
)

const (
	beginWorkMarker        = "<!-- BEGIN WORK -->"
	endWorkMarker          = "<!-- END WORK -->"
	labelMarkerRegexp      = "<!-- LABEL: (.*) -->"
	beginAssigneeMarkerFmt = "<!-- BEGIN ASSIGNEE: %s -->"
	endAssigneeMarker      = "<!-- END ASSIGNEE -->"
)

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token")
	org := flag.String("org", "sourcegraph", "GitHub organization to list issues from")
	dry := flag.Bool("dry", false, "If true, do not update GitHub tracking issues in-place, but print them to stdout")
	verbose := flag.Bool("verbose", false, "If true, print the resulting tracking issue bodies to stdout")

	flag.Parse()

	if err := run(*token, *org, *dry, *verbose); err != nil {
		log.Fatal(err)
	}
}

func run(token, org string, dry, verbose bool) (err error) {
	if token == "" {
		return fmt.Errorf("no -token given")
	}

	if org == "" {
		return fmt.Errorf("no -org given")
	}

	ctx := context.Background()
	cli := graphql.NewClient("https://api.github.com/graphql", graphql.WithHTTPClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))),
	)

	tracking, err := listTrackingIssues(ctx, cli, fmt.Sprintf("org:%q label:tracking is:open", org))
	if err != nil {
		return err
	}

	if len(tracking) == 0 {
		log.Printf("No tracking issues found. Exiting.")
		return nil
	}

	err = loadTrackingIssues(ctx, cli, org, tracking)
	if err != nil {
		return err
	}

	var toUpdate []*Issue
	for _, issue := range tracking {
		work := issue.Workloads().Markdown(issue.LabelAllowlist)
		if updated, err := issue.UpdateWork(work); err != nil {
			log.Printf("failed to patch work section in %q %s: %v", issue.Title, issue.URL, err)
		} else if !updated {
			log.Printf("%q %s not modified.", issue.Title, issue.URL)
		} else if !dry {
			log.Printf("%q %s modified", issue.Title, issue.URL)
			toUpdate = append(toUpdate, issue.Issue)
		} else {
			log.Printf("%q %s modified, but not updated due to -dry=true.", issue.Title, issue.URL)
		}

		if verbose {
			log.Printf("%q %s body\n%s\n\n", issue.Title, issue.URL, issue.Body)
		}
	}

	if len(toUpdate) > 0 {
		return updateIssues(ctx, cli, toUpdate)
	}

	return nil
}
