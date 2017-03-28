package tracking

import (
	"context"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gcstracker"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// Limits to the number of errors we can recieve from external GitHub
// requests before giving up on fetching more data
//
// If GitHub returns an error on a fetch for, e.g., a single repo's
// languages, we want to continue to try with the next repo (unless
// it was a rate limit error, which is hanleded separately). Only
// after a sufficient number of unexplained errors do we want to
// give up
var maxOrgMemberErrors = 1
var maxRepoDetailsErrors = 4

// TrackUserGitHubData handles fetching limited information about
// a user's GitHub profile and sends it to Google Cloud Storage
// for analytics
func TrackUserGitHubData(actor *auth.Actor, event string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.TrackUserGitHubData: %s", err)
		}
	}()

	gcsClient, err := gcstracker.New(actor)
	if err != nil {
		return errors.Wrap(err, "gcstracker.New")
	}
	if gcsClient == nil {
		return nil
	}

	// Since the newly-authenticated actor (and their GitHubToken) has
	// not yet been associated with the request's context, we need to
	// create a temporary Context object that contains that linkage in
	// order to pull data from the GitHub API
	tempCtx := auth.WithActor(context.Background(), actor)

	// Fetch orgs and org members data
	// ListAllOrgs may return partial results
	orgs, err := backend.Orgs.ListAllOrgs(tempCtx, &sourcegraph.OrgListOptions{})
	if err != nil {
		log15.Warn("backend.Orgs.ListAllOrgs: failed to fetch some user organizations", "source", "GitHub", "error", err)
	}

	orgMembersErrCounter := 0
	owd := make(map[string]([]*github.User))
	for _, org := range orgs.Orgs {
		members, err := backend.Orgs.ListAllOrgMembers(tempCtx, &sourcegraph.OrgListOptions{OrgName: org.Login})
		if err != nil {
			// ListAllOrgMembers may return partial results
			// Don't give up unless maxOrgMemberErrors errors are caught
			orgMembersErrCounter = orgMembersErrCounter + 1
			if orgMembersErrCounter > maxOrgMemberErrors {
				log15.Warn("Orgs.ListAllOrgMembers: failed to fetch some user org members (max errors exceeded)", "source", "GitHub", "error", err)
				break
			} else {
				log15.Warn("Orgs.ListAllOrgMembers: failed to fetch some user org members", "source", "GitHub", "error", err)
			}
		}
		owd[org.Login] = members
	}

	// Fetch repo data
	tempCtx = context.WithValue(tempCtx, extgithub.GitHubTrackingContextKey, "TrackUserGitHubData")
	ghRwd, err := listAllGitHubReposWithDetails(tempCtx, &github.RepositoryListOptions{})
	if err != nil {
		log15.Warn("tracking.listAllGitHubReposWithDetails: failed to fetch some user repo details", "source", "GitHub", "error", err)
	}
	rwd := &sourcegraph.GitHubReposWithDetailsList{ReposWithDetails: ghRwd}

	// Add new TrackedObject
	tos := gcsClient.NewTrackedObjects(event)

	err = tos.AddOrgsWithDetailsObjects(owd)
	if err != nil {
		return errors.Wrap(err, "gcstracker.AddOrgsWithDetailsObjects")
	}
	err = tos.AddReposWithDetailsObjects(rwd)
	if err != nil {
		return errors.Wrap(err, "gcstracker.AddReposWithDetailsObjects")
	}

	gcsClient.Write(tos)
	return nil
}

// listAllGitHubReposWithDetails is a convenience wrapper around
// listGitHubReposWithDetailsPage to get ALL repos, rather than just a single
// page of them
//
// This method may return an error and a partial list of repositories
func listAllGitHubReposWithDetails(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.GitHubRepoWithDetails, error) {
	// only pull a maximum of 1,000 repos
	const perPage = 100
	const maxPage = 10
	op := *opt
	op.PerPage = perPage

	if !extgithub.HasAuthedUser(ctx) {
		return nil, nil
	}
	var allRepos []*sourcegraph.GitHubRepoWithDetails
	for page := 1; page <= maxPage; page++ {
		op.Page = page
		repos, err := listGitHubReposWithDetailsPage(ctx, &op)
		// Try to append repos to allRepos regardless of an error, as
		// listGitHubReposWithDetailsPage may have returned partial results
		allRepos = append(allRepos, repos...)
		// If an error is returned, we know that subsequent requests would be meaningless
		if err != nil {
			return allRepos, errors.Wrap(err, "tracking.listGitHubReposWithDetailsPage")
		}
		if len(repos) < perPage {
			break
		}
	}
	return allRepos, nil
}

// listGitHubReposWithDetailsPage lists repos that are accessible to the authenticated
// user, along with (1) full language details, and (2) a history of recent
// commits by time
//
// Note that this method only returns a single (paginated) API call's results
//
// If this method receives an error response from GitHub, it only returns that error if it's an
// abuse or rate limit issue, or if it has encountered more than maxRepoDetailsErrors of them.
// In those cases, it returns the data it has collected to that point
func listGitHubReposWithDetailsPage(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.GitHubRepoWithDetails, error) {
	ghRepos, _, err := extgithub.Client(ctx).Repositories.List("", opt)
	if err != nil {
		return nil, errors.Wrap(err, "Repositories.List")
	}

	var rwds []*sourcegraph.GitHubRepoWithDetails
	// Initially generate a full list of repos with basic information
	for _, ghRepo := range ghRepos {
		rwds = append(rwds, toGitHubRepoWithDetails(ghRepo))
	}

	// Next, loop through each repo again to append language and commit information. If we reach severe
	// enough or too many GitHub errors, return a partially completed list
	repoErrCounter := 0
	for i, ghRepo := range ghRepos {
		ghLanguages, _, err := extgithub.Client(ctx).Repositories.ListLanguages(*ghRepo.Owner.Login, *ghRepo.Name)
		if err != nil {
			repoErrCounter = repoErrCounter + 1
			rwds[i].ErrorFetchingDetails = true

			// Only return if error implies additional requests GitHub will fail (i.e.,
			// abuse, rate limit). Otherwise, keep trying
			if extgithub.IsRateLimitError(err) {
				return rwds, errors.Wrap(err, "Repositories.ListLanguages (GitHub rate limit exceeded)")
			}
			// Continue on to the next repository (and don't both checking for commits) if
			// the repository is unavailable for legal reasons (e.g. DMCA violations)
			if extgithub.IsLegalError(err) {
				continue
			}
			// Finally, if we've failed more than maxRepoDetailsErrors times (e.g. github.com
			// or our proxy is down), return and move on
			if repoErrCounter > maxRepoDetailsErrors {
				return rwds, errors.Wrap(err, "Repositories.ListLanguages (max errors exceeded)")
			}
		} else {
			var languages []*sourcegraph.GitHubRepoLanguage
			for k, v := range ghLanguages {
				languages = append(languages, &sourcegraph.GitHubRepoLanguage{Language: k, Count: v})
			}
			rwds[i].Languages = languages
		}

		commits, _, err := extgithub.Client(ctx).Repositories.ListCommits(*ghRepo.Owner.Login, *ghRepo.Name, nil)
		if err != nil {
			repoErrCounter = repoErrCounter + 1
			rwds[i].ErrorFetchingDetails = true

			// Only return if error implies additional requests GitHub will fail (i.e.,
			// abuse, rate limit). Otherwise, move past this repository and try the next
			if extgithub.IsRateLimitError(err) {
				return rwds, errors.Wrap(err, "Repositories.ListCommits (GitHub rate limit exceeded)")
			}
			// Finally, if we've failed more than maxRepoDetailsErrors times (e.g. github.com
			// or our proxy is down), return and move on
			if repoErrCounter > maxRepoDetailsErrors {
				return rwds, errors.Wrap(err, "Repositories.ListCommits (max errors exceeded)")
			}
		} else {
			var commitTimes []*time.Time
			for _, ghCommit := range commits {
				commitTimes = append(commitTimes, ghCommit.Commit.Committer.Date)
			}
			rwds[i].CommitTimes = commitTimes
		}
	}
	return rwds, nil
}

func toGitHubRepoWithDetails(ghrepo *github.Repository) *sourcegraph.GitHubRepoWithDetails {
	repo := extgithub.ToRepo(ghrepo)
	return &sourcegraph.GitHubRepoWithDetails{
		URI:         repo.URI,
		Owner:       repo.Owner,
		Name:        repo.Name,
		Fork:        repo.Fork,
		Private:     repo.Private,
		CreatedAt:   repo.CreatedAt,
		Languages:   make([]*sourcegraph.GitHubRepoLanguage, 0),
		CommitTimes: make([]*time.Time, 0),
	}
}
