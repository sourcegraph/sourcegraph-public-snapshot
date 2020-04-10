// Command tracking-issue uses the GitHub API to produce an iteration's tracking issue task list.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/machinebox/graphql"
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
	cli := graphql.NewClient("https://api.github.com/graphql", graphql.WithHTTPClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))),
	)

	issues, prs, err := listIssuesAndPullRequests(ctx, cli, org, milestone, strings.Split(labels, ","))
	if err != nil {
		return err
	}

	tracking, err := trackingIssue(org, milestone, issues)
	if err != nil {
		return err
	}

	workloads := workloads(issues, prs, milestone)
	body, err := patchIssueBody(tracking, generate(workloads))
	if err != nil {
		return err
	}

	if body != tracking.Body {
		tracking.Body = body
	}

	if update {
		err = updateIssue(cli, tracking)
	}

	fmt.Println(tracking.Body)

	return err
}

func trackingIssue(org, milestone string, issues []*Issue) (*Issue, error) {
	var tracking []*Issue
	for _, issue := range issues {
		if issue.IsTrackingIssue(org, milestone) {
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

func patchIssueBody(issue *Issue, work string) (body string, err error) {
	const (
		openingMarker = "<!-- BEGIN PLANNED WORK -->"
		closingMarker = "<!-- END PLANNED WORK -->"
	)
	return patch(issue.Body, work, openingMarker, closingMarker)
}

func updateIssue(cli *graphql.Client, issue *Issue) (err error) {
	r := graphql.NewRequest(`mutation($input: UpdateIssueInput!) {
		updateIssue(input: $input) { issue { updatedAt } }
	}`)

	type UpdateIssueInput struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	}

	r.Var("input", &UpdateIssueInput{
		ID:   issue.ID,
		Body: issue.Body,
	})

	ctx := context.Background()
	return cli.Run(ctx, r, nil)
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

type Workload struct {
	Assignee          string
	Days              float64
	Issues            []*Issue
	PullRequests      []*PullRequest
	IssuePullRequests map[*Issue][]*PullRequest
	PullRequestIssues map[*PullRequest][]*Issue
}

func (wl *Workload) PrintTo(w io.Writer) error {
	_, err := fmt.Fprintf(w, "\n@%s: __%.2fd__\n\n", wl.Assignee, wl.Days)
	if err != nil {
		return err
	}

	for _, issue := range wl.Issues {
		if err = issue.PrintTo(w); err != nil {
			return err
		}

		for _, pr := range wl.IssuePullRequests[issue] {
			if err = pr.PrintTo(w); err != nil {
				return err
			}
		}
	}

	// Put all PRs that aren't linked to issues top-level
	for _, pr := range wl.PullRequests {
		if issues := wl.PullRequestIssues[pr]; len(issues) == 0 {
			if err = pr.PrintTo(w); err != nil {
				return err
			}
		}
	}

	return nil
}

func workloads(issues []*Issue, prs []*PullRequest, milestone string) map[string]*Workload {
	workloads := map[string]*Workload{}

	workload := func(assignee string) *Workload {
		w := workloads[assignee]
		if w == nil {
			w = &Workload{Assignee: assignee}
			workloads[assignee] = w
		}
		return w
	}

	for _, pr := range prs {
		w := workload(pr.Author)
		w.PullRequests = append(w.PullRequests, pr)
	}

	for _, issue := range issues {
		w := workload(Assignee(issue.Assignees))

		w.Issues = append(w.Issues, issue)
		if w.IssuePullRequests == nil {
			w.IssuePullRequests = make(map[*Issue][]*PullRequest)
		}

		if w.PullRequestIssues == nil {
			w.PullRequestIssues = make(map[*PullRequest][]*Issue)
		}

		linked := issue.LinkedPullRequests(prs)

		w.IssuePullRequests[issue] = linked
		for _, pr := range prs {
			w.PullRequestIssues[pr] = append(w.PullRequestIssues[pr], issue)
		}

		if issue.Milestone == milestone {
			estimate := Estimate(issue.Labels)
			w.Days += Days(estimate)
		} else {
			issue.Deprioritised = true
		}
	}

	return workloads
}

func generate(workloads map[string]*Workload) string {
	assignees := make([]string, 0, len(workloads))
	for assignee := range workloads {
		assignees = append(assignees, assignee)
	}

	sort.Strings(assignees)

	var b strings.Builder
	for _, assignee := range assignees {
		_ = workloads[assignee].PrintTo(&b)
	}
	return b.String()
}

func Days(estimate string) float64 {
	d, _ := strconv.ParseFloat(strings.TrimSuffix(estimate, "d"), 64)
	return d
}

func Estimate(labels []string) string {
	const prefix = "estimate/"
	for _, label := range labels {
		if strings.HasPrefix(label, prefix) {
			return label[len(prefix):]
		}
	}
	return ""
}

var matcher = regexp.MustCompile(`https://app\.hubspot\.com/contacts/2762526/company/\d+`)

func Customer(body string) string {
	customer := matcher.FindString(body)
	if customer == "" {
		return "👩"
	}
	return "[👩](" + customer + ")"
}

func Assignee(assignees []string) string {
	if len(assignees) == 0 {
		return "Unassigned"
	}
	return assignees[0]
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
	Author     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   time.Time

	Deprioritised bool
}

func (issue *Issue) PrintTo(w io.Writer) error {
	state := " "
	if strings.EqualFold(issue.State, "closed") {
		state = "x"
	}

	estimate := Estimate(issue.Labels)

	if estimate != "" {
		estimate = "__" + estimate + "__ "
	}

	_, err := fmt.Fprintf(w, "- [%s] %s [#%d](%s) %s%s\n",
		state,
		issue.title(),
		issue.Number,
		issue.URL,
		estimate,
		issue.Emojis(),
	)

	return err
}

func (issue *Issue) Emojis() string {
	categories := Categories(issue.Labels, issue.Repository, issue.Body)
	return Emojis(categories)
}

func (issue *Issue) IsTrackingIssue(org, milestone string) bool {
	return has("tracking", issue.Labels) &&
		strings.HasPrefix(issue.Repository, org) &&
		issue.Milestone == milestone
}

func Emojis(categories map[string]string) string {
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

func has(label string, labels []string) bool {
	for _, l := range labels {
		if label == l {
			return true
		}
	}
	return false
}

func (issue *Issue) title() string {
	var title string

	if issue.Private {
		title = issue.Repository
	} else {
		title = issue.Title
	}

	// Cross off issues that were originally planned
	// for the milestone but are no longer in it.
	if issue.Deprioritised {
		title = "~" + strings.TrimSpace(title) + "~"
	}

	return title
}

func (issue *Issue) LinkedPullRequests(prs []*PullRequest) (linked []*PullRequest) {
	for _, pr := range prs {
		if strings.Contains(pr.Body, "#"+strconv.Itoa(issue.Number)) {
			linked = append(linked, pr)
		}
	}
	return linked
}

func (issue *Issue) Redact() {
	if issue.Private {
		issue.Title = "REDACTED"
		issue.Labels = RedactLabels(issue.Labels)
	}
}

func RedactLabels(labels []string) []string {
	redacted := labels[:0]
	for _, label := range labels {
		if strings.HasPrefix(label, "estimate/") || strings.HasPrefix(label, "planned/") {
			redacted = append(redacted, label)
		}
	}
	return redacted
}

type PullRequest struct {
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
	Author     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   time.Time
	BeganAt    time.Time // Time of the first authored commit
}

func (pr *PullRequest) PrintTo(w io.Writer) error {
	state := " "
	if strings.EqualFold(pr.State, "merged") {
		state = "x"
	}

	var duration string
	if state == "x" {
		duration = "__" + pr.Duration() + "__ "
	}

	_, err := fmt.Fprintf(w, "  - [%s] %s [#%d](%s) %s%s\n",
		state,
		pr.title(),
		pr.Number,
		pr.URL,
		duration,
		pr.Emojis(),
	)

	return err
}

func (pr *PullRequest) Emojis() string {
	categories := Categories(pr.Labels, pr.Repository, pr.Body)
	categories["pull-request"] = ":shipit:"
	return Emojis(categories)
}

func (pr *PullRequest) title() string {
	var title string

	if pr.Private {
		title = pr.Repository
	} else {
		title = pr.Title
	}

	if strings.EqualFold(pr.State, "closed") {
		title = "~" + strings.TrimSpace(title) + "~"
	}

	return title
}

func (pr *PullRequest) Duration() string {
	if d := pr.ClosedAt.Sub(pr.BeganAt); d > 0 {
		return duration(d)
	}
	return ""
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func duration(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%s", days, d)

	return b.String()
}

func (pr *PullRequest) Redact() {
	if pr.Private {
		pr.Title = "REDACTED"
		pr.Labels = RedactLabels(pr.Labels)
	}
}

func Categories(labels []string, repository, body string) map[string]string {
	categories := make(map[string]string, len(labels))

	switch repository {
	case "sourcegraph/customer":
		categories["customer"] = Customer(body)
	case "sourcegraph/security-prs":
		categories["security"] = Emoji("security")
	}

	for _, label := range labels {
		if label == "customer" {
			categories[label] = Customer(body)
		} else if emoji := Emoji(label); emoji != "" {
			categories[label] = emoji
		}
	}

	return categories
}

func Emoji(category string) string {
	switch category {
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

func listIssuesAndPullRequests(ctx context.Context, cli *graphql.Client, org, milestone string, labels []string) (issues []*Issue, prs []*PullRequest, _ error) {
	type searchNode struct {
		Typename   string `json:"__typename"`
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
		Author    struct{ Login string }
		Assignees struct{ Nodes []struct{ Login string } }
		Labels    struct{ Nodes []struct{ Name string } }
		Milestone struct{ Title string }
		Commits   struct {
			Nodes []struct {
				Commit struct{ AuthoredDate time.Time }
			}
		}
		CreatedAt time.Time
		UpdatedAt time.Time
		ClosedAt  time.Time
	}

	type search struct {
		PageInfo struct {
			EndCursor   string
			HasNextPage bool
		}
		Nodes []searchNode
	}

	var (
		milestonedCount    = 100
		demilestonedCount  = 100
		milestonedCursor   string
		demilestonedCursor string
		milestonedQuery    = listIssuesSearchQuery(org, milestone, labels, false)
		demilestonedQuery  = listIssuesSearchQuery(org, milestone, labels, true)
	)

	for {
		var data struct {
			Milestoned, Demilestoned search
		}

		var q strings.Builder
		q.WriteString("query(" +
			"$demilestonedCount: Int!," +
			"$demilestonedCursor: String," +
			"$demilestonedQuery: String!," +
			"$milestonedCount: Int!," +
			"$milestonedCursor: String," +
			"$milestonedQuery: String!) {\n")

		q.WriteString(listIssuesGraphQLQuery("milestoned"))
		q.WriteString(listIssuesGraphQLQuery("demilestoned"))

		q.WriteString("}")

		r := graphql.NewRequest(q.String())
		r.Var("milestonedCount", milestonedCount)
		r.Var("demilestonedCount", demilestonedCount)

		if milestonedCursor != "" {
			r.Var("milestonedCursor", milestonedCursor)
		}

		if demilestonedCursor != "" {
			r.Var("demilestonedCursor", demilestonedCursor)
		}

		r.Var("milestonedQuery", milestonedQuery)
		r.Var("demilestonedQuery", demilestonedQuery)

		err := cli.Run(ctx, r, &data)
		if err != nil {
			return nil, nil, err
		}

		nodes := append(data.Milestoned.Nodes, data.Demilestoned.Nodes...)

		for _, n := range nodes {
			switch n.Typename {
			case "PullRequest":
				pr := &PullRequest{
					ID:         n.ID,
					Title:      n.Title,
					Body:       n.Body,
					State:      n.State,
					Number:     n.Number,
					URL:        n.URL,
					Repository: n.Repository.NameWithOwner,
					Private:    n.Repository.IsPrivate,
					Assignees:  make([]string, 0, len(n.Assignees.Nodes)),
					Labels:     make([]string, 0, len(n.Labels.Nodes)),
					Milestone:  n.Milestone.Title,
					Author:     n.Author.Login,
					CreatedAt:  n.CreatedAt,
					UpdatedAt:  n.UpdatedAt,
					ClosedAt:   n.ClosedAt,
					BeganAt:    n.Commits.Nodes[0].Commit.AuthoredDate,
				}

				for _, assignee := range n.Assignees.Nodes {
					pr.Assignees = append(pr.Assignees, assignee.Login)
				}

				for _, label := range n.Labels.Nodes {
					pr.Labels = append(pr.Labels, label.Name)
				}

				prs = append(prs, pr)

			case "Issue":
				issue := &Issue{
					ID:         n.ID,
					Title:      n.Title,
					Body:       n.Body,
					State:      n.State,
					Number:     n.Number,
					URL:        n.URL,
					Repository: n.Repository.NameWithOwner,
					Private:    n.Repository.IsPrivate,
					Assignees:  make([]string, 0, len(n.Assignees.Nodes)),
					Labels:     make([]string, 0, len(n.Labels.Nodes)),
					Milestone:  n.Milestone.Title,
					Author:     n.Author.Login,
					CreatedAt:  n.CreatedAt,
					UpdatedAt:  n.UpdatedAt,
					ClosedAt:   n.ClosedAt,
				}

				for _, assignee := range n.Assignees.Nodes {
					issue.Assignees = append(issue.Assignees, assignee.Login)
				}

				for _, label := range n.Labels.Nodes {
					issue.Labels = append(issue.Labels, label.Name)
				}

				issues = append(issues, issue)
			}
		}

		var hasNextPage bool
		if data.Milestoned.PageInfo.HasNextPage {
			hasNextPage = true
			milestonedCursor = data.Milestoned.PageInfo.EndCursor
		} else {
			milestonedCount = 0
		}

		if data.Demilestoned.PageInfo.HasNextPage {
			hasNextPage = true
			demilestonedCursor = data.Demilestoned.PageInfo.EndCursor
		} else {
			demilestonedCount = 0
		}

		if !hasNextPage {
			break
		}
	}

	return issues, prs, nil
}

func listIssuesGraphQLQuery(alias string) string {
	const searchQuery = `%s: search(first: $%sCount, type: ISSUE, after: $%sCursor query: $%sQuery) {
		pageInfo {
			endCursor
			hasNextPage
		}
		nodes {
			... on Issue {
				%s
			}
			... on PullRequest {
				%s
			}
		}
	}`

	return fmt.Sprintf(searchQuery,
		alias,
		alias,
		alias,
		alias,
		searchNodeFields(false),
		searchNodeFields(true),
	)
}

func searchNodeFields(isPR bool) string {
	fields := `
		__typename
		id, title, body, state, number, url
		createdAt, closedAt
		repository { nameWithOwner, isPrivate }
		author { login }
		assignees(first: 25) { nodes { login } }
		labels(first: 25) { nodes { name } }
		milestone { title }
	`

	if isPR {
		fields += `
			commits(first: 1) { nodes { commit { authoredDate } } }
		`
	}

	return fields
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
