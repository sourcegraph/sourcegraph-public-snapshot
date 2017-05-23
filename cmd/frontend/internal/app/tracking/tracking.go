package tracking

import (
	"context"
	"encoding/json"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/gcstracker"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
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

// TrackUserGitHubData handles user data logging during auth flows
//
// Specifically, fetching limited information about
// a user's GitHub profile and sending it to Google Cloud Storage
// for analytics, as well as updating user data properties in HubSpot
func TrackUserGitHubData(a *actor.Actor, event string, name string, company string, location string, webSessionID string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.TrackUserGitHubData: %s", err)
		}
	}()

	// If the user is in a dev environment, don't do any data pulls from GitHub, or any tracking
	if env.Version == "dev" {
		return
	}

	// Generate a single set of user parameters for HubSpot and Slack exports
	contactParams := &hubspot.ContactProperties{
		UserID:            a.Login,
		UID:               a.UID,
		GitHubLink:        "https://github.com/" + a.Login,
		LookerLink:        "https://sourcegraph.looker.com/dashboards/9?User%20ID=" + a.Login,
		GitHubName:        name,
		GitHubCompany:     company,
		GitHubLocation:    location,
		IsPrivateCodeUser: false,
	}
	for _, v := range a.GitHubScopes {
		if v == "repo" {
			contactParams.IsPrivateCodeUser = true
			break
		}
	}

	// Update or create user contact information in HubSpot
	hsResponse, err := trackHubSpotContact(a.Email, event, contactParams)
	if err != nil {
		log15.Warn("trackHubSpotContact: failed to create or update HubSpot contact on auth", "source", "HubSpot", "error", err)
	}

	gcsClient, err := gcstracker.New(a, webSessionID)
	if err != nil {
		log15.Error("Error creating a new GCS client", "error", err)
		return
	}

	// Since the newly-authenticated actor (and their GitHubToken) has
	// not yet been associated with the request's context, we need to
	// create a temporary Context object that contains that linkage in
	// order to pull data from the GitHub API
	tempCtx := actor.WithActor(context.Background(), a)

	// Fetch orgs and org members data
	// ListAllOrgs may return partial results
	orgList, err := graphqlbackend.ListAllOrgs(tempCtx, &sourcegraph.ListOptions{})
	if err != nil {
		log15.Warn("graphqlbackend.ListAllOrgs: failed to fetch some user organizations", "source", "GitHub", "error", err)
	}

	orgMembersErrCounter := 0
	owd := make(map[string]([]*github.User))
	for _, org := range orgList.Orgs {
		members, err := graphqlbackend.ListAllOrgMembers(tempCtx, &sourcegraph.ListMembersOptions{OrgName: org.Login})
		if err != nil {
			// ListAllOrgMembers may return partial results
			// Don't give up unless maxOrgMemberErrors errors are caught
			orgMembersErrCounter = orgMembersErrCounter + 1
			if orgMembersErrCounter > maxOrgMemberErrors {
				log15.Warn("graphqlbackend.ListAllOrgMembers: failed to fetch some user org members (max errors exceeded)", "source", "GitHub", "error", err)
				break
			} else {
				log15.Warn("graphqlbackend.ListAllOrgMembers: failed to fetch some user org members", "source", "GitHub", "error", err)
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

	tos.AddOrgsWithDetailsObjects(owd)
	tos.AddReposWithDetailsObjects(rwd)
	err = gcsClient.Write(tos)
	if err != nil {
		log15.Error("Error writing to GCS", "error", err)
		return
	}

	// Finally, post signup notification to Slack
	if event == "SignupCompleted" {
		err = notifySlackOnSignup(a, contactParams, hsResponse, tos)
		if err != nil {
			log15.Error("Error sending new signup details to Slack", "error", err)
			return
		}
	}
}

// listAllGitHubReposWithDetails is a convenience wrapper around
// listGitHubReposWithDetailsPage to get ALL repos, rather than just a single
// page of them
//
// This method may return an error and a partial list of repositories
func listAllGitHubReposWithDetails(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.GitHubRepoWithDetails, error) {
	// only pull a maximum of 500 repos
	const perPage = 100
	const maxPage = 5
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
		// If the repo is uninteresting for being old or not having a primary language, skip it
		if time.Since(ghRepo.PushedAt.Time).Hours() > 24*365 || ghRepo.Language == nil {
			rwds[i].Skipped = true
			continue
		}

		ghLanguages, resp, err := extgithub.Client(ctx).Repositories.ListLanguages(*ghRepo.Owner.Login, *ghRepo.Name)
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
			if resp.StatusCode == 451 {
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
		PushedAt:    repo.PushedAt,
		Languages:   make([]*sourcegraph.GitHubRepoLanguage, 0),
		CommitTimes: make([]*time.Time, 0),
	}
}

func trackHubSpotContact(email string, eventLabel string, params *hubspot.ContactProperties) (*hubspot.ContactResponse, error) {
	if email == "" {
		return nil, errors.New("User must have a valid email address.")
	}

	c, err := hubspotutil.Client()
	if err != nil {
		return nil, errors.Wrap(err, "hubspotutil.Client")
	}

	if eventLabel == "SignupCompleted" {
		today := time.Now().Truncate(24 * time.Hour)
		// Convert to milliseconds
		params.RegisteredAt = today.UTC().Unix() * 1000
	}

	// Create or update the contact
	resp, err := c.CreateOrUpdateContact(email, params)
	if err != nil {
		return nil, err
	}

	// Log the event if relevant (in this case, for "SignupCompleted" events)
	if _, ok := hubspotutil.EventNameToHubSpotID[eventLabel]; ok {
		err := c.LogEvent(email, hubspotutil.EventNameToHubSpotID[eventLabel], map[string]string{})
		if err != nil {
			return nil, errors.Wrap(err, "LogEvent")
		}
	}

	// Parse response
	hsResponse := &hubspot.ContactResponse{}
	err = json.Unmarshal(resp, hsResponse)
	if err != nil {
		return nil, err
	}

	return hsResponse, nil
}
