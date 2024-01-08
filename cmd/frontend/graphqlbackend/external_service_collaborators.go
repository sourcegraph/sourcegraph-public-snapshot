package graphqlbackend

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func (r *invitableCollaboratorResolver) OwnerField() string {
	return ""
}

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
	sort.SliceStable(recentCommitters, func(i, j int) bool {
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
	sort.SliceStable(invitable, func(i, j int) bool {
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
