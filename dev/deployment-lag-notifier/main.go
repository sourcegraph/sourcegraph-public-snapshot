package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	_ "github.com/joho/godotenv/autoload"
)

// Flags are command-line arguments that configure the application behavior
// away from the defaults
type Flags struct {
	DryRun          bool
	Environment     string
	SlackWebhookURL string
	NumCommits      int
	AllowedAge      string
}

// Parse parses the CLI flags and stores them in a configuration struct
func (f *Flags) Parse() {
	flag.BoolVar(&f.DryRun, "dry-run", false, "Print to stdout instead of sending to Slack")
	flag.StringVar(&f.Environment, "env", Getenv("SG_ENVIRONMENT", "cloud"), "Environment to check against")
	flag.StringVar(&f.SlackWebhookURL, "slack-webhook-url", os.Getenv("SLACK_WEBHOOK_URL"), "Slack webhook URL to post to")
	flag.IntVar(&f.NumCommits, "num-commits", 30, "Number of commits to allow deployed version to drift from main")
	flag.StringVar(&f.AllowedAge, "allowed-age", "3h", "Duration (in time.Duration format) deployed version can differ from tip of main")
	flag.Parse()
}

// environments represent the currently available environment targets we may care about
var environments = map[string]string{
	"cloud": "https://sourcegraph.com",
	"k8s":   "https://k8s.sgdev.org",
}

// Getenv wraps os.Getenv but allows a default fallback value
func Getenv(env, def string) string {
	val, present := os.LookupEnv(env)
	if !present {
		val = def
	}
	return val
}

// getLiveVersion makes an HTTP GET request to a given Sourcegraph deployment version endpoint to get the running version
// information
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

	if resp.StatusCode != http.StatusOK {
		return version, errors.Newf("received non-200 status code %v: %s", resp.StatusCode, err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return version, err
	}

	return getCommitFromLiveVersion(string(body))
}

// getCommitFromLiveVersion strips the SHA from the live version string
func getCommitFromLiveVersion(liveVersion string) (string, error) {
	// Response is in format taggedversion-build_date_hash
	parts := strings.Split(liveVersion, "_")

	if len(parts) != 3 {
		return liveVersion, errors.Newf("unknown version format %s", liveVersion)
	}

	version := parts[2]

	// New version format for continuous builds includes the tagged version, which needs to be stripped
	parts = strings.Split(version, "-")
	if len(parts) != 2 {
		return version, errors.Newf("Unable to get SHA from version with format %s", version)
	}

	sha := parts[1]

	return sha, nil
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

// commitTooOld compares the age of the current commit to the age of the tip of main
// and if the threshold (set by flags.CommitAge) is exceeded, return true
func commitTooOld(curr, tip Commit, threshold time.Duration) (bool, time.Duration) {
	drift := tip.Date.Sub(curr.Date)
	if drift > threshold {
		return true, drift
	}
	return false, drift
}

func main() {
	flags := &Flags{}
	flags.Parse()

	client := http.Client{}

	url, ok := environments[flags.Environment]
	if !ok {
		var s string
		for k, v := range environments {
			s += fmt.Sprintf("\t%s: %s\n", k, v)
		}
		log.Fatalf("Environment \"%s\" not found. Valid options are: \n%s\n", flags.Environment, s)
	}

	allowedAge, err := time.ParseDuration(flags.AllowedAge)
	if err != nil {
		log.Fatal(err)
	}

	version, err := getLiveVersion(&client, url)
	if err != nil {
		log.Fatal(err)
	}

	commitLog, err := getCommitLog(&client, flags.NumCommits)
	if err != nil {
		log.Fatal(err)
	}

	currentCommit, err := getCommit(&client, version)
	if err != nil {
		log.Fatal(err)
	}

	slack := NewSlackClient(flags.SlackWebhookURL)

	inAllowedNumCommits := checkForCommit(version, commitLog)

	timeExceeded, drift := commitTooOld(currentCommit, commitLog[0], allowedAge)

	// Always at least print locally when running a dry-run
	if !inAllowedNumCommits || timeExceeded || flags.DryRun {

		td := TemplateData{
			VersionAge:       time.Now().Sub(currentCommit.Date).Truncate(time.Second).String(),
			Version:          version,
			Environment:      flags.Environment,
			CommitTooOld:     timeExceeded,
			Threshold:        allowedAge.String(),
			Drift:            drift.String(),
			InAllowedCommits: inAllowedNumCommits,
			NumCommits:       flags.NumCommits,
		}

		msg, err := createMessage(td)
		if !flags.DryRun {
			err = slack.PostMessage(msg)
			if err != nil {
				log.Fatal(err)
			}
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Cloud is not current!")
		fmt.Println(msg.String())
	}

	log.Printf("Now: %s\n", time.Now().String())
	log.Printf("%s: %s\n", flags.Environment, currentCommit.Date.String())
	log.Printf("main: %s\n", commitLog[0].Date.String())

}
