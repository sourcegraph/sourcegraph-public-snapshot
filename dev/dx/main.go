package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var (
	BuildCommit = "dev"
)

var stdOut *output.Output

func main() {
	if err := dx.RunContext(context.Background(), os.Args); err != nil {
		// We want to prefer an already-initialized std.Out no matter what happens,
		// because that can be configured (e.g. with '--disable-output-detection'). Only
		// if something went horribly wrong and std.Out is not yet initialized should we
		// attempt an initialization here.
		if stdOut == nil {
			stdOut = output.NewOutput(os.Stdout, output.OutputOpts{})
		}
		// Do not treat error message as a format string
		log.Fatal(err)
	}
}

var dx = &cli.App{
	Usage:       "The internal CLI used by the DevX team",
	Description: "TODO",
	Version:     BuildCommit,
	Compiled:    time.Now(),
	Commands:    []*cli.Command{betsCommand},
}

type IssueStore struct {
	cache map[int]*github.Issue
	gh    *github.Client
	roots []*TrackingIssue
}

type TrackingIssue struct {
	*github.Issue
	parent   *TrackingIssue
	children []*TrackingIssue
}

func (t *TrackingIssue) PrettyString() string {
	return fmt.Sprintf("%8s %s", fmt.Sprintf("#%d", t.GetNumber()), t.GetTitle())
}

func (t *TrackingIssue) IsImportant() bool {
	return strings.HasPrefix(t.GetTitle(), "⭐")
}

func NewIssueStore(gh *github.Client) *IssueStore {
	return &IssueStore{
		cache: make(map[int]*github.Issue),
		gh:    gh,
	}
}

func (s *IssueStore) Populate(ctx context.Context) error {
	if err := s.search(ctx); err != nil {
		return err
	}
	if err := s.findTrackedIssues(ctx); err != nil {
		return err
	}

	includes := func(issues []*TrackingIssue, issue *TrackingIssue) bool {
		for _, i := range issues {
			if i.GetNumber() == issue.GetNumber() {
				return true
			}
		}
		return false
	}

	for index, root := range s.roots {
		for _, other := range s.roots {
			if includes(other.children, root) {
				s.roots = append(s.roots[:index], s.roots[index+1:]...)
			}
		}
	}
	return nil
}

func (s *IssueStore) get(ctx context.Context, number int) (*github.Issue, error) {
	if issue, ok := s.cache[number]; ok {
		return issue, nil
	}

	issue, _, err := s.gh.Issues.Get(ctx, "sourcegraph", "sourcegraph", number)
	if err != nil {
		return nil, err
	}
	s.cache[issue.GetNumber()] = issue
	return issue, nil
}

func (s *IssueStore) findTrackedIssues(ctx context.Context) error {
	for _, root := range s.roots {
		body := root.GetBody()
		scanner := bufio.NewScanner(strings.NewReader(body))
		for scanner.Scan() {
			line := scanner.Text()
			res := taskRegexp.FindStringSubmatch(line)
			for _, id := range res {
				n, err := strconv.Atoi(id)
				if err != nil {
					continue
				}
				issue, err := s.get(ctx, n)
				if err != nil {
					return err
				}
				var dup bool
				for _, child := range root.children {
					if child.Issue.GetNumber() == issue.GetNumber() {
						dup = true
						break
					}
				}
				if !dup {
					root.children = append(root.children, &TrackingIssue{parent: root, Issue: issue})
				}
			}
		}
	}
	return nil
}

func (s *IssueStore) search(ctx context.Context) error {
	opts := &github.SearchOptions{
		Sort:        "created",
		TextMatch:   true,
		ListOptions: github.ListOptions{},
	}
	for {
		res, resp, err := s.gh.Search.Issues(ctx, "repo:sourcegraph/sourcegraph Q3B1 label:team/devx", opts)
		if err != nil {
			return err
		}
		for _, i := range res.Issues {
			if strings.HasPrefix(i.GetTitle(), "⭐") {
				s.cache[i.GetNumber()] = i
				s.roots = append(s.roots, &TrackingIssue{Issue: i})
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}

var betsCommand = &cli.Command{
	Name:      "bets",
	Usage:     "",
	UsageText: "",
	Subcommands: []*cli.Command{
		{
			Name:        "list",
			Description: "TODO",
			Action: func(cmd *cli.Context) error {
				gh := newGHClient(cmd.Context)
				store := NewIssueStore(gh)
				if err := store.Populate(cmd.Context); err != nil {
					return err
				}
				for _, issue := range store.roots {
					fmt.Printf("%s\n", issue.PrettyString())
					for _, child := range issue.children {
						if child.IsImportant() {
							fmt.Printf("\t%s\n", child.PrettyString())
						}
					}
				}
				return nil
			},
		},
	},
}

var taskRegexp = regexp.MustCompile(`\s*- \[[x ]?\] ?(?:https://github.com/sourcegraph/sourcegraph/issues/(\d+))|(?:#(\d+))`)

func newGHClient(ctx context.Context) *github.Client {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))
	gh := github.NewClient(tc)
	return gh
}
