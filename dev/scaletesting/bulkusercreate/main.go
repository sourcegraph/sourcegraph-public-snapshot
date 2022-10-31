package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"math"
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

var (
	emailDomain = "scaletesting.sourcegraph.com"

	out      *output.Output
	store    *state
	gh       *github.Client
	progress output.Progress
)

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

	ctx := context.Background()
	out = output.NewOutput(os.Stdout, output.OutputOpts{})

	// GHE cert has validity issues so hack around it for now
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.githubToken},
	))

	var err error
	gh, err = github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
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

	store, err = newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

	// load or generate users
	var users []*user
	if users, err = store.loadUsers(); err != nil {
		log.Fatal(err)
	}

	if len(users) == 0 {
		if users, err = store.generateUsers(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated user jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming user jobs from %s", cfg.resume)
	}

	// load or generate orgs
	var orgs []*org
	if orgs, err = store.loadOrgs(); err != nil {
		log.Fatal(err)
	}

	if len(orgs) == 0 {
		if orgs, err = store.generateOrgs(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated org jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming org jobs from %s", cfg.resume)
	}

	// load or generate teams
	var teams []*team
	if teams, err = store.loadTeams(); err != nil {
		log.Fatal(err)
	}

	if len(teams) == 0 {
		if teams, err = store.generateTeams(cfg); err != nil {
			log.Fatal(err)
		}
		writeSuccess(out, "generated team jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming team jobs from %s", cfg.resume)
	}

	start := time.Now()

	g := group.New().WithMaxConcurrency(1000)
	if cfg.action == "create" {
		bars := []output.ProgressBar{
			{Label: "Creating orgs", Max: float64(cfg.orgCount)},
			{Label: "Creating teams", Max: float64(cfg.teamCount)},
			{Label: "Creating users", Max: float64(cfg.userCount)},
			{Label: "Adding users to teams", Max: float64(cfg.teamCount * 50)},
		}
		progress = out.Progress(bars, nil)
		var usersDone int64
		var orgsDone int64
		var teamsDone int64
		var membershipsDone int64

		for _, o := range orgs {
			currentOrg := o
			g.Go(func() {
				executeCreateOrg(ctx, currentOrg, cfg.orgAdmin, &orgsDone)
			})
		}
		g.Wait()

		for _, t := range teams {
			currentTeam := t
			g.Go(func() {
				executeCreateTeam(ctx, currentTeam, &teamsDone)
			})
		}
		g.Wait()

		for _, u := range users {
			currentUser := u
			g.Go(func() {
				executeCreateUser(ctx, currentUser, &usersDone)
			})
		}
		g.Wait()

		totalMemberships := len(teams) * 50
		membershipsPerUser := int(math.Ceil(float64(totalMemberships) / float64(cfg.userCount)))
		teamsToSkip := int(math.Ceil(float64(cfg.teamCount) / (float64(totalMemberships) / float64(cfg.userCount))))

		for i, u := range users {
			currentUser := u
			currentIter := i

			g.Go(func() {
				executeCreateTeamMembershipsForUser(
					ctx,
					&teamMembershipOpts{
						currentUser:        currentUser,
						teams:              teams,
						membershipsPerUser: membershipsPerUser,
						teamIndex:          currentIter,
						teamIncrement:      teamsToSkip,
					},
					&membershipsDone)
			})
		}
		g.Wait()
	}

	if cfg.action == "delete" {
		for _, u := range users {
			currentUser := u
			g.Go(func() {
				executeDeleteUser(ctx, currentUser)
			})
		}

		for _, t := range teams {
			currentTeam := t
			g.Go(func() {
				executeDeleteTeam(ctx, currentTeam)
			})
		}
	}
	g.Wait()

	end := time.Now()
	writeInfo(out, "Started at %s, finished at %s", start.String(), end.String())

	allUsers, err := store.countAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	completedUsers, err := store.countCompletedUsers()
	if err != nil {
		log.Fatal(err)
	}
	allOrgs, err := store.countAllOrgs()
	if err != nil {
		log.Fatal(err)
	}
	completedOrgs, err := store.countCompletedOrgs()
	if err != nil {
		log.Fatal(err)
	}
	allTeams, err := store.countAllTeams()
	if err != nil {
		log.Fatal(err)
	}
	completedTeams, err := store.countCompletedTeams()
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

func executeDeleteTeam(ctx context.Context, currentTeam *team) {
	existingTeam, resp, grErr := gh.Teams.GetTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)

	if grErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "Failed to get team %s, reason: %s", currentTeam.Name, grErr)
	}

	grErr = nil
	if existingTeam != nil {
		_, grErr = gh.Teams.DeleteTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)
		if grErr != nil {
			writeFailure(out, "Failed to delete team %s, reason: %s", currentTeam.Name, grErr)
			currentTeam.Failed = grErr.Error()
			if grErr = store.saveTeam(currentTeam); grErr != nil {
				log.Fatal(grErr)
			}
			return
		}
	}

	currentTeam.Created = false
	currentTeam.Failed = ""
	currentTeam.TotalMembers = 0
	if grErr = store.saveTeam(currentTeam); grErr != nil {
		log.Fatal(grErr)
	}

	writeSuccess(out, "Deleted team %s", currentTeam.Name)
}

func executeDeleteUser(ctx context.Context, currentUser *user) {
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
			if grErr = store.saveUser(currentUser); grErr != nil {
				log.Fatal(grErr)
			}
			return
		}
	}

	currentUser.Created = false
	currentUser.Failed = ""
	if grErr = store.saveUser(currentUser); grErr != nil {
		log.Fatal(grErr)
	}

	writeSuccess(out, "Deleted user %s", currentUser.Login)
}

type teamMembershipOpts struct {
	currentUser *user
	teams       []*team

	membershipsPerUser int
	teamIndex          int
	teamIncrement      int
}

func executeCreateTeamMembershipsForUser(ctx context.Context, opts *teamMembershipOpts, membershipsDone *int64) {
	// users need to be member of the team's parent org to join the team
	userState := "active"
	userRole := "member"

	for j := 0; j < opts.membershipsPerUser; j++ {
		index := (opts.teamIndex + (j * opts.teamIncrement)) % len(opts.teams)
		candidateTeam := opts.teams[index]

		if candidateTeam.TotalMembers >= 50 {
			continue
		}

		// add user to team's parent org first
		_, _, mErr := gh.Organizations.EditOrgMembership(ctx, opts.currentUser.Login, candidateTeam.Org, &github.Membership{
			State:        &userState,
			Role:         &userRole,
			Organization: &github.Organization{Login: &candidateTeam.Org},
			User:         &github.User{Login: &opts.currentUser.Login},
		})

		if mErr != nil {
			writeFailure(out, "Failed to add user %s to organization %s, reason: %s", opts.currentUser.Login, candidateTeam.Org, mErr)
			candidateTeam.Failed = mErr.Error()
			if mErr = store.saveTeam(candidateTeam); mErr != nil {
				log.Fatal(mErr)
			}
			continue
		}

		// this is an idempotent operation so no need to check existing membership
		_, _, mErr = gh.Teams.AddTeamMembershipBySlug(ctx, candidateTeam.Org, candidateTeam.Name, opts.currentUser.Login, nil)

		if mErr != nil {
			writeFailure(out, "Failed to add user %s to team %s, reason: %s", opts.currentUser, candidateTeam.Name, mErr)
			candidateTeam.Failed = mErr.Error()
			if mErr = store.saveTeam(candidateTeam); mErr != nil {
				log.Fatal(mErr)
			}
			continue
		}

		candidateTeam.TotalMembers += 1
		atomic.AddInt64(membershipsDone, 1)
		progress.SetValue(3, float64(*membershipsDone))

		if mErr = store.saveTeam(candidateTeam); mErr != nil {
			log.Fatal(mErr)
		}

		//writeSuccess(out, "Added member %s to team %s", currentUser.Login, candidateTeam.Name)
	}
}

func executeCreateUser(ctx context.Context, currentUser *user, usersDone *int64) {
	if currentUser.Created && currentUser.Failed == "" {
		atomic.AddInt64(usersDone, 1)
		progress.SetValue(2, float64(*usersDone))
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
		if uErr = store.saveUser(currentUser); uErr != nil {
			log.Fatal(uErr)
		}
		writeInfo(out, "user with login %s already exists", currentUser.Login)
		atomic.AddInt64(usersDone, 1)
		progress.SetValue(2, float64(*usersDone))
		return
	}

	_, _, uErr = gh.Admin.CreateUser(ctx, currentUser.Login, currentUser.Email)
	if uErr != nil {
		writeFailure(out, "Failed to create user with login %s, reason: %s", currentUser.Login, uErr)
		currentUser.Failed = uErr.Error()
		if uErr = store.saveUser(currentUser); uErr != nil {
			log.Fatal(uErr)
		}
		return
	}

	currentUser.Created = true
	currentUser.Failed = ""
	atomic.AddInt64(usersDone, 1)
	progress.SetValue(2, float64(*usersDone))
	if uErr = store.saveUser(currentUser); uErr != nil {
		log.Fatal(uErr)
	}

	//writeSuccess(out, "Created user with login %s", currentUser.Login)
}

func executeCreateTeam(ctx context.Context, currentTeam *team, teamsDone *int64) {
	if currentTeam.Created && currentTeam.Failed == "" {
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))
		return
	}

	existingTeam, resp, tErr := gh.Teams.GetTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)

	if tErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "failed to get team with name %s, reason: %s", currentTeam.Name, tErr)
		return
	}

	tErr = nil
	if existingTeam != nil {
		currentTeam.Created = true
		currentTeam.Failed = ""
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))

		if tErr = store.saveTeam(currentTeam); tErr != nil {
			log.Fatal(tErr)
		}
	} else {
		// Create the team if not exists
		if _, _, tErr = gh.Teams.CreateTeam(ctx, currentTeam.Org, github.NewTeam{Name: currentTeam.Name}); tErr != nil {
			writeFailure(out, "Failed to create team with name %s, reason: %s", currentTeam.Name, tErr)
			currentTeam.Failed = tErr.Error()
			if tErr = store.saveTeam(currentTeam); tErr != nil {
				log.Fatal(tErr)
			}
		}

		currentTeam.Created = true
		currentTeam.Failed = ""
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))

		if tErr = store.saveTeam(currentTeam); tErr != nil {
			log.Fatal(tErr)
		}
	}
}

func executeCreateOrg(ctx context.Context, currentOrg *org, orgAdmin string, orgsDone *int64) {
	if currentOrg.Created && currentOrg.Failed == "" {
		atomic.AddInt64(orgsDone, 1)
		progress.SetValue(0, float64(*orgsDone))
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
		atomic.AddInt64(orgsDone, 1)
		progress.SetValue(0, float64(*orgsDone))

		if oErr = store.saveOrg(currentOrg); oErr != nil {
			log.Fatal(oErr)
		}
		return
	}

	_, _, oErr = gh.Admin.CreateOrg(ctx, &github.Organization{Login: &currentOrg.Login}, orgAdmin)

	if oErr != nil {
		writeFailure(out, "Failed to create org with login %s, reason: %s", currentOrg.Login, oErr)
		currentOrg.Failed = oErr.Error()
		if oErr = store.saveOrg(currentOrg); oErr != nil {
			log.Fatal(oErr)
		}
		return
	}

	atomic.AddInt64(orgsDone, 1)
	progress.SetValue(0, float64(*orgsDone))

	currentOrg.Created = true
	currentOrg.Failed = ""
	if oErr = store.saveOrg(currentOrg); oErr != nil {
		log.Fatal(oErr)
	}

	//writeSuccess(out, "Created org with login %s", currentOrg.Login)
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
