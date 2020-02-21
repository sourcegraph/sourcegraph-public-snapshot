// Command tracking-issue uses the GitHub API to produce an iteration's tracking issue task list.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "GitHub personal access token")
	org := flag.String("org", "sourcegraph", "GitHub organization to list issues from")
	milestone := flag.String("milestone", "", "GitHub milestone to filter issues by")
	labels := flag.String("labels", "", "Comma separated list of labels to filter issues by")
	update := flag.Bool("update", false, "Update GitHub tracking issue in-place")

	flag.Parse()

	if err := run(*token, *org, *milestone, *labels, *update); err != nil {
		log.Fatal(err)
	}
}

func run(token, org, milestone, labels string, update bool) (err error) {
	if token == "" {
		return fmt.Errorf("no -token given")
	}

	if org == "" {
		return fmt.Errorf("no -org given")
	}

	if milestone == "" {
		return fmt.Errorf("no -milestone given")
	}

	ctx := context.Background()
	cli := githubv4.NewClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)),
	)

	issues, err := listIssues(ctx, cli, org, milestone, strings.Split(labels, ","))
	if err != nil {
		return err
	}

	tracking, err := trackingIssue(org, milestone, issues)
	if err != nil {
		return err
	}

	work := generate(issues, milestone)

	body, err := patchIssueBody(tracking, work)
	if err != nil {
		return err
	}

	if update && body != tracking.Body {
		tracking.Body = body
		err = updateIssue(cli, tracking)
	}

	fmt.Println(tracking.Body)
	return err
}

func trackingIssue(org, milestone string, issues []*Issue) (*Issue, error) {
	var tracking []*Issue
	for _, issue := range issues {
		if isTrackingIssue(issue, org, milestone) {
			tracking = append(tracking, issue)
		}
	}

	switch len(tracking) {
	case 0:
		return nil, errors.New("no tracking issue found")
	case 1:
		return tracking[0], nil
	default:
		return nil, errors.New("more than one tracking issue found")
	}
}

func isTrackingIssue(issue *Issue, org, milestone string) bool {
	return has("tracking", issue.Labels) &&
		strings.HasPrefix(issue.Repository, org) &&
		issue.Milestone == milestone
}

func has(label string, labels []string) bool {
	for _, l := range labels {
		if label == l {
			return true
		}
	}
	return false
}

func patchIssueBody(issue *Issue, work string) (body string, err error) {
	const (
		openingMarker = "<!-- BEGIN PLANNED WORK -->"
		closingMarker = "<!-- END PLANNED WORK -->"
	)
	return patch(issue.Body, work, openingMarker, closingMarker)
}

func updateIssue(cli *githubv4.Client, issue *Issue) (err error) {
	var m struct {
		UpdateIssue struct {
			Issue struct {
				UpdatedAt time.Time
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	input := githubv4.UpdateIssueInput{
		ID:   issue.ID,
		Body: githubv4.NewString(githubv4.String(issue.Body)),
	}

	ctx := context.Background()
	return cli.Mutate(ctx, &m, input, nil)
}

func patch(s, replacement, opening, closing string) (string, error) {
	start := strings.Index(s, opening)
	if start == -1 {
		return s, errors.New("could not find opening marker in issue body")
	}

	end := strings.Index(s, closing)
	if end == -1 {
		return s, errors.New("could not find closing marker in issue body")
	}

	return s[:start+len(opening)] + replacement + s[end:], nil
}

func generate(issues []*Issue, milestone string) string {
	var (
		assignees []string
		workloads = map[string]float64{}
		items     = map[string][]string{}
	)

	for _, issue := range issues {
		state := state(issue.State)
		estimate := estimate(issue.Labels)
		categories := categories(issue)
		assignee := assignee(issue.Assignees)
		title := title(issue, milestone)

		item := fmt.Sprintf("- [%s] %s [#%d](%s) __%s__ %s\n",
			state,
			title,
			issue.Number,
			issue.URL,
			estimate,
			emojis(categories),
		)

		if len(items[assignee]) == 0 {
			assignees = append(assignees, assignee)
		}

		items[assignee] = append(items[assignee], item)

		// Exclude work that is no longer planned
		if issue.Milestone == milestone {
			workloads[assignee] += days(estimate)
		}
	}

	sort.Strings(assignees)

	var w strings.Builder
	for _, assignee := range assignees {
		fmt.Fprintf(&w, "\n%s: __%.2fd__\n\n", assignee, workloads[assignee])

		for _, item := range items[assignee] {
			fmt.Fprint(&w, item)
		}
	}

	return w.String()
}

func title(issue *Issue, milestone string) string {
	var title string

	if issue.Private {
		title = issue.Repository
	} else {
		title = issue.Title
	}

	// Cross off issues that were originally planned
	// for the milestone but are no longer in it.
	if issue.Milestone != milestone {
		title = "~" + strings.TrimSpace(title) + "~"
	}

	return title
}

func days(estimate string) float64 {
	d, _ := strconv.ParseFloat(strings.TrimSuffix(estimate, "d"), 64)
	return d
}

func estimate(labels []string) string {
	const prefix = "estimate/"
	for _, label := range labels {
		if strings.HasPrefix(label, prefix) {
			return label[len(prefix):]
		}
	}
	return "?d"
}

func state(state string) string {
	if strings.EqualFold(state, "closed") {
		return "x"
	}
	return " "
}

func categories(issue *Issue) map[string]string {
	categories := make(map[string]string, len(issue.Labels))

	switch issue.Repository {
	case "sourcegraph/customer":
		categories["customer"] = emoji("customer", issue)
	case "sourcegraph/security-issues":
		categories["security"] = emoji("security", issue)
	}

	for _, label := range issue.Labels {
		if emoji := emoji(label, issue); emoji != "" {
			categories[label] = emoji
		}
	}

	return categories
}

func emoji(category string, issue *Issue) string {
	switch category {
	case "customer":
		return customer(issue)
	case "roadmap":
		return "🛠️"
	case "debt":
		return "🧶"
	case "spike":
		return "🕵️"
	case "bug":
		return "🐛"
	case "security":
		return "🔒"
	default:
		return ""
	}
}

func emojis(categories map[string]string) string {
	sorted := make([]string, 0, len(categories))
	length := 0

	for _, emoji := range categories {
		sorted = append(sorted, emoji)
		length += len(emoji)
	}

	sort.Strings(sorted)

	s := make([]byte, 0, length)
	for _, emoji := range sorted {
		s = append(s, emoji...)
	}

	return string(s)
}

var matcher = regexp.MustCompile(`https://app\.hubspot\.com/contacts/2762526/company/\d+`)

func customer(issue *Issue) string {
	customer := matcher.FindString(issue.Body)
	if customer == "" {
		return "👩"
	}
	return "[👩](" + customer + ")"
}

func assignee(assignees []string) string {
	if len(assignees) == 0 {
		return "Unassigned"
	}
	return "@" + assignees[0]
}

type Issue struct {
	ID         string
	Title      string
	Body       string
	Number     int
	URL        string
	State      string
	Repository string
	Private    bool
	Labels     []string
	Assignees  []string
	Milestone  string
}

func listIssues(ctx context.Context, cli *githubv4.Client, org, milestone string, labels []string) (issues []*Issue, _ error) {
	type issue struct {
		ID         string
		Title      string
		Body       string
		State      string
		Number     int
		URL        string
		Repository struct {
			NameWithOwner string
			IsPrivate     bool
		}
		Assignees struct{ Nodes []struct{ Login string } } `graphql:"assignees(first: 25)"`
		Labels    struct{ Nodes []struct{ Name string } }  `graphql:"labels(first:25)"`
		Milestone struct{ Title string }
	}

	type search struct {
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
		Nodes []struct {
			issue `graphql:"... on Issue"`
		}
	}

	var q struct {
		Milestoned   search `graphql:"milestoned: search(first: $milestonedCount, type: ISSUE, after: $milestonedCursor, query: $milestonedQuery)"`
		Demilestoned search `graphql:"demilestoned: search(first: $demilestonedCount, type: ISSUE, after: $demilestonedCursor, query: $demilestonedQuery)"`
	}

	variables := map[string]interface{}{
		"milestonedCount":    githubv4.Int(100),
		"demilestonedCount":  githubv4.Int(100),
		"milestonedCursor":   (*githubv4.String)(nil),
		"demilestonedCursor": (*githubv4.String)(nil),
		"milestonedQuery":    githubv4.String(listIssuesSearchQuery(org, milestone, labels, false)),
		"demilestonedQuery":  githubv4.String(listIssuesSearchQuery(org, milestone, labels, true)),
	}

	var emptyIssue issue

	for {
		err := cli.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}

		nodes := append(q.Milestoned.Nodes, q.Demilestoned.Nodes...)

		for _, n := range nodes {
			// GitHub's GraphQL API sometimes sends empty issue nodes.
			if reflect.DeepEqual(n.issue, emptyIssue) {
				continue
			}

			i := n.issue

			issue := &Issue{
				ID:         i.ID,
				Title:      i.Title,
				Body:       i.Body,
				State:      i.State,
				Number:     i.Number,
				URL:        i.URL,
				Repository: i.Repository.NameWithOwner,
				Private:    i.Repository.IsPrivate,
				Assignees:  make([]string, 0, len(i.Assignees.Nodes)),
				Labels:     make([]string, 0, len(i.Labels.Nodes)),
				Milestone:  i.Milestone.Title,
			}

			for _, assignee := range i.Assignees.Nodes {
				issue.Assignees = append(issue.Assignees, assignee.Login)
			}

			for _, label := range i.Labels.Nodes {
				issue.Labels = append(issue.Labels, label.Name)
			}

			issues = append(issues, issue)
		}

		var hasNextPage bool
		if q.Milestoned.PageInfo.HasNextPage {
			hasNextPage = true
			variables["milestonedCursor"] = githubv4.NewString(q.Milestoned.PageInfo.EndCursor)
		} else {
			variables["milestonedCount"] = githubv4.Int(0)
		}

		if q.Demilestoned.PageInfo.HasNextPage {
			hasNextPage = true
			variables["demilestonedCursor"] = githubv4.NewString(q.Demilestoned.PageInfo.EndCursor)
		} else {
			variables["demilestonedCount"] = githubv4.Int(0)
		}

		if !hasNextPage {
			break
		}
	}

	return issues, nil
}

func listIssuesSearchQuery(org, milestone string, labels []string, demilestoned bool) string {
	var q strings.Builder

	fmt.Fprintf(&q, "org:%q", org)

	if demilestoned {
		fmt.Fprintf(&q, ` -milestone:%q label:"planned/%s"`, milestone, milestone)
	} else {
		fmt.Fprintf(&q, " milestone:%q", milestone)
	}

	for _, label := range labels {
		if label != "" {
			fmt.Fprintf(&q, " label:%q", label)
		}
	}

	return q.String()
}
