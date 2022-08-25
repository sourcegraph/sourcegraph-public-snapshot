package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockStatusMessages func(context.Context, *types.User) ([]StatusMessage, error)

// FetchStatusMessages fetches repo related status messages. When fetching
// external service sync errors we'll fetch any external services owned by the
// user. In addition, if the user is a site admin we'll also fetch site level
// external services.
func FetchStatusMessages(ctx context.Context, db database.DB, u *types.User) ([]StatusMessage, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx, u)
	}
	if u == nil {
		return nil, errors.New("nil user")
	}
	var messages []StatusMessage

	// We first fetch affiliated sync errors since this will also find all the
	// external services the user cares about.
	externalServiceSyncErrors, err := db.ExternalServices().GetAffiliatedSyncErrors(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}
	// Return early since we don't have any affiliated external services
	if len(externalServiceSyncErrors) == 0 {
		return messages, nil
	}

	extsvcIDs := make([]int64, 0, len(externalServiceSyncErrors))

	for id, failure := range externalServiceSyncErrors {
		extsvcIDs = append(extsvcIDs, id)

		if failure != "" {
			messages = append(messages, StatusMessage{
				ExternalServiceSyncError: &ExternalServiceSyncError{
					Message:           failure,
					ExternalServiceId: id,
				},
			})
		}
	}

	// If the user is not a site-admin we can't rely on the stats table for
	// counts and need to query the repo/gitserver_repos tables. But since the
	// COUNTs aren't cheap, we filter down and use limit=1 to check whether >=1
	// repos with the given parameters exist.
	if !u.SiteAdmin {
		opts := database.ReposListOptions{
			NoCloned:           true,
			ExternalServiceIDs: extsvcIDs,
			LimitOffset: &database.LimitOffset{
				Limit: 1,
			},
		}
		notCloned, err := db.Repos().ListMinimalRepos(ctx, opts)
		if err != nil {
			return nil, errors.Wrap(err, "listing not-cloned repos")
		}
		if len(notCloned) > 0 {
			messages = append(messages, StatusMessage{
				Cloning: &CloningProgress{
					Message: "Some repositories cloning...",
				},
			})
		}

		// Look for any repository that we could not sync
		opts = database.ReposListOptions{
			FailedFetch:        true,
			ExternalServiceIDs: extsvcIDs,
			LimitOffset: &database.LimitOffset{
				Limit: 1,
			},
		}
		failedSync, err := db.Repos().ListMinimalRepos(ctx, opts)
		if err != nil {
			return nil, errors.Wrap(err, "counting repo sync failures")
		}
		if len(failedSync) > 0 {
			messages = append(messages, StatusMessage{
				SyncError: &SyncError{
					Message: "Some repositories could not be synced",
				},
			})
		}
		return messages, nil
	}

	// If the user is a site-admin we assume they can see all repositories
	// (since authz was only enforced for admins in the old sourcegraph-dot-com
	// Cloud v1 model).
	stats, err := db.RepoStatistics().GetRepoStatistics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "loading repo statistics")
	}

	if uncloned := stats.NotCloned + stats.Cloning; uncloned > 0 {
		var sentences []string
		if stats.NotCloned > 0 {
			sentences = append(sentences, fmt.Sprintf("%d %s enqueued for cloning.", stats.NotCloned, pluralize(stats.NotCloned, "repository", "repositories")))
		}
		if stats.Cloning > 0 {
			sentences = append(sentences, fmt.Sprintf("%d %s currently cloning...", stats.Cloning, pluralize(stats.Cloning, "repository", "repositories")))
		}
		messages = append(messages, StatusMessage{
			Cloning: &CloningProgress{
				Message: strings.Join(sentences, " "),
			},
		})
	}

	if stats.FailedFetch > 0 {
		messages = append(messages, StatusMessage{
			SyncError: &SyncError{
				Message: fmt.Sprintf("%d %s could not be synced", stats.FailedFetch, pluralize(stats.FailedFetch, "repository", "repositories")),
			},
		})
	}

	return messages, nil
}

func pluralize(count int, singularNoun, pluralNoun string) string {
	if count == 1 {
		return singularNoun
	}
	return pluralNoun
}

type CloningProgress struct {
	Message string
}

type IndexingProgress struct {
	Message string
}

type ExternalServiceSyncError struct {
	Message           string
	ExternalServiceId int64
}

type SyncError struct {
	Message string
}

type IndexingError struct {
	Message string
}

type StatusMessage struct {
	Cloning                  *CloningProgress          `json:"cloning"`
	Indexing                 *IndexingProgress         `json:"indexing"`
	ExternalServiceSyncError *ExternalServiceSyncError `json:"external_service_sync_error"`
	SyncError                *SyncError                `json:"sync_error"`
	IndexingError            *IndexingError            `json:"indexing_error"`
}
