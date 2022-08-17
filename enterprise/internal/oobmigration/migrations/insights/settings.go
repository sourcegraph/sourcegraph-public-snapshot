package insights

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (m *insightsMigrator) getSettings(ctx context.Context, tx *basestore.Store, userID, orgID *int) (string, []settings, error) {
	if userID != nil {
		return m.getSettingsForUser(ctx, tx, *userID)
	}

	if orgID != nil {
		return m.getSettingsForOrg(ctx, tx, *orgID)
	}

	return m.getGlobalSettings(ctx, tx)
}

func (m *insightsMigrator) getSettingsForUser(ctx context.Context, tx *basestore.Store, userID int) (string, []settings, error) {
	users, err := scanUserOrOrgs(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForUserSelectUserQuery, userID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve user by id")
	}
	if len(users) == 0 {
		return "", nil, nil
	}
	user := users[0]

	settings, err := scanSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForUserSelectSettingsQuery, userID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve user settings")
	}

	subjectName := user.name
	if user.displayName != nil && *user.displayName != "" {
		subjectName = *user.displayName
	}

	return subjectName, settings, nil
}

const insightsMigratorGetSettingsForUserSelectUserQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForUser
SELECT u.username, u.display_name
FROM users u
WHERE id = %s AND deleted_at IS NULL
LIMIT 1
`

const insightsMigratorGetSettingsForUserSelectSettingsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForUser
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.author_user_id
WHERE user_id = %s AND EXISTS (
	SELECT
	FROM users
	WHERE id = %s AND deleted_at IS NULL
)
ORDER BY id DESC
LIMIT 1
`

func (m *insightsMigrator) getSettingsForOrg(ctx context.Context, tx *basestore.Store, orgID int) (string, []settings, error) {
	orgs, err := scanUserOrOrgs(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForOrgSelectOrgQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve org by id")
	}
	if len(orgs) == 0 {
		return "", nil, nil
	}
	org := orgs[0]

	settings, err := scanSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForOrgSelectSettingsQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve org settings")
	}

	subjectName := org.name
	if org.displayName != nil && *org.displayName != "" {
		subjectName = *org.displayName
	}

	return subjectName, settings, nil
}

const insightsMigratorGetSettingsForOrgSelectOrgQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForOrg
SELECT name, display_name
FROM orgs
WHERE id = %s AND deleted_at IS NULL
LIMIT 1
`

const insightsMigratorGetSettingsForOrgSelectSettingsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForOrg
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.author_user_id
WHERE org_id = %s
ORDER BY id DESC
LIMIT 1
`

func (m *insightsMigrator) getGlobalSettings(ctx context.Context, tx *basestore.Store) (string, []settings, error) {
	settings, err := scanSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetGlobalSettingsQuery)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve global settings")
	}

	return "Global", settings, nil
}

const insightsMigratorGetGlobalSettingsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getGlobalSettings
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.author_user_id
WHERE user_id IS NULL AND org_id IS NULL
ORDER BY id DESC
LIMIT 1
`
