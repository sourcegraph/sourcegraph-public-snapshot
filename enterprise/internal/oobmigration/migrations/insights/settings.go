package insights

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// getSettingsForJob retrieves the current settings from the database. A subject name relating to the settings
// will also be returned. If the user or organization identifiers are set, the most specific relevant settings
// will be returned (users > orgs > global settings).
func (m *insightsMigrator) getSettingsForJob(ctx context.Context, tx *basestore.Store, job insightsMigrationJob) (string, []settings, error) {
	if job.userID != nil {
		return m.getSettingsForUser(ctx, tx, *job.userID)
	}

	if job.orgID != nil {
		return m.getSettingsForOrg(ctx, tx, *job.orgID)
	}

	return m.getGlobalSettings(ctx, tx)
}

func (m *insightsMigrator) getSettingsForUser(ctx context.Context, tx *basestore.Store, userID int32) (string, []settings, error) {
	// Retrieve settings attached to user
	settings, err := scanSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForUserSelectSettingsQuery, userID, userID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve user settings")
	}

	// Retrieve user record to construct subject name
	user, ok, err := scanFirstUserOrOrg(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForUserSelectUserQuery, userID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve user by id")
	}
	if !ok {
		return "", nil, nil
	}

	subjectName := user.name
	if user.displayName != nil && *user.displayName != "" {
		subjectName = *user.displayName
	}

	return subjectName, settings, nil
}

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

const insightsMigratorGetSettingsForUserSelectUserQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForUser
SELECT u.username, u.display_name
FROM users u
WHERE id = %s AND deleted_at IS NULL
LIMIT 1
`

func (m *insightsMigrator) getSettingsForOrg(ctx context.Context, tx *basestore.Store, orgID int32) (string, []settings, error) {
	// Retrieve settings attached to org
	settings, err := scanSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForOrgSelectSettingsQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve org settings")
	}

	// Retrieve org record to construct subject name
	org, ok, err := scanFirstUserOrOrg(tx.Query(ctx, sqlf.Sprintf(insightsMigratorGetSettingsForOrgSelectOrgQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve org by id")
	}
	if !ok {
		return "", nil, nil
	}

	subjectName := org.name
	if org.displayName != nil && *org.displayName != "" {
		subjectName = *org.displayName
	}

	return subjectName, settings, nil
}

const insightsMigratorGetSettingsForOrgSelectSettingsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForOrg
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.author_user_id
WHERE org_id = %s
ORDER BY id DESC
LIMIT 1
`

const insightsMigratorGetSettingsForOrgSelectOrgQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/settings.go:getSettingsForOrg
SELECT name, display_name
FROM orgs
WHERE id = %s AND deleted_at IS NULL
LIMIT 1
`

func (m *insightsMigrator) getGlobalSettings(ctx context.Context, tx *basestore.Store) (string, []settings, error) {
	// Retrieve settings attached to not specific user or org
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
