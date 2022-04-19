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
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Flags struct {
	GitHubToken          string
	DryRun               bool
	Environment          string
	SlackToken           string
	SlackAnnounceWebhook string
	HoneycombToken       string
	OkayHQToken          string
	BaseDir              string
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubToken, "github.token", os.Getenv("GITHUB_TOKEN"), "mandatory github token")
	flag.StringVar(&f.Environment, "environment", "", "Environment being deployed")
	flag.BoolVar(&f.DryRun, "dry", false, "Pretend to post notifications, printing to stdout instead")
	flag.StringVar(&f.SlackToken, "slack.token", "", "mandatory slack api token")
	flag.StringVar(&f.SlackAnnounceWebhook, "slack.webhook", "", "Slack Webhook URL to post the results on")
	flag.StringVar(&f.HoneycombToken, "honeycomb.token", "", "mandatory honeycomb api token")
	flag.StringVar(&f.OkayHQToken, "okayhq.token", "", "mandatory okayhq api token")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	flags := &Flags{}
	flags.Parse()
	if flags.Environment == "" {
		log.Fatalf("-environment must be specified: preprod or production.")
	}

	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: flags.GitHubToken},
	)))
	if flags.GitHubToken == "" {
		log.Println("warning: using unauthenticated github client")
		ghc = github.NewClient(http.DefaultClient)
	}

	changedFiles, err := getChangedFiles()
	if err != nil {
		log.Fatal(err)
	}
	if len(changedFiles) == 0 {
		fmt.Println(":warning: No relevant changes, skipping notifications and exiting normally.")
		return
	}

	manifestRevision, err := getRevision()
	if err != nil {
		log.Fatal(err)
	}

	dd := NewManifestDeploymentDiffer(changedFiles)
	dn := NewDeploymentNotifier(
		ghc,
		dd,
		flags.Environment,
		manifestRevision,
	)

	report, err := dn.Report(ctx)
	if err != nil {
		if errors.Is(err, ErrNoRelevantChanges) {
			fmt.Println(":warning: No relevant changes, skipping notifications and exiting normally.")
			return
		}
		log.Fatal(err)
	}

	// Tracing
	var traceURL string
	if flags.HoneycombToken != "" {
		traceURL, err = reportDeployTrace(report, flags.HoneycombToken, flags.DryRun)
		if err != nil {
			log.Fatal("trace: ", err.Error())
		}
	}

	// Metrics
	if flags.OkayHQToken != "" {
		err := reportDeploymentMetrics(report, flags.OkayHQToken, flags.DryRun)
		if err != nil {
			log.Fatal("metrics: ", err.Error())
		}
		fmt.Println("okayhq")
	}

	// Notifcations
	slc := slack.New(flags.SlackToken)
	teammates := team.NewTeammateResolver(ghc, slc)
	if flags.DryRun {
		fmt.Println("Github\n---")
		for _, pr := range report.PullRequests {
			fmt.Println("-", pr.GetNumber())
		}
		out, err := renderComment(report, traceURL)
		if err != nil {
			log.Fatalf("can't render GitHub comment %q", err)
		}
		fmt.Println(out)
		fmt.Println("Slack\n---")
		out, err = slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			log.Fatalf("can't render Slack post %q", err)
		}
		fmt.Println(out)
	} else {
		out, err := slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			log.Fatalf("can't render Slack post %q", err)
		}
		err = postSlackUpdate(flags.SlackAnnounceWebhook, out)
		if err != nil {
			log.Fatalf("can't post Slack update %q", err)
		}
	}
}

func getChangedFiles() ([]string, error) {
	diffCommand := []string{"diff", "--name-only", "@^"}
	if output, err := exec.Command("git", diffCommand...).Output(); err != nil {
		return nil, err
	} else {
		strOutput := string(output)
		strOutput = strings.TrimSpace(strOutput)
		if strOutput == "" {
			return nil, nil
		}
		return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
	}
}

func getRevision() (string, error) {
	diffCommand := []string{"rev-list", "-1", "HEAD", "."}
	output, err := exec.Command("git", diffCommand...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

type okayEvent struct {
	Name       string                 `json:"event"`
	Timestamp  time.Time              `json:"timestamp"`
	Properties map[string]interface{} `json:"properties"`
}

type OkayMetricsClient struct {
	token  string
	cli    *http.Client
	events []*okayEvent
	mu     sync.Mutex
}

func NewOkayMetricsClient(client *http.Client, token string) *OkayMetricsClient {
	return &OkayMetricsClient{
		cli:   client,
		token: token,
	}
}

func (o *OkayMetricsClient) post(event *okayEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, "https://app.okayhq.com/api/webhooks/events", buf)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", o.token))
	resp, err := o.cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("can't read response body")
		}
		defer resp.Body.Close()
		return errors.Newf("failed to submit custom metric to OkayHQ: %q", string(body))
	}
	return nil
}

func (o *OkayMetricsClient) Push(name string, ts time.Time, properties map[string]interface{}) error {
	if name == "" {
		return errors.New("Okay metrics event name can't be blank")
	}
	if ts.IsZero() {
		return errors.New("Okay metrics event timestamp name can't be zero")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.events = append(o.events, &okayEvent{
		Name:       name,
		Timestamp:  ts,
		Properties: properties,
	})

	return nil
}

func (o *OkayMetricsClient) Flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var errs error
	for _, event := range o.events {
		err := o.post(event)
		if err != nil {
			errs = errors.Append(err)
		}
	}
	// Reset the internal events buffer
	o.events = nil
	return errs
}

func reportDeploymentMetrics(report *DeploymentReport, token string, dryRun bool) error {
	if dryRun {
		return nil
	}

	deployTime, err := time.Parse(time.RFC822Z, report.DeployedAt)
	if err != nil {
		return errors.Wrap(err, "r.DeployedAt")
	}

	okayCli := NewOkayMetricsClient(http.DefaultClient, token)

	for _, pr := range report.PullRequests {
		durationInMin := deployTime.Sub(pr.GetMergedAt()) / time.Minute
		okayCli.Push("qa.deployment", deployTime, map[string]interface{}{
			// context
			"environment":           report.Environment,
			"author":                pr.GetUser().GetLogin(),
			"pull_request.number":   pr.GetNumber(),
			"pull_request.title":    pr.GetTitle(),
			"pull_request.revision": pr.GetMergeCommitSHA(),
			"pull_request.url":      pr.GetHTMLURL(),

			// duration
			"duration.minutes": durationInMin,
		})
	}
	return okayCli.Flush()
}

func reportDeployTrace(report *DeploymentReport, token string, dryRun bool) (string, error) {
	honeyConfig := libhoney.Config{
		APIKey:  token,
		APIHost: "https://api.honeycomb.io/",
		Dataset: "deploy-sourcegraph",
	}
	if dryRun {
		honeyConfig.Transmission = &transmission.WriterSender{} // prints events to stdout instead
	}
	if err := libhoney.Init(honeyConfig); err != nil {
		return "", errors.Wrap(err, "libhoney.Init")
	}
	defer libhoney.Close()
	trace, err := GenerateDeploymentTrace(report)
	if err != nil {
		return "", errors.Wrap(err, "GenerateDeploymentTrace")
	}
	var sendErrs error
	for _, event := range trace.Spans {
		if err := event.Send(); err != nil {
			sendErrs = errors.Append(sendErrs, err)
		}
	}
	if sendErrs != nil {
		return "", errors.Wrap(err, "trace.Spans.Send")
	}
	if err := trace.Root.Send(); err != nil {
		return "", errors.Wrap(err, "trace.Root.Send")
	}
	traceURL, err := buildTraceURL(&honeyConfig, trace.ID, trace.Root.Timestamp.Unix())
	if err != nil {
		log.Println("warning: buildTraceURL: ", err.Error())
	} else {
		log.Println("trace: ", traceURL)
	}
	return traceURL, nil
}
