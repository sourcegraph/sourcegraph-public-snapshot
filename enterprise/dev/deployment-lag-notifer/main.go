package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	_ "github.com/joho/godotenv/autoload"
)

// Flags are command-line arguments that configure the application behavior
// away from the defaults
type Flags struct {
	DryRun          bool
	Environment     string
	SlackToken      string
	SlackWebhookURL string
	SgDir           string
}

func (f *Flags) Parse() {
	flag.BoolVar(&f.DryRun, "dry-run", false, "Print to stdout instead of sending to Slack")
	flag.StringVar(&f.Environment, "env", Getenv("SG_ENVIRONMENT", "cloud"), "Environment to check against. Default: cloud")
	flag.StringVar(&f.SlackToken, "slack-token", os.Getenv("SLACK_TOKEN"), "Slack token")
	flag.StringVar(&f.SlackWebhookURL, "slack-webhook-url", os.Getenv("SLACK_WEBHOOK_URL"), "Slack webhook URL to post to")
	flag.Parse()
}

var environments = map[string]string{
	"cloud":   "https://sourcegraph.com",
	"k8s":     "https://k8s.sgdev.org",
	"preprod": "https://preview.sgdev.dev",
}

func Getenv(env, def string) string {
	val, present := os.LookupEnv(env)
	if !present {
		val = def
	}
	return val
}

func getLiveVersion(client *http.Client, url string) (string, error) {
	var version string

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/__version", url), nil)
	if err != nil {
		return version, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return version, err
	}

	if resp.StatusCode > http.StatusOK {
		return version, errors.Newf("received non-200 status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return version, err

	}

	// Response is in format build_date_hash
	parts := strings.Split(string(body), "_")

	if len(parts) != 3 {
		return version, errors.Newf("unknown version format %s", string(body))
	}

	version = parts[2]

	return version, nil
}

// GithubResponse is the response payload from requesting GET /repos/:author/:repo/commits
type GithubResponse []struct {
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name string `json:"name"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// Commit is a singular Git commit to a repo
type Commit struct {
	Sha     string
	Author  string
	Message string
}

func getCommitLog(client *http.Client) ([]Commit, error) {
	var commits []Commit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := "https://api.github.com/repos/sourcegraph/sourcegraph/commits"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return commits, err
	}

	q := req.URL.Query()
	q.Add("branch", "main")
	q.Add("per_page", "20")
	q.Add("page", "1")

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return commits, err
	}

	if resp.StatusCode > http.StatusOK {
		return commits, errors.Newf("received non-200 status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return commits, err
	}

	// fmt.Println(string(body))

	var gh GithubResponse
	err = json.Unmarshal(body, &gh)
	if err != nil {
		return commits, err
	}

	for _, g := range gh {
		lines := strings.Split(g.Commit.Message, "\n")
		message := g.Sha[:7]
		commits = append(commits,
			Commit{Sha: message, Author: g.Commit.Author.Name, Message: lines[0]})
	}

	return commits, nil
}

// checkForCommit checks for the current version in the
// last 20 commits
func checkForCommit(version string, commits []Commit) bool {
	found := false
	for _, c := range commits {
		if c.Sha == version[:7] {
			found = true
		}
	}

	return found
}

// createMessage posts the message to Slack or stdout if dry is set
func createMessage(version, environment string, commits []Commit) (bytes.Buffer, error) {
	var msg bytes.Buffer
	var slackTemplate = `:worried: *{{.Environment}}*'s deployments are out of date

Current version: ` + "`{{ .Version }}`" + `

Last 20 commits:
` + "```" + `
SHA{{"\t"}}Author{{"\t"}}Message
---{{"\t"}}---{{"\t"}}---
{{- range .CommitLog }}
{{ .Sha }}{{"\t"}}{{ .Author }}{{"\t"}}{{ .Message }}
{{- end }}
` + "```" + `

<https://sourcegraph.com|Handbook Page> | <https://github.com/sourcegraph/deploy-sourcegraph-cloud/pulls|deploy-sourcegraph-cloud> 

cc <!subteam^S02NFV6A536|devops-support>
`

	type templateData struct {
		Version     string
		CommitLog   []Commit
		Environment string
	}

	td := templateData{Version: version, CommitLog: commits, Environment: environment}
	// td := templateData{Version: version, Environment: environment}

	tpl, err := template.New("slack-message").Parse(slackTemplate)
	if err != nil {
		return msg, err
	}

	// tw := tabwriter.NewWriter(&msg, 0, 8, 2, '\t', 0)
	tw := tabwriter.NewWriter(&msg, 0, 8, 1, '\t', 0)

	err = tpl.Execute(tw, td)
	if err != nil {
		return msg, err
	}

	tw.Flush()

	return msg, nil
}

func main() {
	flags := &Flags{}
	flags.Parse()

	client := http.Client{}

	version, err := getLiveVersion(&client, environments[flags.Environment])
	if err != nil {
		log.Fatal(err)
	}

	commitLog, err := getCommitLog(&client)
	if err != nil {
		log.Fatal(err)
	}

	slack := NewSlackClient(flags.SlackWebhookURL)

	current := checkForCommit(version, commitLog)
	if current || !current {
		msg, err := createMessage(version[:7], flags.Environment, commitLog)
		if !flags.DryRun {
			err = slack.PostMessage(msg)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Println(string(msg.String()))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Cloud is current? %v\n", checkForCommit(version, commitLog))
}
