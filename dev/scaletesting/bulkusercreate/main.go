package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alecthomas/chroma/lexers/g"
	github "github.com/google/go-github/v41/github"
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
	subOrgCount    int
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
	flag.IntVar(&cfg.subOrgCount, "suborg.count", 10, "Amount of sub-orgs to create or delete")
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

	switch cfg.action {
	case "create":
		create(ctx, orgs, cfg)

	case "delete":
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
		g.Wait()

		for _, t := range localTeams {
			currentTeam := t
			g.Go(func() {
				executeDeleteTeamMembershipsForTeam(ctx, currentTeam.Org, currentTeam.Name)
			})
		}
		g.Wait()

	case "validate":
		localTeams, err := store.loadTeams()
		if err != nil {
			log.Fatal("Failed to load teams", err)
		}

		localRepos, err := store.loadRepos()
		if err != nil {
			log.Fatal("Failed to load repos", err)
		}

		teamSizes := make(map[int]int)
		for _, t := range localTeams {
			users, _, err := gh.Teams.ListTeamMembersBySlug(ctx, t.Org, t.Name, &github.TeamListTeamMembersOptions{
				Role:        "member",
				ListOptions: github.ListOptions{PerPage: 100},
			})
			if err != nil {
				log.Fatal(err)
			}
			teamSizes[len(users)]++
		}

		for k, v := range teamSizes {
			writeInfo(out, "Found %d teams with %d members", v, k)
		}

		remoteOrgs := getGitHubOrgs(ctx)
		remoteUsers := getGitHubUsers(ctx)

		writeInfo(out, "Total orgs on instance: %d", len(remoteOrgs))
		writeInfo(out, "Total users on instance: %d", len(remoteUsers))

		g3 := group.New().WithMaxConcurrency(1000)

		var orgRepoCount int64
		var teamRepoCount int64
		var userRepoCount int64

		for i, r := range localRepos {
			cI := i
			cR := r

			g3.Go(func() {
				writeInfo(out, "Processing repo %d", cI)
			retryRepoContributors:
				contributors, res, err := gh.Repositories.ListContributors(ctx, cR.Owner, cR.Name, &github.ListContributorsOptions{
					Anon:        "false",
					ListOptions: github.ListOptions{},
				})
				if err != nil {
					log.Fatalf("Failed getting contributors for repo %s/%s: %s", cR.Owner, cR.Name, err)
				}
				if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
					time.Sleep(30 * time.Second)
					goto retryRepoContributors
				}
				if len(contributors) != 0 {
					// Permissions assigned on user level
					atomic.AddInt64(&userRepoCount, 1)
					return
				}

			retryRepoTeams:
				teams, res, err := gh.Repositories.ListTeams(ctx, cR.Owner, cR.Name, &github.ListOptions{})
				if err != nil {
					log.Fatalf("Failed getting teams for repo %s/%s: %s", cR.Owner, cR.Name, err)
				}
				if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
					time.Sleep(30 * time.Second)
					goto retryRepoTeams
				}
				if len(teams) != 0 {
					// Permissions assigned on user level
					atomic.AddInt64(&teamRepoCount, 1)
					return
				}

				// If we get this far the repo is org-wide
				atomic.AddInt64(&orgRepoCount, 1)
			})
		}
		g3.Wait()

		writeInfo(out, "Total org-scoped repos: %d", orgRepoCount)
		writeInfo(out, "Total team-scoped repos: %d", teamRepoCount)
		writeInfo(out, "Total user-scoped repos: %d", userRepoCount)
	}

	end := time.Now()
	writeInfo(out, "Started at %s, finished at %s", start.String(), end.String())
}

func generateUserOAuthCsv(ctx context.Context, users []*user, tokensDone int64) {
	tg := group.NewWithResults[userToken]().WithMaxConcurrency(1000)
	for _, u := range users {
		currentU := u
		tg.Go(func() userToken {
			token := executeCreateUserImpersonationToken(ctx, currentU)
			atomic.AddInt64(&tokensDone, 1)
			progress.SetValue(5, float64(tokensDone))
			return userToken{
				login: currentU.Login,
				token: token,
			}
		})
	}
	pairs := tg.Wait()

	csvFile, err := os.Create("users.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("Failed creating csv: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	_ = csvwriter.Write([]string{"login", "token"})

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

func categorizeOrgRepos(cfg config, repos []*repo, orgs []*org) (*org, map[*org][]*repo) {
	repoCategories := make(map[*org][]*repo)

	// 1% of repos divided equally over sub-orgs
	var mainOrg *org
	var subOrgs []*org
	reposPerSubOrg := (len(repos) / 100) / cfg.subOrgCount
	for _, o := range orgs {
		if strings.HasPrefix(o.Login, "sub-org") {
			subOrgs = append(subOrgs, o)
		} else {
			mainOrg = o
		}
	}

	for i, o := range subOrgs {
		subOrgRepos := repos[i*reposPerSubOrg : (i+1)*reposPerSubOrg]
		repoCategories[o] = subOrgRepos
	}

	// rest assigned to main org
	repoCategories[mainOrg] = repos[len(subOrgs)*reposPerSubOrg:]

	return mainOrg, repoCategories
}

func executeAssignOrgRepos(ctx context.Context, reposPerOrg map[*org][]*repo, users []*user, reposDone *int64, g group.Group) {
	for o, repos := range reposPerOrg {
		currentOrg := o
		currentRepos := repos

		var res *github.Response
		var err error
		for _, r := range currentRepos {
			currentRepo := r
			g.Go(func() {
				if currentOrg.Login == currentRepo.Owner {
					//writeInfo(out, "Repository %s already owned by %s", r.Name, r.Owner)
					// The repository is already transferred
					atomic.AddInt64(reposDone, 1)
					progress.SetValue(4, float64(*reposDone))
					return
				}

				for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
					if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
						time.Sleep(30 * time.Second)
					}

					_, res, err = gh.Repositories.Transfer(ctx, "blank200k", currentRepo.Name, github.TransferRequest{NewOwner: currentOrg.Login})

					if res.StatusCode == 422 {
						body, err := io.ReadAll(res.Body)
						if err != nil {
							log.Fatalf("Failed reading response body: %s", err)
						}
						if strings.Contains(string(body), "Repositories cannot be transferred to the original owner") {
							//writeInfo(out, "Repository %s already owned by %s [not saved to state, current owner: %s]", r.Name, currentOrg.Login, r.Owner)
							// The repository is already transferred but not yet saved in the state
							break
						}
					}

					if err != nil {
						if _, ok := err.(*github.AcceptedError); ok {
							//writeInfo(out, "Repository %s scheduled for transfer as a background job", r.Name)
							// AcceptedError means the transfer is scheduled as a background job
							break
						} else {
							log.Fatalf("Failed to transfer repository %s from %s to %s: %s", currentRepo.Name, currentRepo.Owner, currentOrg.Login, err)
						}
					}
				}

				//writeInfo(out, "Repository %s transferred to %s", r.Name, r.Owner)
				atomic.AddInt64(reposDone, 1)
				progress.SetValue(4, float64(*reposDone))
				currentRepo.Owner = currentOrg.Login
				if err = store.saveRepo(currentRepo); err != nil {
					log.Fatalf("Failed to save repository %s: %s", currentRepo.Name, err)
				}
			})
		}

		g.Wait()

		if strings.HasPrefix(currentOrg.Login, "sub-org") {
			// add 2000 users to sub-orgs
			index, err := strconv.ParseInt(strings.TrimPrefix(currentOrg.Login, "sub-org-"), 10, 32)
			if err != nil {
				log.Fatalf("Failed to parse index from sub-org id: %s", err)
			}
			usersToAdd := users[index*2000 : (index+1)*2000]

			for _, u := range usersToAdd {
				currentUser := u
				var uRes *github.Response
				var uErr error
				g.Go(func() {
					for uRes == nil || uRes.StatusCode == 502 || uRes.StatusCode == 504 {
						if uRes != nil && (uRes.StatusCode == 502 || uRes.StatusCode == 504) {
							time.Sleep(30 * time.Second)
						}

						memberState := "active"
						memberRole := "member"
						_, uRes, uErr = gh.Organizations.EditOrgMembership(ctx, currentUser.Login, currentOrg.Login, &github.Membership{
							State: &memberState,
							Role:  &memberRole,
						})
						if uErr != nil {
							log.Fatalf("Failed edit membership of user %s in org %s: %s", currentUser.Login, currentOrg.Login, uErr)
						}
					}
				})
			}
		}
	}
}

func categorizeTeamRepos(cfg config, mainOrgRepos []*repo, teams []*team) map[*team][]*repo {
	// 1% of teams
	teamsLarge := int(math.Ceil(float64(cfg.teamCount) * 0.01))
	// 0.5% of repos per team
	reposLarge := int(math.Floor(float64(len(mainOrgRepos)) * 0.005))

	// 4% of teams
	teamsMedium := int(math.Ceil(float64(cfg.teamCount) * 0.04))
	// 0.04% of repos per team
	reposMedium := int(math.Floor(float64(len(mainOrgRepos)) * 0.0004))

	// 95% of teams
	teamsSmall := int(math.Ceil(float64(cfg.teamCount) * 0.95))
	// remainder of repos divided over remainder of teams
	reposSmall := int(math.Floor(float64(len(mainOrgRepos)-(reposMedium*teamsMedium)-(reposLarge*teamsLarge)) / float64(teamsSmall)))

	teamCategories := make(map[*team][]*repo)

	for i := 0; i < teamsSmall; i++ {
		currentTeam := teams[i]
		teamRepos := mainOrgRepos[i*reposSmall : (i+1)*reposSmall]
		teamCategories[currentTeam] = teamRepos
	}

	for i := 0; i < teamsMedium; i++ {
		currentTeam := teams[teamsSmall+i]
		startIndex := (teamsSmall * reposSmall) + (i * reposMedium)
		endIndex := (teamsSmall * reposSmall) + ((i + 1) * reposMedium)
		teamRepos := mainOrgRepos[startIndex:endIndex]
		teamCategories[currentTeam] = teamRepos
	}

	for i := 0; i < teamsLarge; i++ {
		currentTeam := teams[teamsSmall+teamsMedium+i]
		startIndex := (teamsSmall * reposSmall) + (teamsMedium * reposMedium) + (i * reposLarge)
		endIndex := (teamsSmall * reposSmall) + (teamsMedium * reposMedium) + ((i + 1) * reposLarge)
		teamRepos := mainOrgRepos[startIndex:endIndex]
		teamCategories[currentTeam] = teamRepos
	}

	remainderIndex := (teamsSmall * reposSmall) + (teamsMedium * reposMedium) + (teamsLarge * reposLarge)
	remainingRepos := mainOrgRepos[remainderIndex:]
	for i, r := range remainingRepos {
		t := teams[i%teamsSmall]
		teamCategories[t] = append(teamCategories[t], r)
	}

	teamsWithNils := make(map[*team][]*repo)
	for t, rr := range teamCategories {
		for _, r := range rr {
			if r == nil {
				teamsWithNils[t] = rr
				break
			}
		}
	}

	return teamCategories
}

func executeAssignTeamRepos(ctx context.Context, reposPerTeam map[*team][]*repo, reposDone *int64, g group.Group) {
	for t, repos := range reposPerTeam {
		currentTeam := t
		currentRepos := repos

		g.Go(func() {
			for _, r := range currentRepos {
				currentRepo := r
				if r.Owner == fmt.Sprintf("%s/%s", currentTeam.Org, currentTeam.Name) {
					// team is already owner
					//writeInfo(out, "Repository %s already owned by %s", r.Name, currentTeam.Name)
					atomic.AddInt64(reposDone, 1)
					progress.SetValue(4, float64(*reposDone))
					continue
				}

				var res *github.Response
				var err error
				for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
					if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
						time.Sleep(30 * time.Second)
					}

					res, err = gh.Teams.AddTeamRepoBySlug(ctx, currentTeam.Org, currentTeam.Name, currentTeam.Org, currentRepo.Name, &github.TeamAddTeamRepoOptions{Permission: "push"})

					if res.StatusCode == 422 {
						body, err := io.ReadAll(res.Body)
						if err != nil {
							log.Fatalf("Failed reading response body: %s", err)
						}
						log.Fatalf("Failed to assign repo %s to team %s: %s", currentRepo.Name, currentTeam.Name, string(body))
					}

					if err != nil {
						log.Fatalf("Failed to transfer repository %s from %s to %s: %s", currentRepo.Name, currentRepo.Owner, currentTeam.Name, err)
					}
				}

				atomic.AddInt64(reposDone, 1)
				progress.SetValue(4, float64(*reposDone))
				currentRepo.Owner = fmt.Sprintf("%s/%s", currentTeam.Org, currentTeam.Name)
				if err = store.saveRepo(r); err != nil {
					log.Fatalf("Failed to save repository %s: %s", currentRepo.Name, err)
				}
				//writeInfo(out, "Repository %s transferred to %s", r.Name, currentTeam.Name)
			}
		})
	}
}

func categorizeUserRepos(mainOrgRepos []*repo, users []*user) map[*repo][]*user {
	repoUsers := make(map[*repo][]*user)
	usersPerRepo := 3
	for i, r := range mainOrgRepos {
		usersForRepo := users[i*usersPerRepo : (i+1)*usersPerRepo]
		repoUsers[r] = usersForRepo
	}

	return repoUsers
}

func executeAssignUserRepos(ctx context.Context, usersPerRepo map[*repo][]*user, reposDone *int64, g group.Group) {
	for r, users := range usersPerRepo {
		currentRepo := r
		currentUsers := users

		g.Go(func() {
			for _, u := range currentUsers {
				var res *github.Response
				var err error
				for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
					if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
						time.Sleep(30 * time.Second)
					}

					var invitation *github.CollaboratorInvitation
					invitation, res, err = gh.Repositories.AddCollaborator(ctx, currentRepo.Owner, currentRepo.Name, u.Login, &github.RepositoryAddCollaboratorOptions{Permission: "push"})
					if err != nil {
						log.Fatalf("Failed to add user %s as a collaborator to repo %s: %s", u.Login, currentRepo.Name, err)
					}

					// AddCollaborator returns a 201 when an invitation is created.
					//
					// A 204 is returned when:
					// * an existing collaborator is added as a collaborator
					// * an organization member is added as an individual collaborator
					// * an existing team member (whose team is also a repository collaborator) is added as an individual collaborator
					if res.StatusCode == 201 {
						res, err = gh.Users.AcceptInvitation(ctx, invitation.GetID())
						if err != nil {
							log.Fatalf("Failed to accept collaborator invitation for user %s on repo %s: %s", u.Login, currentRepo.Name, err)
						}
					}
				}
			}
			atomic.AddInt64(reposDone, 1)
			progress.SetValue(4, float64(*reposDone))
			//writeInfo(out, "Repository %s transferred to users", r.Name)
		})
	}
}

func executeDeleteTeam(ctx context.Context, currentTeam *team) {
	existingTeam, resp, grErr := gh.Teams.GetTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)

	if grErr != nil && resp.StatusCode != 404 {
		writeFailure(out, "Failed to get team %s, reason: %s\n", currentTeam.Name, grErr)
	}

	grErr = nil
	if existingTeam != nil {
		_, grErr = gh.Teams.DeleteTeamBySlug(ctx, currentTeam.Org, currentTeam.Name)
		if grErr != nil {
			writeFailure(out, "Failed to delete team %s, reason: %s\n", currentTeam.Name, grErr)
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
		writeFailure(out, "Failed to get user %s, reason: %s\n", currentUser.Login, grErr)
		return
	}

	grErr = nil
	if existingUser != nil {
		_, grErr = gh.Admin.DeleteUser(ctx, currentUser.Login)

		if grErr != nil {
			writeFailure(out, "Failed to delete user with login %s, reason: %s\n", currentUser.Login, grErr)
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

func executeCreateTeamMembershipsForTeam(ctx context.Context, t *team, users []*user, membershipsDone *int64) {
	// users need to be member of the team's parent org to join the team
	userState := "active"
	userRole := "member"

	for _, u := range users {
		// add user to team's parent org first
		var res *github.Response
		var err error
		for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
			if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
				time.Sleep(30 * time.Second)
			}
			_, res, err = gh.Organizations.EditOrgMembership(ctx, u.Login, t.Org, &github.Membership{
				State:        &userState,
				Role:         &userRole,
				Organization: &github.Organization{Login: &t.Org},
				User:         &github.User{Login: &u.Login},
			})

			if err != nil {
				if err = t.setFailedAndSave(err); err != nil {
					log.Fatal(err)
				}
				continue
			}
		}

		res = nil
		for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
			if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
				time.Sleep(30 * time.Second)
			}
			// this is an idempotent operation so no need to check existing membership
			_, res, err = gh.Teams.AddTeamMembershipBySlug(ctx, t.Org, t.Name, u.Login, nil)
			if err != nil {
				if err = t.setFailedAndSave(err); err != nil {
					log.Fatal(err)
				}
				continue
			}
		}

		t.TotalMembers += 1
		atomic.AddInt64(membershipsDone, 1)
		progress.SetValue(3, float64(*membershipsDone))

		if err = store.saveTeam(t); err != nil {
			log.Fatal(err)
		}
	}
}

func getGitHubOrgs(ctx context.Context) []*github.Organization {
	var orgs []*github.Organization
	var since int64
	for true {
		//writeInfo(out, "Fetching org page, last ID seen is %d", since)
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
			//writeInfo(out, "Fetching team page %d for org %s", currentPage, o.Login)
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
		//writeInfo(out, "Fetching user page, last ID seen is %d", since)
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
		writeFailure(out, "Failed to get user %s, reason: %s\n", u.Login, uErr)
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
		writeFailure(out, "Failed to create user with login %s, reason: %s\n", u.Login, uErr)
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
	auth, _, err := gh.Admin.CreateUserImpersonation(ctx, u.Login, &github.ImpersonateUserOptions{Scopes: []string{"repo", "read:org", "read:user_email"}})
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
		writeFailure(out, "failed to get team with name %s, reason: %s\n", t.Name, tErr)
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
		var res *github.Response
		var err error
		for res == nil || res.StatusCode == 502 || res.StatusCode == 504 {
			if res != nil && (res.StatusCode == 502 || res.StatusCode == 504) {
				// give some breathing room
				time.Sleep(30 * time.Second)
			}

			if _, res, err = gh.Teams.CreateTeam(ctx, t.Org, github.NewTeam{Name: t.Name}); err != nil {
				if err = t.setFailedAndSave(err); err != nil {
					log.Fatalf("Failed saving to state: %s", err)
				}
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
		writeFailure(out, "Failed to get org %s, reason: %s\n", o.Login, oErr)
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
		writeFailure(out, "Failed to create org with login %s, reason: %s\n", o.Login, oErr)
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
