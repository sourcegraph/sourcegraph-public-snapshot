package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type InsightPermissionsValidator struct {
	baseInsightResolver

	once    sync.Once
	userIds []int
	orgIds  []int
	err     error
}

func (v *InsightPermissionsValidator) loadUserContext(ctx context.Context) error {
	v.once.Do(func() {
		userIds, orgIds, err := getUserPermissions(ctx, database.Orgs(v.postgresDB))
		if err != nil {
			v.err = errors.Wrap(err, "unable to load user permissions context")
			return
		}
		v.userIds = userIds
		v.orgIds = orgIds
	})

	return v.err
}

func (v *InsightPermissionsValidator) validateUserAccessForView(ctx context.Context, insightId string) error {
	err := v.loadUserContext(ctx)
	if err != nil {
		return err
	}
	results, err := v.insightStore.GetAll(ctx, store.InsightQueryArgs{UniqueID: insightId, UserID: v.userIds, OrgID: v.orgIds})
	if err != nil {
		return errors.Wrap(err, "GetAll")
	}
	// 🚨 SECURITY: if the user context doesn't get any response here that means they cannot see the insight.
	// The important assumption is that the store is returning only insights visible to the user, as well as the assumption
	// that there is no split between viewers / editors. We will return a generic not found error to prevent leaking
	// insight existence.
	if len(results) == 0 {
		return errors.New("insight not found")
	}

	return nil
}
