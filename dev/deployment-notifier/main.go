package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

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
	flag.StringVar(&f.Environment, "environment", "production", "Environment being deployed")
	flag.BoolVar(&f.DryRun, "dry", false, "Pretend to post notifications, printing to stdout instead")
	flag.StringVar(&f.SlackToken, "slack.token", "", "mandatory slack api token")
	flag.StringVar(&f.SlackAnnounceWebhook, "slack.webhook", "", "Slack Webhook URL to post the results on")
	flag.StringVar(&f.HoneycombToken, "honeycomb.token", "", "mandatory honeycomb api token")
	flag.Parse()
}

var logger log.Logger

func main() {
	ctx := context.Background()
	liblog := log.Init(log.Resource{Name: "deployment-notifier"})
	defer liblog.Sync()
	logger = log.Scoped("main")

	flags := &Flags{}
	flags.Parse()
	if flags.Environment == "" {
		logger.Fatal("-environment must be specified. 'production' is the only valid option")
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
			logger.Fatal("can't render GitHub comment", log.Error(err))
		}
		fmt.Println(out)
		fmt.Println("Slack\n---")
		presenter, err := slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			logger.Fatal("can't render Slack post", log.Error(err))
		}

		fmt.Println(presenter.toString())
	} else {
		presenter, err := slackSummary(ctx, teammates, report, traceURL)
		if err != nil {
			logger.Fatal("can't render Slack post", log.Error(err))
		}
		err = postSlackUpdate(flags.SlackAnnounceWebhook, presenter)
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
