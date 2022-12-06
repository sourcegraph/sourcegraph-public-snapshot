package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

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
		delete(ctx, cfg)

	case "validate":
		validate(ctx)
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
