package graphqlbackend

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

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
	authUserEmails, err := database.UserEmails(r.db).ListByUser(ctx, database.UserEmailsListOptions{
		UserID: a.UID,
	})
	if err != nil {
		return nil, err
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
	client := github.NewV4Client(githubUrl, &auth.OAuthBearerToken{Token: githubCfg.Token}, nil)

	// We'll only look in 20 repos. We limit ourselves here to prevent having our github token run
	// into rate limits (which could affect repo discovery / cloning / other API operations we need
	// to perform for the user elsewhere.)
	const maxReposToScan = 20
	fewRepos := githubCfg.Repos
	if len(fewRepos) > maxReposToScan {
		fewRepos = fewRepos[:maxReposToScan]
	}

	// In parallel collect all recent committers info for the few repos we're going to scan.
	var (
		wg                  sync.WaitGroup
		mu                  sync.Mutex
		allRecentCommitters []*invitableCollaboratorResolver
	)
	for _, repoName := range fewRepos {
		owner, name, err := github.SplitRepositoryNameWithOwner(repoName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to split repository name")
		}
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()
			recentCommits, err := client.RecentCommitters(ctx, &github.RecentCommittersParams{
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
						email:     author.Email,
						name:      author.Name,
						avatarURL: author.AvatarURL,
						date:      parsedTime,
					})
				}
			}
		})
	}
	wg.Wait()

	// Sort committers by most-recent-first. This ensures that the top of the list of people you can
	// invite are people who recently committed to code, which means they're more active and more
	// likely the person you want to invite (compared to e.g. if we hit a very old repo and the
	// committer is say no longer working at that organization.)
	sort.Slice(allRecentCommitters, func(i, j int) bool {
		a := allRecentCommitters[i].date
		b := allRecentCommitters[j].date
		return a.After(b)
	})

	// Eliminate committers who are duplicates, don't have an email, have a noreply@github.com
	// email (which happens when you make edits via the GitHub web UI), or committers with the same
	// email address as this authenticated user (can't invite ourselves, we already have an account.)
	var (
		invitable   []*invitableCollaboratorResolver
		deduplicate = map[string]struct{}{}
	)
	for _, recentCommitter := range allRecentCommitters {
		if recentCommitter.email == "" || strings.Contains(recentCommitter.email, "noreply") {
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
		invitable = append(invitable, recentCommitter)
	}
	return invitable, nil
}

type invitableCollaboratorResolver struct {
	email     string
	name      string
	avatarURL string
	date      time.Time
}

func (i *invitableCollaboratorResolver) Name() string        { return i.name }
func (i *invitableCollaboratorResolver) Email() string       { return i.email }
func (i *invitableCollaboratorResolver) DisplayName() string { return i.name }
func (i *invitableCollaboratorResolver) AvatarURL() *string  { return &i.avatarURL }
func (i *invitableCollaboratorResolver) User() *UserResolver { return nil }
