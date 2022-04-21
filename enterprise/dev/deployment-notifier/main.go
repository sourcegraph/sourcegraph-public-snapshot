package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
		elapsed := deployTime.Sub(pr.GetMergedAt())
		event := OkayEvent{
			Name:        "deployment",
			Timestamp:   deployTime,
			GitHubLogin: pr.GetUser().GetLogin(),
			UniqueKey:   []string{"environment", "pull_request.number", "services"},
			Properties: map[string]interface{}{
				"environment":           report.Environment,
				"pull_request.number":   pr.GetNumber(),
				"pull_request.title":    pr.GetTitle(),
				"pull_request.revision": pr.GetMergeCommitSHA(),
				"pull_request.url":      pr.GetHTMLURL(),
				"services":              report.Services,
			},
			Metrics: map[string]OkayMetric{
				"elapsed": {
					Type:  "durationMs",
					Value: float64(elapsed / time.Millisecond),
				},
			},
		}

		err := okayCli.Push("qa.deployment", &event)
		if err != nil {
			return err
		}
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
