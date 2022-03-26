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
	BaseDir              string
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubToken, "github.token", os.Getenv("GITHUB_TOKEN"), "mandatory github token")
	flag.StringVar(&f.Environment, "environment", "", "Environment being deployed")
	flag.BoolVar(&f.DryRun, "dry", false, "Pretend to post notifications, printing to stdout instead")
	flag.StringVar(&f.SlackToken, "slack.token", "", "mandatory slack api token")
	flag.StringVar(&f.SlackAnnounceWebhook, "slack.webhook", "", "Slack Webhook URL to post the results on")
	flag.StringVar(&f.HoneycombToken, "honeycomb.token", "", "mandatory honeycomb api token")
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

	honeyConfig := libhoney.Config{
		APIKey:  flags.HoneycombToken,
		Dataset: "deploy-sourcegraph",
	}
	if flags.DryRun {
		honeyConfig.Transmission = &transmission.WriterSender{} // prints events to stdout instead
	}
	if err := libhoney.Init(honeyConfig); err != nil {
		log.Fatal(err)
	}
	events, err := GenerateDeploymentTrace(report)
	if err != nil {
		log.Fatal(err)
	}
	var sendErrs error
	for _, event := range events {
		if err := event.Send(); err != nil {
			sendErrs = errors.Append(sendErrs, err)
		}
	}
	if sendErrs != nil {
		log.Fatal(err)
	}
	libhoney.Close()

	slc := slack.New(flags.SlackToken)
	teammates := team.NewTeammateResolver(ghc, slc)

	if flags.DryRun {
		fmt.Println("Github\n---")
		for _, pr := range report.PullRequests {
			fmt.Println("-", pr.GetNumber())
		}
		out, err := renderComment(report)
		if err != nil {
			log.Fatalf("can't render GitHub comment %q", err)
		}
		fmt.Println(out)
		fmt.Println("Slack\n---")
		out, err = slackSummary(ctx, teammates, report)
		if err != nil {
			log.Fatalf("can't render Slack post %q", err)
		}
		fmt.Println(out)
	} else {
		out, err := slackSummary(ctx, teammates, report)
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
