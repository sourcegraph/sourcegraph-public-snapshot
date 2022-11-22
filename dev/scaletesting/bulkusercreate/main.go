package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
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

	userCount      int
	teamCount      int
	orgCount       int
	orgAdmin       string
	action         string
	resume         string
	retry          int
	generateTokens bool
}

var (
	emailDomain = "scaletesting.sourcegraph.com"

	out      *output.Output
	store    *state
	gh       *github.Client
	progress output.Progress
)

type userToken struct {
	login string
	token string
}

func main() {
	var cfg config

	flag.StringVar(&cfg.githubToken, "github.token", "", "(required) GitHub personal access token for the destination GHE instance")
	flag.StringVar(&cfg.githubURL, "github.url", "", "(required) GitHub base URL for the destination GHE instance")
	flag.StringVar(&cfg.githubUser, "github.login", "", "(required) GitHub user to authenticate with")
	flag.StringVar(&cfg.githubPassword, "github.password", "", "(required) password of the GitHub user to authenticate with")
	flag.IntVar(&cfg.userCount, "user.count", 100, "Amount of users to create or delete")
	flag.IntVar(&cfg.teamCount, "team.count", 20, "Amount of teams to create or delete")
	flag.IntVar(&cfg.orgCount, "org.count", 10, "Amount of orgs to create or delete")
	flag.StringVar(&cfg.orgAdmin, "org.admin", "", "Login of admin of orgs")

	flag.IntVar(&cfg.retry, "retry", 5, "Retries count")
	flag.StringVar(&cfg.action, "action", "create", "Whether to 'create' or 'delete' users")
	flag.StringVar(&cfg.resume, "resume", "state.db", "Temporary state to use to resume progress if interrupted")
	flag.BoolVar(&cfg.generateTokens, "generateTokens", false, "Generate new impersonation OAuth tokens for users")

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

	store, err = newState(cfg.resume)
	if err != nil {
		log.Fatal(err)
	}

	// load or generate orgs (used by both create and delete actions)
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

	start := time.Now()

	g := group.New().WithMaxConcurrency(1000)
	if cfg.action == "create" {
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

		bars := []output.ProgressBar{
			{Label: "Creating orgs", Max: float64(cfg.orgCount)},
			{Label: "Creating teams", Max: float64(cfg.teamCount)},
			{Label: "Creating users", Max: float64(cfg.userCount)},
			{Label: "Adding users to teams", Max: float64(cfg.userCount)},
		}
		if cfg.generateTokens {
			bars = append(bars, output.ProgressBar{Label: "Generating OAuth tokens", Max: float64(cfg.userCount)})
		}
		progress = out.Progress(bars, nil)
		var usersDone int64
		var orgsDone int64
		var teamsDone int64
		var tokensDone int64
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

		//totalMemberships := len(teams) * 8
		//membershipsPerUser := int(math.Ceil(float64(totalMemberships) / float64(cfg.userCount)))
		membershipsPerTeam := 8
		//teamsToSkip := int(math.Ceil(float64(cfg.teamCount) / float64(membershipsPerUser)))

		//for i, u := range users {
		//	currentUser := u
		//	currentIter := i
		//
		//	g.Go(func() {
		//		executeCreateTeamMembershipsForUser(
		//			ctx,
		//			&teamMembershipOpts{
		//				currentUser:        currentUser,
		//				teams:              teams,
		//				membershipsPerUser: membershipsPerUser,
		//				userIndex:          currentIter,
		//				//teamIncrement:      teamsToSkip,
		//			},
		//			&membershipsDone)
		//	})
		//}

		g2 := group.New().WithMaxConcurrency(100)
		for i, t := range teams {
			currentTeam := t
			currentIter := i
			var usersToAssign []*user

			for j := currentIter * membershipsPerTeam; j < ((currentIter + 1) * membershipsPerTeam); j++ {
				usersToAssign = append(usersToAssign, users[j])
			}

			g2.Go(func() {
				executeCreateTeamMembershipsForTeam(ctx, currentTeam, usersToAssign, &membershipsDone)
			})
		}
		g2.Wait()

		var repos []*repo
		if repos, err = store.loadRepos(); err != nil {
			log.Fatal(err)
		}

		if len(repos) == 0 {
			remoteRepos := getGitHubRepos(ctx)

			if err := store.insertRepos(remoteRepos); err != nil {
				log.Fatal(err)
			}
			writeSuccess(out, "Fetched %d private repos and stored in state", len(remoteRepos))
		} else {
			writeSuccess(out, "resuming repo jobs from %s", cfg.resume)
		}

		if cfg.generateTokens {
			tg := group.NewWithResults[userToken]().WithMaxConcurrency(1000)
			for _, u := range users {
				currentU := u
				tg.Go(func() userToken {
					token := executeCreateUserImpersonationToken(ctx, currentU)
					atomic.AddInt64(&tokensDone, 1)
					progress.SetValue(4, float64(tokensDone))
					return userToken{
						login: currentU.Login,
						token: token,
					}
				})
			}

			csvFile, err := os.Create("users.csv")
			if err != nil {
				log.Fatalf("Failed creating csv: %s", err)
			}
			defer csvFile.Close()
			csvwriter := csv.NewWriter(csvFile)
			defer csvwriter.Flush()
			_ = csvwriter.Write([]string{"login", "token"})
			pairs := tg.Wait()
			sort.Slice(pairs, func(i, j int) bool {
				comp := strings.Compare(pairs[i].login, pairs[j].login)
				return comp == -1
			})
			for _, pair := range pairs {
				if err = csvwriter.Write([]string{pair.login, pair.token}); err != nil {
					log.Fatalln("error writing pair to file", err)
				}
			}
		}

		g.Wait()

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

		writeSuccess(out, "Successfully added %d users (%d failures)", completedUsers, allUsers-completedUsers)
		writeSuccess(out, "Successfully added %d orgs (%d failures)", completedOrgs, allOrgs-completedOrgs)
		writeSuccess(out, "Successfully added %d teams (%d failures)", completedTeams, allTeams-completedTeams)
	}

	if cfg.action == "delete" {
		localOrgs, err := store.loadOrgs()
		if err != nil {
			log.Fatal("Failed to load orgs", err)
		}

		if len(localOrgs) == 0 {
			// Fetch orgs currently on instance due to lost state
			remoteOrgs := getGitHubOrgs(ctx)

			writeInfo(out, "Storing %d orgs in state", len(remoteOrgs))
			for _, o := range remoteOrgs {
				if strings.HasPrefix(*o.Name, "org-") {
					o := &org{
						Login:   *o.Login,
						Admin:   cfg.orgAdmin,
						Failed:  "",
						Created: true,
					}
					if err := store.saveOrg(o); err != nil {
						log.Fatal(err)
					}
					localOrgs = append(localOrgs, o)
				}
			}
		}

		localUsers, err := store.loadUsers()
		if err != nil {
			log.Fatal("Failed to load users", err)
		}

		localTeams, err := store.loadTeams()
		if err != nil {
			log.Fatal("Failed to load teams", err)
		}

		if len(localUsers) == 0 {
			// Fetch users currently on instance due to lost state
			remoteUsers := getGitHubUsers(ctx)

			writeInfo(out, "Storing %d users in state", len(remoteUsers))
			for _, u := range remoteUsers {
				if strings.HasPrefix(*u.Login, "user-") {
					u := &user{
						Login:   *u.Login,
						Email:   fmt.Sprintf("%s@%s", *u.Login, emailDomain),
						Failed:  "",
						Created: true,
					}
					if err := store.saveUser(u); err != nil {
						log.Fatal(err)
					}
					localUsers = append(localUsers, u)
				}
			}
		}

		if len(localTeams) == 0 {
			// Fetch teams currently on instance due to lost state
			remoteTeams := getGitHubTeams(ctx, localOrgs)

			writeInfo(out, "Storing %d teams in state", len(remoteTeams))
			for _, t := range remoteTeams {
				if strings.HasPrefix(*t.Name, "team-") {
					t := &team{
						Name:         *t.Name,
						Org:          *t.Organization.Login,
						Failed:       "",
						Created:      true,
						TotalMembers: 0, //not important for deleting but subsequent use of state will be problematic
					}
					if err := store.saveTeam(t); err != nil {
						log.Fatal(err)
					}
					localTeams = append(localTeams, t)
				}
			}
		}

		// delete users from instance
		usersToDelete := len(localUsers) - cfg.userCount
		for i := 0; i < usersToDelete; i++ {
			currentUser := localUsers[i]
			if i%100 == 0 {
				writeInfo(out, "Deleted %d out of %d users", i, usersToDelete)
			}
			g.Go(func() {
				executeDeleteUser(ctx, currentUser)
			})
		}

		teamsToDelete := len(localTeams) - cfg.teamCount
		for i := 0; i < teamsToDelete; i++ {
			currentTeam := localTeams[i]
			if i%100 == 0 {
				writeInfo(out, "Deleted %d out of %d teams", i, teamsToDelete)
			}
			g.Go(func() {
				executeDeleteTeam(ctx, currentTeam)
			})
		}

		for _, t := range localTeams {
			currentTeam := t
			g.Go(func() {
				executeDeleteTeamMembershipsForTeam(ctx, currentTeam.Org, currentTeam.Name)
			})
		}

		g.Wait()

		remoteOrgs := getGitHubOrgs(ctx)
		remoteTeams := getGitHubTeams(ctx, localOrgs)
		remoteUsers := getGitHubUsers(ctx)

		writeInfo(out, "Total orgs on instance: %d", len(remoteOrgs))
		writeInfo(out, "Total teams on instance: %d", len(remoteTeams))
		writeInfo(out, "Total users on instance: %d", len(remoteUsers))
	}

	end := time.Now()
	writeInfo(out, "Started at %s, finished at %s", start.String(), end.String())
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

	if grErr = store.deleteTeam(currentTeam); grErr != nil {
		log.Fatal(grErr)
	}

	writeSuccess(out, "Deleted team %s", currentTeam.Name)
}

func executeDeleteUser(ctx context.Context, currentUser *user) {
	existingUser, resp, grErr := gh.Users.Get(ctx, currentUser.Login)

	if grErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "Failed to get user %s, reason: %s", currentUser.Login, grErr)
		return
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
	if grErr = store.deleteUser(currentUser); grErr != nil {
		log.Fatal(grErr)
	}

	writeSuccess(out, "Deleted user %s", currentUser.Login)
}

func executeDeleteTeamMembershipsForTeam(ctx context.Context, org string, team string) {
	teamMembers, _, err := gh.Teams.ListTeamMembersBySlug(ctx, org, team, &github.TeamListTeamMembersOptions{
		Role:        "member",
		ListOptions: github.ListOptions{PerPage: 100},
	})

	if err != nil {
		log.Fatal(err)
	}

	writeInfo(out, "Deleting %d memberships for team %s", len(teamMembers), team)
	for _, member := range teamMembers {
		_, err = gh.Teams.RemoveTeamMembershipBySlug(ctx, org, team, *member.Login)
		if err != nil {
			log.Printf("Failed to remove membership from team %s for user %s: %s", team, *member.Login, err)
		}
	}
}

type teamMembershipOpts struct {
	currentUser *user
	teams       []*team

	membershipsPerUser int
	userIndex          int
	//teamIncrement      int
}

func executeCreateTeamMembershipsForUser(ctx context.Context, opts *teamMembershipOpts, membershipsDone *int64) {
	// users need to be member of the team's parent org to join the team
	userState := "active"
	userRole := "member"

	index := opts.userIndex % len(opts.teams)
	for j := 0; j < opts.membershipsPerUser; j++ {
		candidateTeam := opts.teams[index+j]

		if candidateTeam.TotalMembers >= 8 {
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

func executeCreateTeamMembershipsForTeam(ctx context.Context, t *team, users []*user, membershipsDone *int64) {
	// users need to be member of the team's parent org to join the team
	userState := "active"
	userRole := "member"

	for _, u := range users {
		// add user to team's parent org first
		_, _, mErr := gh.Organizations.EditOrgMembership(ctx, u.Login, t.Org, &github.Membership{
			State:        &userState,
			Role:         &userRole,
			Organization: &github.Organization{Login: &t.Org},
			User:         &github.User{Login: &u.Login},
		})

		if mErr != nil {
			writeFailure(out, "Failed to add user %s to organization %s, reason: %s", u.Login, t.Org, mErr)
			t.Failed = mErr.Error()
			if mErr = store.saveTeam(t); mErr != nil {
				log.Fatal(mErr)
			}
			continue
		}

		// this is an idempotent operation so no need to check existing membership
		_, _, mErr = gh.Teams.AddTeamMembershipBySlug(ctx, t.Org, t.Name, u.Login, nil)

		if mErr != nil {
			writeFailure(out, "Failed to add user %s to team %s, reason: %s", u, t.Name, mErr)
			t.Failed = mErr.Error()
			if mErr = store.saveTeam(t); mErr != nil {
				log.Fatal(mErr)
			}
			continue
		}

		t.TotalMembers += 1
		atomic.AddInt64(membershipsDone, 1)
		progress.SetValue(3, float64(*membershipsDone))

		if mErr = store.saveTeam(t); mErr != nil {
			log.Fatal(mErr)
		}
	}
}

func getGitHubOrgs(ctx context.Context) []*github.Organization {
	var orgs []*github.Organization
	var since int64
	for true {
		writeInfo(out, "Fetching org page, last ID seen is %d", since)
		orgsPage, _, err := gh.Organizations.ListAll(ctx, &github.OrganizationsListOptions{
			Since:       since,
			ListOptions: github.ListOptions{PerPage: 100},
		})
		if err != nil {
			log.Fatal(err)
		}
		if len(orgsPage) != 0 {
			since = *orgsPage[len(orgsPage)-1].ID
			orgs = append(orgs, orgsPage...)
		} else {
			break
		}
	}

	return orgs
}

func getGitHubTeams(ctx context.Context, orgs []*org) []*github.Team {
	var teams []*github.Team
	var currentPage int
	for _, o := range orgs {
		for true {
			writeInfo(out, "Fetching team page %d for org %s", currentPage, o.Login)
			teamsPage, _, err := gh.Teams.ListTeams(ctx, o.Login, &github.ListOptions{
				Page:    currentPage,
				PerPage: 100,
			})
			// not returned in API response but necessary
			for _, t := range teamsPage {
				t.Organization = &github.Organization{Login: &o.Login}
			}
			if err != nil {
				log.Fatal(err)
			}
			if len(teamsPage) != 0 {
				currentPage++
				teams = append(teams, teamsPage...)
			} else {
				break
			}
		}
		currentPage = 0
	}

	return teams
}

func getGitHubUsers(ctx context.Context) []*github.User {
	var users []*github.User
	var since int64
	for true {
		writeInfo(out, "Fetching user page, last ID seen is %d", since)
		usersPage, _, err := gh.Users.ListAll(ctx, &github.UserListOptions{
			Since:       since,
			ListOptions: github.ListOptions{PerPage: 100},
		})
		if err != nil {
			log.Fatal(err)
		}
		if len(usersPage) != 0 {
			since = *usersPage[len(usersPage)-1].ID
			users = append(users, usersPage...)
		} else {
			break
		}
	}

	return users
}

func getGitHubRepos(ctx context.Context) []*github.Repository {
	g := group.NewWithResults[[]*github.Repository]().WithMaxConcurrency(250)
	// 200k repos + some buffer space returning empty pages
	for i := 0; i < 2050; i++ {
		writeInfo(out, "Fetching repo page %d", i)
		page := i
		g.Go(func() []*github.Repository {
			var resp *github.Response
			var reposPage []*github.Repository
			var err error
			for resp == nil || resp.StatusCode == 502 || resp.StatusCode == 504 {
				if resp != nil && (resp.StatusCode == 502 || resp.StatusCode == 504) {
					writeInfo(out, "Response status %d, retrying in a minute", resp.StatusCode)
					time.Sleep(time.Minute)
				}
				reposPage, resp, err = gh.Repositories.ListByOrg(ctx, "blank200k", &github.RepositoryListByOrgOptions{
					Type: "private",
					ListOptions: github.ListOptions{
						Page:    page,
						PerPage: 100,
					},
				})
				if err != nil {
					log.Print(err)
				}
			}
			return reposPage
		})
	}
	var repos []*github.Repository
	for _, repo := range g.Wait() {
		repos = append(repos, repo...)
	}
	return repos
}

func executeCreateUser(ctx context.Context, u *user, usersDone *int64) {
	if u.Created && u.Failed == "" {
		atomic.AddInt64(usersDone, 1)
		progress.SetValue(2, float64(*usersDone))
		return
	}

	existingUser, resp, uErr := gh.Users.Get(ctx, u.Login)
	if uErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "Failed to get user %s, reason: %s", u.Login, uErr)
		return
	}

	uErr = nil
	if existingUser != nil {
		u.Created = true
		u.Failed = ""
		if uErr = store.saveUser(u); uErr != nil {
			log.Fatal(uErr)
		}
		//writeInfo(out, "user with login %s already exists", u.Login)
		atomic.AddInt64(usersDone, 1)
		progress.SetValue(2, float64(*usersDone))
		return
	}

	_, _, uErr = gh.Admin.CreateUser(ctx, u.Login, u.Email)
	if uErr != nil {
		writeFailure(out, "Failed to create user with login %s, reason: %s", u.Login, uErr)
		u.Failed = uErr.Error()
		if uErr = store.saveUser(u); uErr != nil {
			log.Fatal(uErr)
		}
		return
	}

	u.Created = true
	u.Failed = ""
	atomic.AddInt64(usersDone, 1)
	progress.SetValue(2, float64(*usersDone))
	if uErr = store.saveUser(u); uErr != nil {
		log.Fatal(uErr)
	}

	//writeSuccess(out, "Created user with login %s", u.Login)
}

func executeCreateUserImpersonationToken(ctx context.Context, u *user) string {
	auth, _, err := gh.Admin.CreateUserImpersonation(ctx, u.Login, &github.ImpersonateUserOptions{Scopes: []string{"repo"}})
	if err != nil {
		log.Fatal(err)
	}

	return auth.GetToken()
}

func executeCreateTeam(ctx context.Context, t *team, teamsDone *int64) {
	if t.Created && t.Failed == "" {
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))
		return
	}

	existingTeam, resp, tErr := gh.Teams.GetTeamBySlug(ctx, t.Org, t.Name)

	if tErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "failed to get team with name %s, reason: %s", t.Name, tErr)
		return
	}

	tErr = nil
	if existingTeam != nil {
		t.Created = true
		t.Failed = ""
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))

		if tErr = store.saveTeam(t); tErr != nil {
			log.Fatal(tErr)
		}
	} else {
		// Create the team if not exists
		if _, _, tErr = gh.Teams.CreateTeam(ctx, t.Org, github.NewTeam{Name: t.Name}); tErr != nil {
			writeFailure(out, "Failed to create team with name %s, reason: %s", t.Name, tErr)
			t.Failed = tErr.Error()
			if tErr = store.saveTeam(t); tErr != nil {
				log.Fatal(tErr)
			}
		}

		t.Created = true
		t.Failed = ""
		atomic.AddInt64(teamsDone, 1)
		progress.SetValue(1, float64(*teamsDone))

		if tErr = store.saveTeam(t); tErr != nil {
			log.Fatal(tErr)
		}
	}
}

func executeCreateOrg(ctx context.Context, o *org, orgAdmin string, orgsDone *int64) {
	if o.Created && o.Failed == "" {
		atomic.AddInt64(orgsDone, 1)
		progress.SetValue(0, float64(*orgsDone))
		return
	}

	existingOrg, resp, oErr := gh.Organizations.Get(ctx, o.Login)
	if oErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "Failed to get org %s, reason: %s", o.Login, oErr)
		return
	}

	oErr = nil
	if existingOrg != nil {
		o.Created = true
		o.Failed = ""
		atomic.AddInt64(orgsDone, 1)
		progress.SetValue(0, float64(*orgsDone))

		if oErr = store.saveOrg(o); oErr != nil {
			log.Fatal(oErr)
		}
		return
	}

	_, _, oErr = gh.Admin.CreateOrg(ctx, &github.Organization{Login: &o.Login}, orgAdmin)

	if oErr != nil {
		writeFailure(out, "Failed to create org with login %s, reason: %s", o.Login, oErr)
		o.Failed = oErr.Error()
		if oErr = store.saveOrg(o); oErr != nil {
			log.Fatal(oErr)
		}
		return
	}

	atomic.AddInt64(orgsDone, 1)
	progress.SetValue(0, float64(*orgsDone))

	o.Created = true
	o.Failed = ""
	if oErr = store.saveOrg(o); oErr != nil {
		log.Fatal(oErr)
	}

	//writeSuccess(out, "Created org with login %s", o.Login)
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
