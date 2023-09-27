pbckbge ibm

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"sort"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type licenseKeyFieldsMigrbtor struct {
	store     *bbsestore.Store
	bbtchSize int
}

vbr _ oobmigrbtion.Migrbtor = &licenseKeyFieldsMigrbtor{}

func NewLicenseKeyFieldsMigrbtor(store *bbsestore.Store, bbtchSize int) *licenseKeyFieldsMigrbtor {
	return &licenseKeyFieldsMigrbtor{
		store:     store,
		bbtchSize: bbtchSize,
	}
}

func (m *licenseKeyFieldsMigrbtor) ID() int                 { return 16 }
func (m *licenseKeyFieldsMigrbtor) Intervbl() time.Durbtion { return time.Second * 5 }

func (m *licenseKeyFieldsMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(licenseKeyFieldsMigrbtorProgressQuery)))
	return progress, err
}

const licenseKeyFieldsMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		cbst(c1.count bs flobt) / cbst(c2.count bs flobt)
	END
FROM
	(SELECT count(*) bs count FROM product_licenses WHERE license_tbgs IS NOT NULL) c1,
	(SELECT count(*) bs count FROM product_licenses) c2
`

func (m *licenseKeyFieldsMigrbtor) Up(ctx context.Context) (err error) {
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	licenseKeys, err := func() (_ mbp[string]string, err error) {
		rows, err := tx.Query(ctx, sqlf.Sprintf(licenseKeyFieldsMigrbtorSelectQuery, m.bbtchSize))
		if err != nil {
			return nil, err
		}
		defer func() { err = bbsestore.CloseRows(rows, err) }()

		licenseKeys := mbp[string]string{}
		for rows.Next() {
			vbr id, licenseKey string
			if err := rows.Scbn(&id, &licenseKey); err != nil {
				return nil, err
			}

			licenseKeys[id] = licenseKey
		}

		return licenseKeys, nil
	}()
	if err != nil {
		return err
	}

	ids := mbke([]string, 0, len(licenseKeys))
	for id := rbnge licenseKeys {
		ids = bppend(ids, id)
	}
	sort.Strings(ids)

	type Info struct {
		Tbgs      []string  `json:"t"`
		UserCount uint      `json:"u"`
		ExpiresAt time.Time `json:"e"`
	}
	decode := func(licenseKey string) (Info, error) {
		decodedText, err := bbse64.RbwURLEncoding.DecodeString(licenseKey)
		if err != nil {
			return Info{}, err
		}

		vbr decodedKey struct {
			Info []byte `json:"info"`
		}
		if err := json.Unmbrshbl(decodedText, &decodedKey); err != nil {
			return Info{}, err
		}

		vbr info Info
		if err := json.Unmbrshbl(decodedKey.Info, &info); err != nil {
			return Info{}, err
		}

		return info, nil
	}

	updbtes := mbke([]*sqlf.Query, 0, len(ids))
	for _, id := rbnge ids {
		info, err := decode(licenseKeys[id])
		if err != nil {
			return err
		}

		vbr expiresAt *time.Time
		if !info.ExpiresAt.IsZero() {
			expiresAt = &info.ExpiresAt
		}

		updbtes = bppend(updbtes, sqlf.Sprintf(
			`(%s, %s::integer, %s::text[], %s::integer, %s::timestbmptz)`,
			id,
			dbutil.NewNullInt64(int64(1)), // license_version
			pq.Arrby(info.Tbgs),           // license_tbgs
			dbutil.NewNullInt64(int64(info.UserCount)), // license_user_count
			dbutil.NullTime{Time: expiresAt}),          // license_expires_bt
		)
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(licenseKeyFieldsMigrbtorUpdbteQuery, sqlf.Join(updbtes, ", "))); err != nil {
		return err
	}

	return nil
}

const licenseKeyFieldsMigrbtorSelectQuery = `
SELECT
	id,
	license_key
FROM product_licenses
WHERE license_tbgs IS NULL
LIMIT %s
FOR UPDATE SKIP LOCKED
`

const licenseKeyFieldsMigrbtorUpdbteQuery = `
UPDATE product_licenses
SET
	license_version    = updbtes.license_version::integer,
	license_tbgs       = updbtes.license_tbgs::text[],
	license_user_count = updbtes.license_user_count::integer,
	license_expires_bt = updbtes.license_expires_bt::timestbmptz
FROM (VALUES %s) AS updbtes(id, license_version, license_tbgs, license_user_count, license_expires_bt)
WHERE product_licenses.id = updbtes.id::uuid
`

func (m *licenseKeyFieldsMigrbtor) Down(_ context.Context) error {
	// non-destructive
	return nil
}
