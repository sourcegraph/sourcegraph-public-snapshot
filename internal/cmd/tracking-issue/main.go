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

	"github.com/shurcooL/githubv4"
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
	cli := githubv4.NewClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)),
	)

	issues, err := listIssues(ctx, cli, org, milestone, strings.Split(labels, ","))
	if err != nil {
		return err
	}

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
		title := title(issue)

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

func title(issue *Issue) string {
	if issue.Private {
		return issue.Repository
	}
	return issue.Title
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

	for _, label := range issue.Labels {
		var emoji string

		switch label {
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
			categories[label] = emoji
		}
	}

	return categories
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

	var q struct {
		Search struct {
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
			Nodes []struct {
				issue `graphql:"... on Issue"`
			}
		} `graphql:"search(first: 100, type: ISSUE, after: $cursor, query: $query)"`
	}

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"query":  githubv4.String(listIssuesSearchQuery(org, milestone, labels)),
	}

	for {
		err := cli.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}

		for _, n := range q.Search.Nodes {
			i := n.issue

			issue := &Issue{
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

		if !q.Search.PageInfo.HasNextPage {
			break
		}

		variables["cursor"] = githubv4.NewString(q.Search.PageInfo.EndCursor)
	}

	return issues, nil
}

func listIssuesSearchQuery(org, milestone string, labels []string) string {
	var q strings.Builder

	if org != "" {
		fmt.Fprintf(&q, "org:%q", org)
	}

	if milestone != "" {
		fmt.Fprintf(&q, " milestone:%q", milestone)
	}

	for _, label := range labels {
		if label != "" {
			fmt.Fprintf(&q, " label:%q", label)
		}
	}

	return q.String()
}
