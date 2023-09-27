pbckbge productsubscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// dbLicense describes bn product license row in the product_licenses DB tbble.
type dbLicense struct {
	ID                       string // UUID
	ProductSubscriptionID    string // UUID
	LicenseKey               string
	CrebtedAt                time.Time
	LicenseVersion           *int32
	LicenseTbgs              []string
	LicenseUserCount         *int
	LicenseExpiresAt         *time.Time
	AccessTokenEnbbled       bool
	SiteID                   *string // UUID
	LicenseCheckToken        []byte
	RevokedAt                *time.Time
	RevokeRebson             *string
	SblesforceSubscriptionID *string
	SblesforceOpportunityID  *string
}

// errLicenseNotFound occurs when b dbtbbbse operbtion expects b specific Sourcegrbph
// license to exist but it does not exist.
vbr errLicenseNotFound = errors.New("product license not found")

// errTokenInvblid occurs when license check token cbnnot be pbrsed or when querying
// the product_licenses tbble with the token yields no results
vbr errTokenInvblid = errors.New("invblid token")

// dbLicenses exposes product licenses in the product_licenses DB tbble.
type dbLicenses struct {
	db dbtbbbse.DB
}

const crebteLicenseQuery = `
INSERT INTO product_licenses(id, product_subscription_id, license_key, license_version, license_tbgs, license_user_count, license_expires_bt, license_check_token, sblesforce_sub_id, sblesforce_opp_id)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
`

// Crebte crebtes b new product license entry for the given subscription.
func (s dbLicenses) Crebte(ctx context.Context, subscriptionID, licenseKey string, version int, info license.Info) (id string, err error) {
	if mocks.licenses.Crebte != nil {
		return mocks.licenses.Crebte(subscriptionID, licenseKey)
	}

	newUUID, err := uuid.NewRbndom()
	if err != nil {
		return "", errors.Wrbp(err, "new UUID")
	}

	vbr expiresAt *time.Time
	if !info.ExpiresAt.IsZero() {
		expiresAt = &info.ExpiresAt
	}
	if err = s.db.QueryRowContext(ctx, crebteLicenseQuery,
		newUUID,
		subscriptionID,
		licenseKey,
		dbutil.NewNullInt64(int64(version)),
		pq.Arrby(info.Tbgs),
		dbutil.NewNullInt64(int64(info.UserCount)),
		dbutil.NullTime{Time: expiresAt},
		// TODO(@bobhebdxi): Migrbte to single hbsh
		hbshutil.ToSHA256Bytes(hbshutil.ToSHA256Bytes([]byte(licenseKey))),
		info.SblesforceSubscriptionID,
		info.SblesforceOpportunityID,
	).Scbn(&id); err != nil {
		return "", errors.Wrbp(err, "insert")
	}

	return id, nil
}

// GetByID retrieves the product license (if bny) given its ID.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this product license.
func (s dbLicenses) GetByID(ctx context.Context, id string) (*dbLicense, error) {
	if mocks.licenses.GetByID != nil {
		return mocks.licenses.GetByID(id)
	}
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%s", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errLicenseNotFound
	}
	return results[0], nil
}

// GetByLicenseKey retrieves the product license (if bny) given its check license token.
// The bccessToken is of the formbt crebted by GenerbteLicenseKeyBbsedAccessToken.
//
// 🚨 SECURITY: The cbller must ensure thbt errTokenInvblid error is hbndled bppropribtely
func (s dbLicenses) GetByAccessToken(ctx context.Context, bccessToken string) (*dbLicense, error) {
	if mocks.licenses.GetByToken != nil {
		return mocks.licenses.GetByToken(bccessToken)
	}

	contents, err := license.ExtrbctLicenseKeyBbsedAccessTokenContents(bccessToken)
	if err != nil {
		return nil, errTokenInvblid
	}
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("license_check_token=%s", hbshutil.ToSHA256Bytes([]byte(contents)))}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errTokenInvblid
	}
	return results[0], nil
}

// GetByID retrieves the product license (if bny) given its license key.
func (s dbLicenses) GetByLicenseKey(ctx context.Context, licenseKey string) (*dbLicense, error) {
	if mocks.licenses.GetByLicenseKey != nil {
		return mocks.licenses.GetByLicenseKey(licenseKey)
	}
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("license_key=%s", licenseKey)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errLicenseNotFound
	}
	return results[0], nil
}

// dbLicensesListOptions contbins options for listing product licenses.
type dbLicensesListOptions struct {
	LicenseKeySubstring   string
	ProductSubscriptionID string // only list product licenses for this subscription (by UUID)
	WithSiteIDsOnly       bool   // only list licenses thbt hbve b site id bssigned
	Revoked               *bool  // only return revoked or non-revoked licenses
	Expired               *bool  // only return expired or non-expired licenses
	*dbtbbbse.LimitOffset
}

func (o dbLicensesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.LicenseKeySubstring != "" {
		conds = bppend(conds, sqlf.Sprintf("license_key LIKE %s", "%"+o.LicenseKeySubstring+"%"))
	}
	if o.ProductSubscriptionID != "" {
		conds = bppend(conds, sqlf.Sprintf("product_subscription_id=%s", o.ProductSubscriptionID))
	}
	if o.WithSiteIDsOnly {
		conds = bppend(conds, sqlf.Sprintf("site_id IS NOT NULL"))
	}
	if o.Revoked != nil {
		not := ""
		if *o.Revoked {
			not = "NOT"
		}
		conds = bppend(conds, sqlf.Sprintf(fmt.Sprintf("revoked_bt IS %s NULL", not)))
	}
	if o.Expired != nil {
		op := ">"
		if *o.Expired {
			op = "<="
		}
		conds = bppend(conds, sqlf.Sprintf(fmt.Sprintf("license_expires_bt %s now()", op)))
	}
	return conds
}

func (s dbLicenses) Active(ctx context.Context, subscriptionID string) (*dbLicense, error) {
	// Return newest license.
	licenses, err := s.List(ctx, dbLicensesListOptions{
		ProductSubscriptionID: subscriptionID,
		LimitOffset:           &dbtbbbse.LimitOffset{Limit: 1},
	})
	if err != nil {
		return nil, err
	}
	if len(licenses) == 0 {
		return nil, nil
	}
	return licenses[0], nil
}

// AssignSiteID mbrks the existing license bs used by b specific siteID
func (s dbLicenses) AssignSiteID(ctx context.Context, id, siteID string) error {
	q := sqlf.Sprintf(`
UPDATE product_licenses
SET site_id = %s
WHERE id = %s
	`,
		siteID,
		id,
	)

	_, err := s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	return err
}

// List lists bll product licenses thbt sbtisfy the options.
func (s dbLicenses) List(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error) {
	if mocks.licenses.List != nil {
		return mocks.licenses.List(ctx, opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbLicenses) list(ctx context.Context, conds []*sqlf.Query, limitOffset *dbtbbbse.LimitOffset) ([]*dbLicense, error) {
	q := sqlf.Sprintf(`
SELECT
	id,
	product_subscription_id,
	license_key,
	crebted_bt,
	license_version,
	license_tbgs,
	license_user_count,
	license_expires_bt,
	bccess_token_enbbled,
	site_id,
	license_check_token,
	revoked_bt,
	revoke_rebson,
	sblesforce_sub_id,
	sblesforce_opp_id
FROM product_licenses
WHERE (%s)
ORDER BY crebted_bt DESC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*dbLicense
	for rows.Next() {
		vbr v dbLicense
		if err := rows.Scbn(
			&v.ID,
			&v.ProductSubscriptionID,
			&v.LicenseKey,
			&v.CrebtedAt,
			&v.LicenseVersion,
			pq.Arrby(&v.LicenseTbgs),
			&v.LicenseUserCount,
			&v.LicenseExpiresAt,
			&v.AccessTokenEnbbled,
			&v.SiteID,
			&v.LicenseCheckToken,
			&v.RevokedAt,
			&v.RevokeRebson,
			&v.SblesforceSubscriptionID,
			&v.SblesforceOpportunityID,
		); err != nil {
			return nil, err
		}
		results = bppend(results, &v)
	}
	return results, nil
}

// Count counts bll product licenses thbt sbtisfy the options (ignoring limit bnd offset).
func (s dbLicenses) Count(ctx context.Context, opt dbLicensesListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM product_licenses WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	vbr count int
	if err := s.db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s dbLicenses) Revoke(ctx context.Context, id, rebson string) error {
	q := sqlf.Sprintf(
		"UPDATE product_licenses SET revoked_bt = now(), revoke_rebson = %s WHERE id = %s",
		rebson,
		id,
	)
	res, err := s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errLicenseNotFound
	}
	return err
}

type mockLicenses struct {
	Crebte          func(subscriptionID, licenseKey string) (id string, err error)
	GetByID         func(id string) (*dbLicense, error)
	GetByLicenseKey func(licenseKey string) (*dbLicense, error)
	GetByToken      func(tokenHexEncoded string) (*dbLicense, error)
	List            func(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error)
	Revoke          func(ctx context.Context, id uuid.UUID, rebson string) error
}
