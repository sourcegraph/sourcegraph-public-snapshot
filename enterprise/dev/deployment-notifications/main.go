package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

type Flags struct {
	SourcegraphCommit      string
	MockLiveCommit         string
	GitHubToken            string
	Pretend                bool
	GuessSourcegraphCommit bool
	Environment            string
}

func (f *Flags) Parse() {
	flag.StringVar(&f.GitHubToken, "github.token", os.Getenv("GITHUB_TOKEN"), "mandatory github token")
	flag.StringVar(&f.SourcegraphCommit, "sourcegraph.commit", "", "Sourcegraph commit being deployed")
	flag.StringVar(&f.Environment, "environment", "", "Environment being deployed")
	flag.StringVar(&f.MockLiveCommit, "mock.live-commit", "", "Use this commit instead of requesting the commit deployed on the target environment")
	flag.BoolVar(&f.Pretend, "pretend", false, "Pretend to post notifications, printing to stdout instead")
	flag.BoolVar(&f.GuessSourcegraphCommit, "sourcegraph.guess-commit", false, "Attempt at deducting the deployed commit from the changes in the diff")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	flags := &Flags{}
	flags.Parse()
	if flags.Environment == "" {
		log.Fatalf("-enviroment must be specified: preprod or production.")
	}
	if flags.SourcegraphCommit == "" && !flags.GuessSourcegraphCommit {
		log.Fatalf("-sourcegraph.commit must be specified.")
	}

	if flags.GuessSourcegraphCommit {
		commit, err := guessSourcegraphCommit()
		if err != nil || commit == "" {
			log.Fatalf("could not guess commit from changes, %q", err)
		}
		flags.SourcegraphCommit = commit
	}

	ghc := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: flags.GitHubToken},
	)))

	changedFiles, err := getChangedFiles()
	if err != nil {
		log.Fatal(err)
	}

	var vr VersionRequester
	if flags.MockLiveCommit != "" {
		vr = NewMockVersionRequester(flags.MockLiveCommit, nil)
	} else {
		NewAPIVersionRequester(flags.Environment)
	}

	dn := NewDeploymentNotifier(
		ghc,
		vr,
		flags.SourcegraphCommit,
		changedFiles,
	)

	report, err := dn.Report(ctx)
	if err != nil {
		if errors.Is(err, ErrAlreadyDeployed) {
			fmt.Println(":warning: Already deployed, skipping notifications and exiting normally.")
			return
		}
		log.Fatal(err)
	}

	if flags.Pretend {
		out, _ := renderComment(report)
		fmt.Println(out)
		for _, pr := range report.PullRequests {
			fmt.Println(pr.GetNumber())
		}
	} else {
		panic("not implemented")
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

func GitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(),
		// Don't use the system wide git config.
		"GIT_CONFIG_NOSYSTEM=1",
		// And also not any other, because they can mess up output, change defaults, .. which can do unexpected things.
		"GIT_CONFIG=/dev/null")

	return InRoot(cmd)
}

func InRoot(cmd *exec.Cmd) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), errors.Wrapf(err, "'%s' failed: %s", strings.Join(cmd.Args, " "), out)
	}

	return string(out), nil
}
