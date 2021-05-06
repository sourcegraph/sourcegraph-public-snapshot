package database

import (
	"context"
	"database/sql"
	"encoding/binary"
	"hash/fnv"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
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

const newBoolFmtStr = `
INSERT INTO feature_flags (
	flag_name,
	flag_type,
	value,
) VALUES (
	%s,
	'bool',
	%s
) RETURNING (
	flag_name,
	rollout,
	created_at,
	updated_at,
	deleted_at
);
`

func (f *FeatureFlagStore) NewBool(ctx context.Context, name string, value bool) (*types.FeatureFlag, error) {
	rows, err := f.Query(ctx, sqlf.Sprintf(newBoolFmtStr, name, value))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("expected a returned row")
	}
	return scanFeatureFlag(rows)
}

const newBoolVarFmtStr = `
INSERT INTO feature_flags (
	flag_name,
	flag_type,
	rollout
) VALUES (
	%s,
	'bool_var',
	%s
) RETURNING (
	flag_name,
	rollout,
	created_at,
	updated_at,
	deleted_at
);
`

func (f *FeatureFlagStore) NewBoolVar(ctx context.Context, name string, rollout int) (*types.FeatureFlag, error) {
	rows, err := f.Query(ctx, sqlf.Sprintf(newBoolVarFmtStr, name, rollout))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("expected a returned row")
	}
	return scanFeatureFlag(rows)
}

func scanFeatureFlag(rows *sql.Rows) (*types.FeatureFlag, error) {
	out := types.FeatureFlag{}
	err := rows.Scan(
		&out.Name,
		&out.Rollout,
		&out.CreatedAt,
		&out.UpdatedAt,
		&dbutil.NullTime{out.DeletedAt},
	)
	return &out, err
}

const newFeatureFlagOverrideFmtStr = `
INSERT INTO feature_flag_overrides (
	namespace_user_id,
	flag_name,
	flag_value
) VALUES (
	%s,
	%s,
	%s
);
`

func (f *FeatureFlagStore) NewUserOverride(ctx context.Context, userID int32, flagName string, flagValue bool) error {
	rows, err := f.Query(ctx, sqlf.Sprintf(newFeatureFlagOverrideFmtStr, userID, flagName, flagValue))
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

const listUserFlagsAndOverridesFmtStr = `
WITH uo AS (
	SELECT *
	FROM feature_flag_overrides
	WHERE namespace_user_id = %s
)
SELECT
	f.flag_name,
	f.flag_type,
	f.bool_value,
	f.rollout,
	uo.flag_value AS user_override
FROM feature_flags f
LEFT JOIN uo ON f.flag_name = uo.flag_name
`

// UserFlags returns the calculated values for feature flags for the given userID
func (f *FeatureFlagStore) UserFlags(ctx context.Context, userID int32) (map[string]bool, error) {
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserFlagsAndOverridesFmtStr, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]bool, 10) // guess on size to avoid small allocations
	for rows.Next() {
		var (
			name          string
			flag_type     string
			bool_value    *bool
			rollout       *int
			user_override *bool
		)

		if err := rows.Scan(&name, &flag_type, &bool_value, &rollout, &user_override); err != nil {
			return nil, err
		}

		if user_override != nil {
			res[name] = *user_override
			continue
		}

		switch flag_type {
		case "bool":
			if bool_value != nil {
				// This should always be non-nil based on table constraints
				res[name] = *bool_value
			}
		case "bool_var":
			if rollout != nil {
				// This should always be non-nil based on table constraints
				res[name] = hashUserAndFlag(userID, name)%10000 < uint32(*rollout)
			}
		default:
			panic("unknown flag type")
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

const listAnonymousUserFlagsFmtStr = `
SELECT
	f.flag_name,
	f.flag_type,
	f.bool_value,
	f.rollout,
FROM feature_flags f
`

// AnonymousUserFlags returns the calculated values for feature flags for the given anonymousUID
func (f *FeatureFlagStore) AnonymousUserFlags(ctx context.Context, anonymousUID string) (map[string]bool, error) {
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserFlagsAndOverridesFmtStr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]bool, 10) // guess on size to avoid small allocations
	for rows.Next() {
		var (
			name       string
			flag_type  string
			bool_value *bool
			rollout    *int
		)

		if err := rows.Scan(&name, &flag_type, &bool_value, &rollout); err != nil {
			return nil, err
		}

		switch flag_type {
		case "bool":
			if bool_value != nil {
				// This should always be non-nil based on table constraints
				res[name] = *bool_value
			}
		case "bool_var":
			if rollout != nil {
				// This should always be non-nil based on table constraints
				res[name] = hashAnonymousUserAndFlag(anonymousUID, name)%10000 < uint32(*rollout)
			}
		default:
			panic("unknown flag type")
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
	f.bool_value,
FROM feature_flags f
`

func (f *FeatureFlagStore) UserlessFeatureFlags(ctx context.Context) (map[string]bool, error) {
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserlessFlagsFmtStr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]bool, 10) // guess on size to avoid small allocations
	for rows.Next() {
		var (
			name       string
			flag_type  string
			bool_value *bool
		)

		if err := rows.Scan(&name, &flag_type, &bool_value); err != nil {
			return nil, err
		}

		switch flag_type {
		case "bool":
			if bool_value != nil {
				// This should always be non-nil based on table constraints
				res[name] = *bool_value
			}
		default:
			// Ignore non-concrete flags since we don't have a user
		}
	}

	return res, nil

}
