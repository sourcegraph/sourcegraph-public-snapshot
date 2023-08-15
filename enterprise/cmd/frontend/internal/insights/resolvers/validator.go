package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type InsightPermissionsValidator struct {
	insightStore   *store.InsightStore
	dashboardStore *store.DBDashboardStore
	orgStore       database.OrgStore

	once    sync.Once
	userIds []int
	orgIds  []int
	err     error

	// loaded allows the cached values to be pre-populated. This can be useful to reuse the validator in some cases
	// where these have already been loaded.
	loaded bool
}

func PermissionsValidatorFromBase(base *baseInsightResolver) *InsightPermissionsValidator {
	return &InsightPermissionsValidator{
		insightStore:   base.insightStore,
		dashboardStore: base.dashboardStore,
		orgStore:       base.postgresDB.Orgs(),
	}
}

func (v *InsightPermissionsValidator) loadUserContext(ctx context.Context) error {
	v.once.Do(func() {
		if v.loaded {
			return
		}
		userIds, orgIds, err := getUserPermissions(ctx, v.orgStore)
		if err != nil {
			v.err = errors.Wrap(err, "unable to load user permissions context")
			return
		}
		v.userIds = userIds
		v.orgIds = orgIds
		v.loaded = true
	})

	return v.err
}

func (v *InsightPermissionsValidator) validateUserAccessForDashboard(ctx context.Context, dashboardId int) error {
	err := v.loadUserContext(ctx)
	if err != nil {
		return err
	}
	hasPermission, err := v.dashboardStore.HasDashboardPermission(ctx, []int{dashboardId}, v.userIds, v.orgIds)
	if err != nil {
		return errors.Wrap(err, "HasDashboardPermissions")
	}
	// 🚨 SECURITY: if the user context doesn't get any response here that means they cannot see the dashboard.
	// The important assumption is that the store is returning only dashboards visible to the user, as well as the assumption
	// that there is no split between viewers / editors. We will return a generic not found error to prevent leaking
	// dashboard existence.
	if !hasPermission {
		return errors.New("dashboard not found")
	}

	return nil
}

func (v *InsightPermissionsValidator) validateUserAccessForView(ctx context.Context, insightId string) error {
	err := v.loadUserContext(ctx)
	if err != nil {
		return err
	}
	results, err := v.insightStore.GetAll(ctx, store.InsightQueryArgs{UniqueID: insightId, UserIDs: v.userIds, OrgIDs: v.orgIds})
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

// WithBaseStore sets the base store for any insight related stores. Used to propagate a transaction into this validator
// for permission checks against code insights tables.
func (v *InsightPermissionsValidator) WithBaseStore(base basestore.ShareableStore) *InsightPermissionsValidator {
	return &InsightPermissionsValidator{
		insightStore:   v.insightStore.With(base),
		dashboardStore: v.dashboardStore.With(base),
		orgStore:       v.orgStore,

		once:    sync.Once{},
		userIds: v.userIds,
		orgIds:  v.orgIds,
		err:     v.err,
		loaded:  v.loaded,
	}
}
