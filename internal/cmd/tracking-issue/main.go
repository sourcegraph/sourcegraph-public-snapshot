// Command tracking-issue uses the GitHub API to produce an iteration's tracking issue task list.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

func main() {
	milestone := flag.String("milestone", "", "GitHub milestone to filter issues by")
	labels := flag.String("labels", "", "Comma separated list of labels to filter issues by")

	flag.Parse()

	if err := run(*milestone, *labels); err != nil {
		log.Fatal(err)
	}
}

func run(milestone, labels string) error {
	cli := github.NewClient(nil)
	ctx := context.Background()

	m, err := getMilestoneByTitle(ctx, cli, milestone)
	if err != nil {
		return err
	}

	issues, err := listIssues(ctx, cli, *m.Number, strings.Split(labels, ",")...)
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
		category := category(issue)
		assignee := assignee(issue.Assignee)

		item := fmt.Sprintf("- [%s] %s [#%d](%s) __%s__ %s\n",
			state,
			*issue.Title,
			*issue.Number,
			*issue.HTMLURL,
			estimate,
			category,
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

func category(issue *github.Issue) string {
	for _, l := range issue.Labels {
		switch *l.Name {
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
		}
	}
	return "❓"
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

func getMilestoneByTitle(ctx context.Context, cli *github.Client, title string) (*github.Milestone, error) {
	opt := &github.MilestoneListOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {
		milestones, resp, err := cli.Issues.ListMilestones(ctx, "sourcegraph", "sourcegraph", opt)
		if err != nil {
			return nil, err
		}

		for _, m := range milestones {
			if *m.Title == title {
				return m, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return nil, errors.New("milestone not found")
}

func listIssues(ctx context.Context, cli *github.Client, milestone int, labels ...string) (issues []*github.Issue, _ error) {
	opt := &github.IssueListByRepoOptions{
		Milestone:   strconv.Itoa(milestone),
		Labels:      labels,
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		page, resp, err := cli.Issues.ListByRepo(ctx, "sourcegraph", "sourcegraph", opt)
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
