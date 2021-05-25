package database

import (
	"context"
	"database/sql"
	"encoding/binary"
	"hash/fnv"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FeatureFlagStore struct {
	*basestore.Store
}

func FeatureFlags(db dbutil.DB) *FeatureFlagStore {
	return &FeatureFlagStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func FeatureFlagsWith(other basestore.ShareableStore) *FeatureFlagStore {
	return &FeatureFlagStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (f *FeatureFlagStore) With(other basestore.ShareableStore) *FeatureFlagStore {
	return &FeatureFlagStore{Store: f.Store.With(other)}
}

func (f *FeatureFlagStore) Transact(ctx context.Context) (*FeatureFlagStore, error) {
	txBase, err := f.Store.Transact(ctx)
	return &FeatureFlagStore{Store: txBase}, err
}

func (f *FeatureFlagStore) NewFeatureFlag(ctx context.Context, flag *types.FeatureFlag) (*types.FeatureFlag, error) {
	const newFeatureFlagFmtStr = `
		INSERT INTO feature_flags (
			flag_name,
			flag_type,
			bool_value,
			rollout
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING 
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		;
	`
	var (
		flagType string
		boolVal  *bool
		rollout  *int
	)
	switch {
	case flag.Bool != nil:
		flagType = "bool"
		boolVal = &flag.Bool.Value
	case flag.BoolVar != nil:
		flagType = "bool_var"
		rollout = &flag.BoolVar.Rollout
	default:
		return nil, errors.New("feature flag must have exactly one type")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFeatureFlagFmtStr,
		flag.Name,
		flagType,
		boolVal,
		rollout))
	return scanFeatureFlag(row)
}

func (f *FeatureFlagStore) NewBoolVar(ctx context.Context, name string, rollout int) (*types.FeatureFlag, error) {
	return f.NewFeatureFlag(ctx, &types.FeatureFlag{
		Name: name,
		BoolVar: &types.FeatureFlagBoolVar{
			Rollout: rollout,
		},
	})
}

func (f *FeatureFlagStore) NewBool(ctx context.Context, name string, value bool) (*types.FeatureFlag, error) {
	return f.NewFeatureFlag(ctx, &types.FeatureFlag{
		Name: name,
		Bool: &types.FeatureFlagBool{
			Value: value,
		},
	})
}

var ErrInvalidColumnState = errors.New("encountered column that is unexpectedly null based on column type")

// rowScanner is an interface that can scan from either a sql.Row or sql.Rows
type rowScanner interface {
	Scan(...interface{}) error
}

func scanFeatureFlag(scanner rowScanner) (*types.FeatureFlag, error) {
	var (
		res      types.FeatureFlag
		flagType string
		boolVal  *bool
		rollout  *int
	)
	err := scanner.Scan(
		&res.Name,
		&flagType,
		&boolVal,
		&rollout,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	switch flagType {
	case "bool":
		if boolVal == nil {
			return nil, ErrInvalidColumnState
		}
		res.Bool = &types.FeatureFlagBool{
			Value: *boolVal,
		}
	case "bool_var":
		if rollout == nil {
			return nil, ErrInvalidColumnState
		}
		res.BoolVar = &types.FeatureFlagBoolVar{
			Rollout: *rollout,
		}
	default:
		return nil, ErrInvalidColumnState
	}

	return &res, nil
}

func (f *FeatureFlagStore) ListFeatureFlags(ctx context.Context) ([]*types.FeatureFlag, error) {
	const listFeatureFlagsQuery = `
		SELECT 
			flag_name,
			flag_type,
			bool_value,
			rollout,
			created_at,
			updated_at,
			deleted_at
		FROM feature_flags
		WHERE deleted_at IS NULL;
	`

	rows, err := f.Query(ctx, sqlf.Sprintf(listFeatureFlagsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*types.FeatureFlag, 0, 10)
	for rows.Next() {
		flag, err := scanFeatureFlag(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, flag)
	}
	return res, nil
}

func (f *FeatureFlagStore) NewOverride(ctx context.Context, override *types.FeatureFlagOverride) (*types.FeatureFlagOverride, error) {
	const newFeatureFlagOverrideFmtStr = `
		INSERT INTO feature_flag_overrides (
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value;
	`
	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFeatureFlagOverrideFmtStr,
		&override.OrgID,
		&override.UserID,
		&override.FlagName,
		&override.Value))
	return scanFeatureFlagOverride(row)
}

// ListUserOverrides lists the overrides that have been specifically set for the given userID.
// NOTE: this does not return any overrides for the user orgs. Those are returned separately
// by ListOrgOverridesForUser so they can be mered in proper priority order.
func (f *FeatureFlagStore) ListUserOverrides(ctx context.Context, userID int32) ([]*types.FeatureFlagOverride, error) {
	const listUserOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE namespace_user_id = %s
			AND deleted_at IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFeatureFlagOverrides(rows)
}

// ListOrgOverridesForUser lists the feature flag overrides for all orgs the given user belongs to.
func (f *FeatureFlagStore) ListOrgOverridesForUser(ctx context.Context, userID int32) ([]*types.FeatureFlagOverride, error) {
	const listUserOverridesFmtString = `
		SELECT
			namespace_org_id,
			namespace_user_id,
			flag_name,
			flag_value
		FROM feature_flag_overrides
		WHERE EXISTS (
			SELECT org_id
			FROM org_members
			WHERE org_members.user_id = %s
				AND feature_flag_overrides.namespace_org_id = org_members.org_id
		) AND deleted_at IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFeatureFlagOverrides(rows)
}

func scanFeatureFlagOverrides(rows *sql.Rows) ([]*types.FeatureFlagOverride, error) {
	var res []*types.FeatureFlagOverride
	for rows.Next() {
		override, err := scanFeatureFlagOverride(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, override)
	}
	return res, nil
}

func scanFeatureFlagOverride(scanner rowScanner) (*types.FeatureFlagOverride, error) {
	var res types.FeatureFlagOverride
	err := scanner.Scan(
		&res.OrgID,
		&res.UserID,
		&res.FlagName,
		&res.Value,
	)
	return &res, err
}

// UserFlags returns the calculated values for feature flags for the given userID. This should
// be the primary entrypoint for getting the user flags since it handles retrieving all the flags,
// the org overrides, and the user overrides, and merges them in priority order.
func (f *FeatureFlagStore) UserFlags(ctx context.Context, userID int32) (map[string]bool, error) {
	g, ctx := errgroup.WithContext(ctx)

	var flags []*types.FeatureFlag
	g.Go(func() error {
		res, err := f.ListFeatureFlags(ctx)
		flags = res
		return err
	})

	var orgOverrides []*types.FeatureFlagOverride
	g.Go(func() error {
		res, err := f.ListOrgOverridesForUser(ctx, userID)
		orgOverrides = res
		return err
	})

	var userOverrides []*types.FeatureFlagOverride
	g.Go(func() error {
		res, err := f.ListUserOverrides(ctx, userID)
		userOverrides = res
		return err
	})

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool, max(len(flags), len(orgOverrides), len(userOverrides)))
	for _, ff := range flags {
		switch {
		case ff.Bool != nil:
			res[ff.Name] = ff.Bool.Value
		case ff.BoolVar != nil:
			res[ff.Name] = hashUserAndFlag(userID, ff.Name)%10000 < uint32(ff.BoolVar.Rollout)
		}

		// Org overrides are higher priority than default
		for _, oo := range orgOverrides {
			res[oo.FlagName] = oo.Value
		}

		// User overrides are higher priority than org overrides
		for _, uo := range userOverrides {
			res[uo.FlagName] = uo.Value
		}
	}

	return res, nil
}

func hashUserAndFlag(userID int32, flagName string) uint32 {
	h := fnv.New32()
	binary.Write(h, binary.LittleEndian, userID)
	h.Write([]byte(flagName))
	return h.Sum32()
}

// AnonymousUserFlags returns the calculated values for feature flags for the given anonymousUID
func (f *FeatureFlagStore) AnonymousUserFlags(ctx context.Context, anonymousUID string) (map[string]bool, error) {
	flags, err := f.ListFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool, len(flags))
	for _, ff := range flags {
		switch {
		case ff.Bool != nil:
			res[ff.Name] = ff.Bool.Value
		case ff.BoolVar != nil:
			res[ff.Name] = hashAnonymousUserAndFlag(anonymousUID, ff.Name)%10000 < uint32(ff.BoolVar.Rollout)
		}
	}

	return res, nil
}

func hashAnonymousUserAndFlag(anonymousUID, flagName string) uint32 {
	h := fnv.New32()
	h.Write([]byte(anonymousUID))
	h.Write([]byte(flagName))
	return h.Sum32()
}

const listUserlessFlagsFmtStr = `
SELECT
	f.flag_name,
	f.flag_type,
	f.bool_var,
FROM feature_flags f
WHERE f.deleted_at IS NULL;
`

func (f *FeatureFlagStore) UserlessFeatureFlags(ctx context.Context) (map[string]bool, error) {
	flags, err := f.ListFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]bool, len(flags))
	for _, ff := range flags {
		switch {
		case ff.Bool != nil:
			res[ff.Name] = ff.Bool.Value
		default:
			// ignore non-concrete feature flags since we have no active user
		}
	}

	return res, nil
}

func max(vals ...int) int {
	res := 0
	for _, val := range vals {
		if val > res {
			res = val
		}
	}
	return res
}
