package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"time"

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

	userCount int
	teamCount int
	orgCount  int
	orgAdmin  string
	action    string
	resume    string
	retry     int
}

var emailDomain = "scaletesting.sourcegraph.com"

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub user to authenticate with")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) password of the GitHub user to authenticate with")
	flag.IntVar(&cfg.userCount, "user.count", 100, "Amount of users to create")
	flag.IntVar(&cfg.orgCount, "org.count", 10, "Amount of orgs to create")
	flag.StringVar(&cfg.orgAdmin, "org.admin", "", "Login of admin of orgs")

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
	if cfg.orgAdmin == "" {
		writeFailure(out, "-org.admin must be provided")
		flag.Usage()
		os.Exit(-1)
	}

	// ratio from https://github.com/sourcegraph/sourcegraph/issues/43052
	cfg.teamCount = (cfg.userCount / 3) * 2

	state, err := newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

	// load or generate users
	var users []*user
	if users, err = state.loadUsers(); err != nil {
		log.Fatal(err)
	}

	if len(users) == 0 {
		if users, err = state.generateUsers(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated user jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming user jobs from %s", cfg.resume)
	}

	// load or generate orgs
	var orgs []*org
	if orgs, err = state.loadOrgs(); err != nil {
		log.Fatal(err)
	}

	if len(orgs) == 0 {
		if orgs, err = state.generateOrgs(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated org jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming org jobs from %s", cfg.resume)
	}

	// load or generate teams
	var teams []*team
	if teams, err = state.loadTeams(); err != nil {
		log.Fatal(err)
	}

	if len(teams) == 0 {
		if teams, err = state.generateTeams(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated team jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming team jobs from %s", cfg.resume)
	}

	bars := []output.ProgressBar{
		{Label: "Creating users", Max: float64(cfg.userCount)},
		{Label: "Creating orgs", Max: float64(cfg.orgCount)},
		{Label: "Creating teams", Max: float64(cfg.teamCount)},
	}
	progress := out.Progress(bars, nil)
	var usersDone int64
	var orgsDone int64
	var teamsDone int64

	g := group.New().WithMaxConcurrency(1000)
	for _, u := range users {
		currentUser := u
		if cfg.action == "create" {
			g.Go(func() {
				if currentUser.Created && currentUser.Failed == "" {
					atomic.AddInt64(&usersDone, 1)
					progress.SetValue(0, float64(usersDone))
					return
				}
				existingUser, resp, uErr := gh.Users.Get(ctx, currentUser.Login)
				if uErr != nil && resp.StatusCode != 404 {
					writeFailure(out, "Failed to get user %s, reason: %s", currentUser.Login, uErr)
					return
				}
				uErr = nil
				if existingUser != nil {
					currentUser.Created = true
					currentUser.Failed = ""
					if uErr = state.saveUser(currentUser); uErr != nil {
						log.Fatal(uErr)
					}
					writeInfo(out, "user with login %s already exists", currentUser.Login)
					atomic.AddInt64(&usersDone, 1)
					progress.SetValue(0, float64(usersDone))
					return
				}
				_, _, uErr = gh.Admin.CreateUser(ctx, currentUser.Login, currentUser.Email)

				if uErr != nil {
					writeFailure(out, "Failed to create user with login %s, reason: %s", currentUser.Login, uErr)
					currentUser.Failed = uErr.Error()
					if uErr = state.saveUser(currentUser); uErr != nil {
						log.Fatal(uErr)
					}
					return
				}
				currentUser.Created = true
				currentUser.Failed = ""
				if uErr = state.saveUser(currentUser); uErr != nil {
					log.Fatal(uErr)
				}
				atomic.AddInt64(&usersDone, 1)
				progress.SetValue(0, float64(usersDone))
			})
		} else if cfg.action == "delete" {
			g.Go(func() {
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

	for _, o := range orgs {
		currentOrg := o
		g.Go(func() {
			if currentOrg.Created && currentOrg.Failed == "" {
				atomic.AddInt64(&orgsDone, 1)
				progress.SetValue(0, float64(orgsDone))
				return
			}
			existingOrg, resp, oErr := gh.Organizations.Get(ctx, currentOrg.Login)
			if oErr != nil && resp.StatusCode != 404 {
				writeFailure(out, "Failed to get org %s, reason: %s", currentOrg.Login, oErr)
				return
			}
			oErr = nil
			if existingOrg != nil {
				currentOrg.Created = true
				currentOrg.Failed = ""
				if oErr = state.saveOrg(currentOrg); oErr != nil {
					log.Fatal(oErr)
				}
				writeInfo(out, "org with login %s already exists", currentOrg.Login)
				atomic.AddInt64(&orgsDone, 1)
				progress.SetValue(0, float64(orgsDone))
				return
			}
			_, _, oErr = gh.Admin.CreateOrg(ctx, &github.Organization{Login: &currentOrg.Login}, cfg.orgAdmin)

			if oErr != nil {
				writeFailure(out, "Failed to create org with login %s, reason: %s", currentOrg.Login, oErr)
				currentOrg.Failed = oErr.Error()
				if oErr = state.saveOrg(currentOrg); oErr != nil {
					log.Fatal(oErr)
				}
				return
			}
			currentOrg.Created = true
			currentOrg.Failed = ""
			if oErr = state.saveOrg(currentOrg); oErr != nil {
				log.Fatal(oErr)
			}
			atomic.AddInt64(&orgsDone, 1)
			progress.SetValue(0, float64(orgsDone))
		})
	}

	for _, t := range teams {
		currentTeam := t
		g.Go(func() {
			if currentTeam.Created && currentTeam.Failed == "" {
				atomic.AddInt64(&teamsDone, 1)
				progress.SetValue(0, float64(teamsDone))
				return
			}
			existingTeam, resp, tErr := gh.Teams.GetTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)
			if tErr != nil && resp.StatusCode != 404 {
				writeFailure(out, "Failed to get team %s, reason: %s", currentTeam.Name, tErr)
				return
			}
			tErr = nil
			if existingTeam != nil {
				currentTeam.Created = true
				currentTeam.Failed = ""

				teamMembers, resp, tErr := gh.Teams.ListTeamMembersBySlug(ctx, currentTeam.Org, currentTeam.Name, nil)
				if tErr != nil && resp.StatusCode != 404 {
					writeFailure(out, "Failed to get members of team %s, reason: %s", currentTeam.Name, tErr)
					return
				}
				// Members have been randomly selected before, so save to state
				if len(teamMembers) != 0 {
					currentTeam.HasMembers = true
				}

				if tErr = state.saveTeam(currentTeam); tErr != nil {
					log.Fatal(tErr)
				}
				writeInfo(out, "team with name %s already exists", currentTeam.Name)
				atomic.AddInt64(&teamsDone, 1)
				progress.SetValue(0, float64(teamsDone))
				return
			}

			// users need to be member of the team's parent org to join the team
			userState := "active"
			userRole := "member"
			var randomTeamMemberLogins []string
			// get 50 random users to be member of this team (as per https://sourcegraph.slack.com/archives/C03GMM0F2BA/p1665658367720609)
			rand.Seed(time.Now().Unix())
			for i := 0; i < 1; i++ {
				randomUser := users[rand.Intn(len(users))]
				_, _, tErr = gh.Organizations.EditOrgMembership(ctx, randomUser.Login, t.Org, &github.Membership{
					State:        &userState,
					Role:         &userRole,
					Organization: &github.Organization{Login: &t.Org},
					User:         &github.User{Login: &randomUser.Login},
				})
				if tErr != nil {
					currentTeam.Failed = tErr.Error()
					currentTeam.HasMembers = false
					if tErr = state.saveTeam(currentTeam); tErr != nil {
						log.Fatal(tErr)
					}
				}
				randomTeamMemberLogins = append(randomTeamMemberLogins, randomUser.Login)
			}

			team, resp, tErr := gh.Teams.CreateTeam(ctx, currentTeam.Org, github.NewTeam{
				Name:        currentTeam.Name,
				Maintainers: randomTeamMemberLogins,
			})
			team.GetID()

			if tErr != nil {
				writeFailure(out, "Failed to create team with name %s, reason: %s", currentTeam.Name, tErr)
				currentTeam.Failed = tErr.Error()
				if tErr = state.saveTeam(currentTeam); tErr != nil {
					log.Fatal(tErr)
				}
				return
			}
			currentTeam.Created = true
			currentTeam.Failed = ""
			if tErr = state.saveTeam(currentTeam); tErr != nil {
				log.Fatal(tErr)
			}
			atomic.AddInt64(&teamsDone, 1)
			progress.SetValue(0, float64(teamsDone))
		})
	}
	g.Wait()

	allUsers, err := state.countAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	completedUsers, err := state.countCompletedUsers()
	if err != nil {
		log.Fatal(err)
	}
	allOrgs, err := state.countAllOrgs()
	if err != nil {
		log.Fatal(err)
	}
	completedOrgs, err := state.countCompletedOrgs()
	if err != nil {
		log.Fatal(err)
	}
	allTeams, err := state.countAllTeams()
	if err != nil {
		log.Fatal(err)
	}
	completedTeams, err := state.countCompletedTeams()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.action == "create" {
		writeSuccess(out, "Successfully added %d users (%d failures)", completedUsers, allUsers-completedUsers)
		writeSuccess(out, "Successfully added %d orgs (%d failures)", completedOrgs, allOrgs-completedOrgs)
		writeSuccess(out, "Successfully added %d teams (%d failures)", completedTeams, allTeams-completedTeams)
	} else if cfg.action == "delete" {
		writeSuccess(out, "Successfully deleted %d users (%d failures)", allUsers-completedUsers, completedUsers)
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
