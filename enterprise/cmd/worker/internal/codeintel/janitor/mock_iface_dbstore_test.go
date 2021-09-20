package janitor

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

// testUploadExpirerMockDBStore returns a mock DBStore instance that has default
// behaviors useful for testing the upload expirer.
func testUploadExpirerMockDBStore(
	globalPolicies []dbstore.ConfigurationPolicy,
	repositoryPolicies map[int][]dbstore.ConfigurationPolicy,
	uploads []dbstore.Upload,
) *MockDBStore {
	repositoryIDs := make([]int, 0, len(repositoryPolicies))
	for repositoryID := range repositoryPolicies {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	sort.Slice(uploads, func(i, j int) bool {
		// Ensure we return uploads in decreasing age
		return uploads[i].FinishedAt.Before(*uploads[j].FinishedAt)
	})

	state := &uploadExpirerMockStore{
		uploads:            uploads,
		repositoryIDs:      repositoryIDs,
		globalPolicies:     globalPolicies,
		repositoryPolicies: repositoryPolicies,
		protected:          map[int]time.Time{},
		expired:            map[int]struct{}{},
	}

	dbStore := NewMockDBStore()
	dbStore.SelectRepositoriesForRetentionScanFunc.SetDefaultHook(state.SelectRepositoriesForRetentionScanFunc)
	dbStore.GetConfigurationPoliciesFunc.SetDefaultHook(state.GetConfigurationPolicies)
	dbStore.GetUploadsFunc.SetDefaultHook(state.GetUploads)
	dbStore.CommitsVisibleToUploadFunc.SetDefaultHook(state.CommitsVisibleToUpload)
	dbStore.UpdateUploadRetentionFunc.SetDefaultHook(state.UpdateUploadRetention)
	return dbStore
}

type uploadExpirerMockStore struct {
	uploads            []dbstore.Upload
	repositoryIDs      []int
	globalPolicies     []dbstore.ConfigurationPolicy
	repositoryPolicies map[int][]dbstore.ConfigurationPolicy
	protected          map[int]time.Time
	expired            map[int]struct{}
}

func (s *uploadExpirerMockStore) GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
	var filtered []dbstore.Upload
	for _, upload := range s.uploads {
		if upload.RepositoryID != opts.RepositoryID {
			continue
		}
		if _, ok := s.expired[upload.ID]; ok {
			continue
		}
		if lastScanned, ok := s.protected[upload.ID]; ok && !lastScanned.Before(*opts.LastRetentionScanBefore) {
			continue
		}

		filtered = append(filtered, upload)
	}

	if len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, len(s.uploads), nil
}

func (s *uploadExpirerMockStore) CommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error) {
	for _, upload := range s.uploads {
		if upload.ID == uploadID {
			return []string{upload.Commit}, nil, nil
		}
	}

	return nil, nil, nil
}

func (s *uploadExpirerMockStore) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error {
	for _, id := range protectedIDs {
		s.protected[id] = time.Now()
	}

	for _, id := range expiredIDs {
		s.expired[id] = struct{}{}
	}

	return nil
}

func (state *uploadExpirerMockStore) SelectRepositoriesForRetentionScanFunc(ctx context.Context, processDelay time.Duration, limit int) (map[int]*time.Time, error) {
	var scannedIDs []int
	if len(state.repositoryIDs) <= limit {
		scannedIDs, state.repositoryIDs = state.repositoryIDs, nil
	} else {
		scannedIDs, state.repositoryIDs = state.repositoryIDs[:limit], state.repositoryIDs[limit:]
	}

	idsMap := map[int]*time.Time{}
	for _, id := range scannedIDs {
		idsMap[id] = nil
	}

	return idsMap, nil
}

func (state *uploadExpirerMockStore) GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, error) {
	if opts.RepositoryID == 0 {
		return state.globalPolicies, nil
	}

	policies, ok := state.repositoryPolicies[opts.RepositoryID]
	if !ok {
		return nil, errors.Errorf("unexpected repository argument %d", opts.RepositoryID)
	}

	return policies, nil
}
