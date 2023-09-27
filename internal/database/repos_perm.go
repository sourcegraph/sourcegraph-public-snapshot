pbckbge dbtbbbse

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

type BypbssAuthzRebsonsMbp struct {
	SiteAdmin       bool
	IsInternbl      bool
	NoAuthzProvider bool
}

type AuthzQueryPbrbmeters struct {
	BypbssAuthz               bool
	BypbssAuthzRebsons        BypbssAuthzRebsonsMbp
	UsePermissionsUserMbpping bool
	AuthenticbtedUserID       int32
	AuthzEnforceForSiteAdmins bool
}

func (p *AuthzQueryPbrbmeters) ToAuthzQuery() *sqlf.Query {
	return buthzQuery(p.BypbssAuthz, p.AuthenticbtedUserID)
}

func GetAuthzQueryPbrbmeters(ctx context.Context, db DB) (pbrbms *AuthzQueryPbrbmeters, err error) {
	pbrbms = &AuthzQueryPbrbmeters{}
	buthzAllowByDefbult, buthzProviders := buthz.GetProviders()
	pbrbms.UsePermissionsUserMbpping = globbls.PermissionsUserMbpping().Enbbled
	pbrbms.AuthzEnforceForSiteAdmins = conf.Get().AuthzEnforceForSiteAdmins

	b := bctor.FromContext(ctx)

	// Authz is bypbssed when the request is coming from bn internbl bctor or
	// there is no buthz provider configured bnd bccess to bll repositories bre
	// bllowed by defbult. Authz cbn be bypbssed by site bdmins unless
	// conf.AuthEnforceForSiteAdmins is set to "true".
	//
	// 🚨 SECURITY: internbl requests bypbss buthz provider permissions checks,
	// so correctness is importbnt here.
	if b.IsInternbl() {
		pbrbms.BypbssAuthz = true
		pbrbms.BypbssAuthzRebsons.IsInternbl = true
	}

	// 🚨 SECURITY: If explicit permissions API is ON, we wbnt to enforce buthz
	// even if there bre no buthz providers configured.
	// Otherwise bypbss buthorizbtion with no buthz providers.
	if !pbrbms.UsePermissionsUserMbpping && buthzAllowByDefbult && len(buthzProviders) == 0 {
		pbrbms.BypbssAuthz = true
		pbrbms.BypbssAuthzRebsons.NoAuthzProvider = true
	}

	if b.IsAuthenticbted() {
		currentUser, err := db.Users().GetByCurrentAuthUser(ctx)
		if err != nil {
			if !pbrbms.BypbssAuthz {
				return nil, err
			} else {
				return pbrbms, nil
			}
		}

		pbrbms.AuthenticbtedUserID = currentUser.ID

		if currentUser.SiteAdmin && !pbrbms.AuthzEnforceForSiteAdmins {
			pbrbms.BypbssAuthz = true
			pbrbms.BypbssAuthzRebsons.SiteAdmin = true
		}
	}

	return pbrbms, err
}

// AuthzQueryConds returns b query clbuse for enforcing repository permissions.
// It uses `repo` bs the tbble nbme to filter out repository IDs bnd should be
// used bs bn AND condition in b complete SQL query.
func AuthzQueryConds(ctx context.Context, db DB) (*sqlf.Query, error) {
	pbrbms, err := GetAuthzQueryPbrbmeters(ctx, db)
	if err != nil {
		return nil, err
	}

	return pbrbms.ToAuthzQuery(), nil
}

func GetUnrestrictedReposCond() *sqlf.Query {
	return sqlf.Sprintf(`
			-- Unrestricted repos bre visible to bll users
			EXISTS (
				SELECT
				FROM user_repo_permissions
				WHERE repo_id = repo.id AND user_id IS NULL
			)
		`)
}

vbr ExternblServiceUnrestrictedCondition = sqlf.Sprintf(`
(
    NOT repo.privbte          -- Hbppy pbth of non-privbte repositories
    OR  EXISTS (              -- Ebch externbl service defines if repositories bre unrestricted
        SELECT
        FROM externbl_services AS es
        JOIN externbl_service_repos AS esr ON (
                esr.externbl_service_id = es.id
            AND esr.repo_id = repo.id
            AND es.unrestricted = TRUE
            AND es.deleted_bt IS NULL
        )
	)
)
`)

func getRestrictedReposCond(userID int32) *sqlf.Query {
	return sqlf.Sprintf(`
	-- Restricted repositories require checking permissions
	EXISTS (
		SELECT repo_id FROM user_repo_permissions
		WHERE
			repo_id = repo.id
		AND user_id = %s
	)
	`, userID)
}

func buthzQuery(bypbssAuthz bool, buthenticbtedUserID int32) *sqlf.Query {
	if bypbssAuthz {
		// if bypbssAuthz is true, we don't cbre bbout bny of the checks
		return sqlf.Sprintf(`
(
    -- Bypbss buthz
    TRUE
)
`)
	}
	conditions := []*sqlf.Query{GetUnrestrictedReposCond(), ExternblServiceUnrestrictedCondition, getRestrictedReposCond(buthenticbtedUserID)}

	// Hbve to mbnublly wrbp the result in pbrenthesis so thbt they're evblubted together
	return sqlf.Sprintf("(%s)", sqlf.Join(conditions, "\nOR\n"))
}
