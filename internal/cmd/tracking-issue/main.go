// Command tracking-issue uses the GitHub API to produce an iteration's tracking issue task list.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token")
	org := flag.String("org", "sourcegraph", "GitHub organization to list issues from")
	milestone := flag.String("milestone", "", "GitHub milestone to filter issues by")
	labels := flag.String("labels", "", "Comma separated list of labels to filter issues by")

	flag.Parse()

	if err := run(*token, *org, *milestone, *labels); err != nil {
		log.Fatal(err)
	}
}

func run(token, org, milestone, labels string) (err error) {
	ctx := context.Background()
	cli := github.NewClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)),
	)

	issues, err := listIssues(ctx, cli, org, strings.Split(labels, ",")...)
	if err != nil {
		return err
	}

	var (
		assignees []string
		workloads = map[string]float64{}
		items     = map[string][]string{}
	)

	for _, issue := range issues {
		if !partOf(milestone, issue) {
			continue
		}

		state := state(issue.State)
		estimate := estimate(issue.Labels)
		categories := categories(issue)
		assignee := assignee(issue.Assignee)
		title := title(issue)

		item := fmt.Sprintf("- [%s] %s [#%d](%s) __%s__ %s\n",
			state,
			title,
			*issue.Number,
			*issue.HTMLURL,
			estimate,
			emojis(categories),
		)

		if len(items[assignee]) == 0 {
			assignees = append(assignees, assignee)
		}

		items[assignee] = append(items[assignee], item)
		workloads[assignee] += days(estimate)
	}

	sort.Strings(assignees)

	for _, assignee := range assignees {
		fmt.Printf("\n%s: __%.2fd__\n\n", assignee, workloads[assignee])

		for _, item := range items[assignee] {
			fmt.Print(item)
		}
	}

	return nil
}

func title(issue *github.Issue) string {
	if *issue.Repository.Private {
		return *issue.Repository.FullName
	}
	return *issue.Title
}

func partOf(milestone string, issue *github.Issue) bool {
	return milestone != "" && issue.Milestone != nil &&
		*issue.Milestone.Title == milestone
}

func days(estimate string) float64 {
	d, _ := strconv.ParseFloat(strings.TrimSuffix(estimate, "d"), 64)
	return d
}

func estimate(labels []github.Label) string {
	const prefix = "estimate/"
	for _, l := range labels {
		if strings.HasPrefix(*l.Name, prefix) {
			return (*l.Name)[len(prefix):]
		}
	}
	return "?d"
}

func state(state *string) string {
	if state != nil && *state == "closed" {
		return "x"
	}
	return " "
}

func categories(issue *github.Issue) map[string]string {
	categories := make(map[string]string, len(issue.Labels))

	for _, l := range issue.Labels {
		var emoji string

		switch *l.Name {
		case "customer":
			emoji = customer(issue)
		case "roadmap":
			emoji = "🛠️"
		case "debt":
			emoji = "🧶"
		case "spike":
			emoji = "🕵️"
		case "bug":
			emoji = "🐛"
		case "security":
			emoji = "🔒"
		}

		if emoji != "" {
			categories[*l.Name] = emoji
		}
	}

	return categories
}

func emojis(categories map[string]string) string {
	// Generous four bytes for each emoji. We don't have
	// to be precise, since append will allocate more if needed.
	s := make([]byte, 0, 4*len(categories))
	for _, emoji := range categories {
		s = append(s, emoji...)
	}
	return string(s)
}

var matcher = regexp.MustCompile(`https://app\.hubspot\.com/contacts/2762526/company/\d+`)

func customer(issue *github.Issue) string {
	if issue == nil || issue.Body == nil {
		return ""
	}

	customer := matcher.FindString(*issue.Body)
	if customer == "" {
		return "👩"
	}

	return "[👩](" + customer + ")"
}

func assignee(user *github.User) string {
	if user == nil || user.Login == nil {
		return "Unassigned"
	}
	return "@" + *user.Login
}

func listIssues(ctx context.Context, cli *github.Client, org string, labels ...string) (issues []*github.Issue, _ error) {
	opt := &github.IssueListOptions{
		Filter:      "all",
		State:       "all",
		Labels:      labels,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		page, resp, err := cli.Issues.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}

		issues = append(issues, page...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return issues, nil
}
