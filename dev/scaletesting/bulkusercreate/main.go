package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v41/github"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/lib/group"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type config struct {
	githubToken    string
	githubURL      string
	githubUser     string
	githubPassword string

	count  int
	action string
	resume string
	retry  int
}

var emailDomain = "scaletesting.sourcegraph.com"

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub user to authenticate with")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) password of the GitHub user to authenticate with")
	flag.IntVar(&cfg.count, "count", 100, "Amount of users to create")
	flag.IntVar(&cfg.retry, "retry", 5, "Retries count")
	flag.StringVar(&cfg.action, "action", "create", "Whether to 'create' or 'delete' users")
	flag.StringVar(&cfg.resume, "resume", "state.db", "Temporary state to use to resume progress if interrupted")

	flag.Parse()

	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	ctx := context.Background()
	// GHE cert has validity issues so hack around it for now
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.githubToken},
	))

	gh, err := github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
	if err != nil {
		writeFailure(out, "Failed to sign-in to GHE")
		log.Fatal(err)
	}

	if cfg.githubURL == "" {
		writeFailure(out, "-github.URL must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubToken == "" {
		writeFailure(out, "-github.token must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubUser == "" {
		writeFailure(out, "-github.login must be provided")
		flag.Usage()
		os.Exit(-1)
	}
	if cfg.githubPassword == "" {
		writeFailure(out, "-github.password must be provided")
		flag.Usage()
		os.Exit(-1)
	}

	state, err := newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

	var users []*user
	if users, err = state.load(); err != nil {
		log.Fatal(err)
	}

	if len(users) == 0 {
		if users, err = state.generate(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming jobs from %s", cfg.resume)
	}

	g := group.New().WithMaxConcurrency(100)
	for _, u := range users {
		currentUser := u
		if cfg.action == "create" {
			g.Go(func() {
				if currentUser.Created && currentUser.Failed == "" {
					return
				}
				existingUser, resp, grErr := gh.Users.Get(ctx, currentUser.Login)
				if grErr != nil && resp.StatusCode != 404 {
					writeFailure(out, "Failed to get user %s, reason: %s", currentUser.Login, grErr)
					return
				}
				grErr = nil
				if existingUser != nil {
					currentUser.Created = true
					currentUser.Failed = ""
					if grErr = state.saveUser(currentUser); grErr != nil {
						log.Fatal(grErr)
					}
					writeInfo(out, "user with login %s already exists", currentUser.Login)
					return
				}
				_, _, grErr = gh.Admin.CreateUser(ctx, currentUser.Login, currentUser.Email)

				if grErr != nil {
					writeFailure(out, "Failed to create user with login %s, reason: %s", currentUser.Login, grErr)
					currentUser.Failed = grErr.Error()
					if grErr = state.saveUser(currentUser); grErr != nil {
						log.Fatal(grErr)
					}
					return
				}
				currentUser.Created = true
				currentUser.Failed = ""
				if grErr = state.saveUser(currentUser); grErr != nil {
					log.Fatal(grErr)
				}
			})
		} else if cfg.action == "delete" {
			g.Go(func() {
				if !currentUser.Created {
					return
				}
				existingUser, resp, grErr := gh.Users.Get(ctx, currentUser.Login)
				if grErr != nil && resp.StatusCode != 404 {
					writeFailure(out, "Failed to get user %s, reason: %s", currentUser.Login, grErr)
				}
				grErr = nil
				if existingUser != nil {
					_, grErr = gh.Admin.DeleteUser(ctx, currentUser.Login)

					if grErr != nil {
						writeFailure(out, "Failed to delete user with login %s, reason: %s", currentUser.Login, grErr)
						currentUser.Failed = grErr.Error()
						if grErr = state.saveUser(currentUser); grErr != nil {
							log.Fatal(grErr)
						}
						return
					}
				}
				currentUser.Created = false
				currentUser.Failed = ""
				if grErr = state.saveUser(currentUser); grErr != nil {
					log.Fatal(grErr)
				}
			})
		}
	}
	g.Wait()

	all, err := state.countAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	completed, err := state.countCompletedUsers()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.action == "create" {
		writeSuccess(out, "Successfully added %d users (%d failures)", completed, all-completed)
	} else if cfg.action == "delete" {
		writeSuccess(out, "Successfully deleted %d users (%d failures)", all-completed, completed)
	}
}

func writeSuccess(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, format, a...))
}

func writeInfo(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("ℹ️", output.StyleYellow, format, a...))
}

func writeFailure(out *output.Output, format string, a ...any) {
	out.WriteLine(output.Linef("❌", output.StyleFailure, format, a...))
}
