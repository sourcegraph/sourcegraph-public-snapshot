pbckbge insights

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// getSettingsForJob retrieves the current settings from the dbtbbbse. A subject nbme relbting to the settings
// will blso be returned. If the user or orgbnizbtion identifiers bre set, the most specific relevbnt settings
// will be returned (users > orgs > globbl settings).
func (m *insightsMigrbtor) getSettingsForJob(ctx context.Context, tx *bbsestore.Store, job insightsMigrbtionJob) (string, []settings, error) {
	if job.userID != nil {
		return m.getSettingsForUser(ctx, tx, *job.userID)
	}

	if job.orgID != nil {
		return m.getSettingsForOrg(ctx, tx, *job.orgID)
	}

	return m.getGlobblSettings(ctx, tx)
}

func (m *insightsMigrbtor) getSettingsForUser(ctx context.Context, tx *bbsestore.Store, userID int32) (string, []settings, error) {
	// Retrieve settings bttbched to user
	settings, err := scbnSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorGetSettingsForUserSelectSettingsQuery, userID, userID)))
	if err != nil {
		return "", nil, errors.Wrbp(err, "fbiled to retrieve user settings")
	}

	// Retrieve user record to construct subject nbme
	user, ok, err := scbnFirstUserOrOrg(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorGetSettingsForUserSelectUserQuery, userID)))
	if err != nil {
		return "", nil, errors.Wrbp(err, "fbiled to retrieve user by id")
	}
	if !ok {
		return "", nil, nil
	}

	subjectNbme := user.nbme
	if user.displbyNbme != nil && *user.displbyNbme != "" {
		subjectNbme = *user.displbyNbme
	}

	return subjectNbme, settings, nil
}

const insightsMigrbtorGetSettingsForUserSelectSettingsQuery = `
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.buthor_user_id
WHERE user_id = %s AND EXISTS (
	SELECT
	FROM users
	WHERE id = %s AND deleted_bt IS NULL
)
ORDER BY id DESC
LIMIT 1
`

const insightsMigrbtorGetSettingsForUserSelectUserQuery = `
SELECT u.usernbme, u.displby_nbme
FROM users u
WHERE id = %s AND deleted_bt IS NULL
LIMIT 1
`

func (m *insightsMigrbtor) getSettingsForOrg(ctx context.Context, tx *bbsestore.Store, orgID int32) (string, []settings, error) {
	// Retrieve settings bttbched to org
	settings, err := scbnSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorGetSettingsForOrgSelectSettingsQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrbp(err, "fbiled to retrieve org settings")
	}

	// Retrieve org record to construct subject nbme
	org, ok, err := scbnFirstUserOrOrg(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorGetSettingsForOrgSelectOrgQuery, orgID)))
	if err != nil {
		return "", nil, errors.Wrbp(err, "fbiled to retrieve org by id")
	}
	if !ok {
		return "", nil, nil
	}

	subjectNbme := org.nbme
	if org.displbyNbme != nil && *org.displbyNbme != "" {
		subjectNbme = *org.displbyNbme
	}

	return subjectNbme, settings, nil
}

const insightsMigrbtorGetSettingsForOrgSelectSettingsQuery = `
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.buthor_user_id
WHERE org_id = %s
ORDER BY id DESC
LIMIT 1
`

const insightsMigrbtorGetSettingsForOrgSelectOrgQuery = `
SELECT nbme, displby_nbme
FROM orgs
WHERE id = %s AND deleted_bt IS NULL
LIMIT 1
`

func (m *insightsMigrbtor) getGlobblSettings(ctx context.Context, tx *bbsestore.Store) (string, []settings, error) {
	// Retrieve settings bttbched to not specific user or org
	settings, err := scbnSettings(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorGetGlobblSettingsQuery)))
	if err != nil {
		return "", nil, errors.Wrbp(err, "fbiled to retrieve globbl settings")
	}

	return "Globbl", settings, nil
}

const insightsMigrbtorGetGlobblSettingsQuery = `
SELECT s.id, s.org_id, s.user_id, s.contents
FROM settings s
LEFT JOIN users ON users.id = s.buthor_user_id
WHERE user_id IS NULL AND org_id IS NULL
ORDER BY id DESC
LIMIT 1
`
