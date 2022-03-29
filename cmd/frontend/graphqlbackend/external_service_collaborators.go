package graphqlbackend

import (
	"context"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// See schema.graphql for an explanation of how this is intended to be used. This is particularly
// for listing collaborators to *some* of the repositories associated with this external service
// *before* they are cloned onto Sourcegraph.
func (r *externalServiceResolver) InvitableCollaborators(ctx context.Context) ([]*invitableCollaboratorResolver, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	// SECURITY: This API should only be exposed for user-added external services, not for example by
	// site-wide (CloudDefault) external services (the API also makes zero sense in that context.)
	//
	// IMPORTANT: This API is allowed for org external services. You may not have access to every repo
	// within an org external service, and so if we expose too much information here it could be an
	// ACL vulnerability. Since we only expose name+email+avatar URL, this is fine to do.
	if r.externalService.IsSiteOwned() {
		return nil, errors.New("InvitableCollaborators may only be used on user-added external services.")
	}
	cfg, err := r.externalService.Configuration()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse external service config")
	}
	githubCfg, ok := cfg.(*schema.GitHubConnection)
	if !ok {
		// We only support GitHub for now as that's the most popular / important.
		// Don't return an error, just an empty list, because e.g. that's what you want if we just
		// can't find any collaborators.
		return []*invitableCollaboratorResolver{}, nil
	}
	baseURL, err := url.Parse(githubCfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse external service URL")
	}
	baseURL = extsvc.NormalizeBaseURL(baseURL)
	githubUrl, _ := github.APIRoot(baseURL)
	client := github.NewV4Client(r.externalService.URN(), githubUrl, &auth.OAuthBearerToken{Token: githubCfg.Token}, nil)

	possibleRepos := githubCfg.Repos
	if len(possibleRepos) == 0 {
		// External service is describing "sync all repos" instead of a specific set. Query a few of
		// those that we'll look for collaborators in.
		repos, err := backend.NewRepos(r.db).List(ctx, database.ReposListOptions{
			// SECURITY: This must be the authenticated user's external service ID.
			ExternalServiceIDs: []int64{r.externalService.ID},
			OrderBy: database.RepoListOrderBy{{
				Field:      "name",
				Descending: false,
			}},
			LimitOffset: &database.LimitOffset{Limit: 1000},
		})
		if err != nil {
			return nil, errors.Wrap(err, "Repos.List")
		}
		for _, repo := range repos {
			// repo.URI is in "github.com/gorilla/mux" form, we need "gorilla/mux" form to match
			// githubCfg.Repos and so we parse the URI and use the path.
			uri, _ := url.Parse("https://" + repo.URI)
			possibleRepos = append(possibleRepos, uri.Path[1:])
		}
	}

	// We'll look in up to 25 repos for collaborators. Each client.RecentCommitters API call uses
	// 1 point in GitHub's GraphQL API rate limiting, and we are allowed 5,000 per hour (which we
	// share with other parts of Sourcegraph such as repo-updater.) and so we could probably safely
	// use up to a few hundred here. However, GitHub's recent commits API is quite slow (it appears
	// to even run separate GraphQL requests for the same client IP in sequence rather than in
	// parallel) and so that is the true limiting factor here.
	//
	// We search within random repositories because many follow a pattern, such as say adding a ton
	// of `company/lsif-java`, `company/lsif-python`, `company/lsif-typescript` etc repos with likely
	// the same collaborators, whereas random sampling may give us dissimilar repositories.
	const maxReposToScan = 25
	pickedRepos := pickReposToScanForCollaborators(possibleRepos, maxReposToScan)

	// In parallel collect all recent committers info for the few repos we're going to scan.
	recentCommitters, err := parallelRecentCommitters(ctx, pickedRepos, client.RecentCommitters)
	if err != nil {
		return nil, err
	}

	authUserEmails, err := database.UserEmails(r.db).ListByUser(ctx, database.UserEmailsListOptions{
		UserID: a.UID,
	})
	if err != nil {
		return nil, err
	}

	userExistsByUsername := func(username string) bool {
		_, err := r.db.Users().GetByUsername(ctx, username)
		return err == nil
	}
	userExistsByEmail := func(email string) bool {
		_, err := r.db.Users().GetByVerifiedEmail(ctx, email)
		return err == nil
	}
	return filterInvitableCollaborators(recentCommitters, authUserEmails, userExistsByUsername, userExistsByEmail), nil
}

type invitableCollaboratorResolver struct {
	likelySourcegraphUsername string
	email                     string
	name                      string
	avatarURL                 string
	date                      time.Time
}

func (i *invitableCollaboratorResolver) Name() string        { return i.name }
func (i *invitableCollaboratorResolver) Email() string       { return i.email }
func (i *invitableCollaboratorResolver) DisplayName() string { return i.name }
func (i *invitableCollaboratorResolver) AvatarURL() *string {
	if i.avatarURL == "" {
		return nil
	}
	return &i.avatarURL
}
func (i *invitableCollaboratorResolver) User() *UserResolver { return nil }

type RecentCommittersFunc func(context.Context, *github.RecentCommittersParams) (*github.RecentCommittersResults, error)

func pickReposToScanForCollaborators(possibleRepos []string, maxReposToScan int) []string {
	var picked []string
	swapRemove := func(i int) {
		s := possibleRepos
		s[i] = s[len(s)-1]
		possibleRepos = s[:len(s)-1]
	}
	for len(picked) < maxReposToScan && len(possibleRepos) > 0 {
		randomRepoIndex := rand.Intn(len(possibleRepos))
		picked = append(picked, possibleRepos[randomRepoIndex])
		swapRemove(randomRepoIndex)
	}
	return picked
}

func parallelRecentCommitters(ctx context.Context, repos []string, recentCommitters RecentCommittersFunc) (allRecentCommitters []*invitableCollaboratorResolver, err error) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, repoName := range repos {
		owner, name, err := github.SplitRepositoryNameWithOwner(repoName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to split repository name")
		}
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()
			recentCommits, err := recentCommitters(ctx, &github.RecentCommittersParams{
				Name:  name,
				Owner: owner,
				First: 100,
			})
			if err != nil {
				log15.Error("InvitableCollaborators: failed to get recent committers", "err", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, commit := range recentCommits.Nodes {
				for _, author := range commit.Authors.Nodes {
					parsedTime, _ := time.Parse(time.RFC3339, author.Date)
					allRecentCommitters = append(allRecentCommitters, &invitableCollaboratorResolver{
						likelySourcegraphUsername: author.User.Login,
						email:                     author.Email,
						name:                      author.Name,
						avatarURL:                 author.AvatarURL,
						date:                      parsedTime,
					})
				}
			}
		})
	}
	wg.Wait()
	return
}

func filterInvitableCollaborators(
	recentCommitters []*invitableCollaboratorResolver,
	authUserEmails []*database.UserEmail,
	userExistsByUsername func(username string) bool,
	userExistsByEmail func(email string) bool,
) []*invitableCollaboratorResolver {
	// Sort committers by most-recent-first. This ensures that the top of the list of people you can
	// invite are people who recently committed to code, which means they're more active and more
	// likely the person you want to invite (compared to e.g. if we hit a very old repo and the
	// committer is say no longer working at that organization.)
	sort.Slice(recentCommitters, func(i, j int) bool {
		a := recentCommitters[i].date
		b := recentCommitters[j].date
		return a.After(b)
	})

	// Eliminate committers who are duplicates, don't have an email, have a noreply@github.com
	// email (which happens when you make edits via the GitHub web UI), or committers with the same
	// email address as this authenticated user (can't invite ourselves, we already have an account.)
	var (
		invitable   []*invitableCollaboratorResolver
		deduplicate = map[string]struct{}{}
	)
	for _, recentCommitter := range recentCommitters {
		likelyBot := strings.Contains(recentCommitter.email, "bot") || strings.Contains(strings.ToLower(recentCommitter.name), "bot")
		if recentCommitter.email == "" || strings.Contains(recentCommitter.email, "noreply") || likelyBot {
			continue
		}
		isOurEmail := false
		for _, email := range authUserEmails {
			if recentCommitter.email == email.Email {
				isOurEmail = true
				continue
			}
		}
		if isOurEmail {
			continue
		}
		if _, duplicate := deduplicate[recentCommitter.email]; duplicate {
			continue
		}
		deduplicate[recentCommitter.email] = struct{}{}

		if len(invitable) > 200 {
			// 200 users is more than enough, don't do any more work (such as checking if users
			// exist.)
			break
		}
		// If a Sourcegraph user with a matching username exists, or a matching email exists, don't
		// consider them someone who is invitable (would be annoying to receive invites after having
		// an account.)
		if userExistsByEmail(recentCommitter.email) {
			continue
		}
		if userExistsByUsername(recentCommitter.likelySourcegraphUsername) {
			continue
		}

		invitable = append(invitable, recentCommitter)
	}

	// domain turns "stephen@sourcegraph.com" -> "sourcegraph.com"
	domain := func(email string) string {
		idx := strings.LastIndex(email, "@")
		if idx == -1 {
			return email
		}
		return email[idx:]
	}

	// Determine the number of invitable people per email domain, then sort so that those with the
	// most similar email domain to others in the list appear first. e.g. all @sourcegraph.com team
	// members should appear before a random @gmail.com contributor.
	invitablePerDomain := map[string]int{}
	for _, person := range invitable {
		current := invitablePerDomain[domain(person.email)]
		invitablePerDomain[domain(person.email)] = current + 1
	}
	sort.Slice(invitable, func(i, j int) bool {
		// First, sort popular personal email domains lower.
		iDomain := domain(invitable[i].email)
		jDomain := domain(invitable[j].email)
		if iDomain != jDomain {
			for _, popularPersonalDomain := range []string{"@gmail.com", "@yahoo.com", "@outlook.com", "@fastmail.com", "@protonmail.com"} {
				if jDomain == popularPersonalDomain {
					return true
				}
				if iDomain == popularPersonalDomain {
					return false
				}
			}

			// Sort domains with most invitable collaborators higher.
			iPeopleWithDomain := invitablePerDomain[iDomain]
			jPeopleWithDomain := invitablePerDomain[jDomain]
			if iPeopleWithDomain != jPeopleWithDomain {
				return iPeopleWithDomain > jPeopleWithDomain
			}
		}

		// Finally, sort most-recent committers higher.
		return invitable[i].date.After(invitable[j].date)
	})
	return invitable
}
