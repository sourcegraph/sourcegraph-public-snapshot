package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/okay"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
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

var logger log.Logger

func main() {
	ctx := context.Background()
	sync := log.Init(log.Resource{Name: "deployment-notifier"})
	defer sync()
	logger = log.Scoped("main", "a script that checks for deployment notifications")

	flags := &Flags{}
	flags.Parse()
	if flags.Environment == "" {
		logger.Fatal("-environment must be specified: preprod or production.")
	}

	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: flags.GitHubToken},
	)))
	if flags.GitHubToken == "" {
		logger.Warn("using unauthenticated github client")
		ghc = github.NewClient(http.DefaultClient)
	}

	changedFiles, err := getChangedFiles()
	if err != nil {
		logger.Error("cannot get changed files", log.Error(err))
	}
	if len(changedFiles) == 0 {
		logger.Info("No relevant changes, skipping notifications and exiting normally.")
		return
	}

	manifestRevision, err := getRevision()
	if err != nil {
		logger.Fatal("cannot get revision", log.Error(err))
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
			logger.Info("No relevant changes, skipping notifications and exiting normally.")
			return
		}
		logger.Fatal("failed to generate report", log.Error(err))
	}

	// Tracing
	var traceURL string
	if flags.HoneycombToken != "" {
		traceURL, err = reportDeployTrace(report, flags.HoneycombToken, flags.DryRun)
		if err != nil {
			logger.Fatal("failed to generate a trace", log.Error(err))
		}
	}

	// Metrics, if token is empty, metrics will be logged at DEBUG level
	err = reportDeploymentMetrics(report, flags.OkayHQToken, flags.DryRun)
	if err != nil {
		logger.Fatal("failed to generate metrics", log.Error(err))
	}

	// Notifcations
	slc := slack.New(flags.SlackToken)
	teammates := team.NewTeammateResolver(ghc, slc)
	if !flags.DryRun {
		fmt.Println("Github\n---")
		for _, pr := range report.PullRequests {
			fmt.Println("-", pr.GetNumber())
		}
		out, err := renderComment(report, traceURL)
		if err != nil {
			logger.Fatal("can't render GitHub comment", log.Error(err))
		}
		fmt.Println(out)
		fmt.Println("Slack\n---")
		out, err = slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			logger.Fatal("can't render Slack post", log.Error(err))
		}
		fmt.Println(out)
	} else {
		out, err := slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			logger.Fatal("can't render Slack post", log.Error(err))
		}
		err = postSlackUpdate(flags.SlackAnnounceWebhook, out)
		if err != nil {
			logger.Fatal("can't post Slack update", log.Error(err))
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

	okayCli := okay.NewClient(http.DefaultClient, token)

	for _, pr := range report.PullRequests {
		elapsed := deployTime.Sub(pr.GetMergedAt())
		event := okay.Event{
			Name:        "deployment",
			Timestamp:   deployTime,
			GitHubLogin: pr.GetUser().GetLogin(),
			UniqueKey:   []string{"unique_key"},
			OkayURL:     pr.GetHTMLURL(),
			Properties: map[string]string{
				"environment":           report.Environment,
				"pull_request.number":   strconv.Itoa(pr.GetNumber()),
				"pull_request.title":    pr.GetTitle(),
				"pull_request.revision": pr.GetMergeCommitSHA(),
				"unique_key":            fmt.Sprintf("%s,%d,%s", report.Environment, pr.GetNumber(), strings.Join(report.Services, ",")),
			},
			Metrics: map[string]okay.Metric{
				"elapsed": {
					Type:  "durationMs",
					Value: float64(elapsed / time.Millisecond),
				},
			},
			Labels: report.Services,
		}

		err := okayCli.Push(&event)
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
		logger.Warn("failed to generate buildTraceURL", log.Error(err))
	} else {
		logger.Info("generated trace", log.String("trace", traceURL))
	}
	return traceURL, nil
}
