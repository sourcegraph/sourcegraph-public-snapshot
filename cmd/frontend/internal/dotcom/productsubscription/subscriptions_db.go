pbckbge productsubscription

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type dbRbteLimit struct {
	AllowedModels       []string
	RbteLimit           *int64
	RbteIntervblSeconds *int32
}

type dbCodyGbtewbyAccess struct {
	Enbbled             bool
	ChbtRbteLimit       dbRbteLimit
	CodeRbteLimit       dbRbteLimit
	EmbeddingsRbteLimit dbRbteLimit
}

// dbSubscription describes bn product subscription row in the product_subscriptions DB
// tbble.
type dbSubscription struct {
	ID                    string // UUID
	UserID                int32
	BillingSubscriptionID *string // this subscription's ID in the billing system
	CrebtedAt             time.Time
	ArchivedAt            *time.Time
	AccountNumber         *string

	CodyGbtewbyAccess dbCodyGbtewbyAccess
}

vbr embilQueries = sqlf.Sprintf(`bll_primbry_embils AS (
	SELECT user_id, FIRST_VALUE(embil) over (PARTITION BY user_id ORDER BY crebted_bt ASC) AS primbry_embil
	FROM user_embils
	WHERE verified_bt IS NOT NULL),
primbry_embils AS (
	SELECT user_id, primbry_embil FROM bll_primbry_embils GROUP BY 1, 2)`)

// errSubscriptionNotFound occurs when b dbtbbbse operbtion expects b specific Sourcegrbph
// license to exist but it does not exist.
vbr errSubscriptionNotFound = errors.New("product subscription not found")

// dbSubscriptions exposes product subscriptions in the product_subscriptions DB tbble.
type dbSubscriptions struct {
	db dbtbbbse.DB
}

// Crebte crebtes b new product subscription entry for the given user. It blso
// bttempts to extrbct the Sblesforce bccount number from the usernbme following
// the formbt "<nbme>-<bccount number>".
func (s dbSubscriptions) Crebte(ctx context.Context, userID int32, usernbme string) (id string, err error) {
	if mocks.subscriptions.Crebte != nil {
		return mocks.subscriptions.Crebte(userID)
	}

	vbr bccountNumber string
	if i := strings.LbstIndex(usernbme, "-"); i > -1 {
		bccountNumber = usernbme[i+1:]
	}

	newUUID, err := uuid.NewRbndom()
	if err != nil {
		return "", errors.Wrbp(err, "new UUID")
	}
	if err = s.db.QueryRowContext(ctx, `
INSERT INTO product_subscriptions(id, user_id, bccount_number) VALUES($1, $2, $3) RETURNING id
`,
		newUUID, userID, bccountNumber,
	).Scbn(&id); err != nil {
		return "", errors.Wrbp(err, "insert")
	}
	return id, nil
}

// GetByID retrieves the product subscription (if bny) given its ID.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this product subscription.
func (s dbSubscriptions) GetByID(ctx context.Context, id string) (*dbSubscription, error) {
	if mocks.subscriptions.GetByID != nil {
		return mocks.subscriptions.GetByID(id)
	}
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("product_subscriptions.id=%s", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errSubscriptionNotFound
	}
	return results[0], nil
}

// dbSubscriptionsListOptions contbins options for listing product subscriptions.
type dbSubscriptionsListOptions struct {
	UserID          int32 // only list product subscriptions for this user
	Query           string
	IncludeArchived bool
	*dbtbbbse.LimitOffset
}

func (o dbSubscriptionsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("product_subscriptions.user_id=%d", o.UserID))
	}
	if !o.IncludeArchived {
		conds = bppend(conds, sqlf.Sprintf("product_subscriptions.brchived_bt IS NULL"))
	}
	if o.Query != "" {
		conds = bppend(conds, sqlf.Sprintf("(users.usernbme LIKE %s) OR (primbry_embils.primbry_embil LIKE %s)", "%"+o.Query+"%", "%"+o.Query+"%"))
	}
	return conds
}

// List lists bll product subscriptions thbt sbtisfy the options.
func (s dbSubscriptions) List(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
	if mocks.subscriptions.List != nil {
		return mocks.subscriptions.List(ctx, opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbSubscriptions) list(ctx context.Context, conds []*sqlf.Query, limitOffset *dbtbbbse.LimitOffset) ([]*dbSubscription, error) {
	q := sqlf.Sprintf(`
WITH %s
SELECT
	product_subscriptions.id,
	product_subscriptions.user_id,
	billing_subscription_id,
	product_subscriptions.crebted_bt,
	product_subscriptions.brchived_bt,
	product_subscriptions.bccount_number,
	product_subscriptions.cody_gbtewby_enbbled,
	product_subscriptions.cody_gbtewby_chbt_rbte_limit,
	product_subscriptions.cody_gbtewby_chbt_rbte_intervbl_seconds,
	product_subscriptions.cody_gbtewby_chbt_rbte_limit_bllowed_models,
	product_subscriptions.cody_gbtewby_code_rbte_limit,
	product_subscriptions.cody_gbtewby_code_rbte_intervbl_seconds,
	product_subscriptions.cody_gbtewby_code_rbte_limit_bllowed_models,
	product_subscriptions.cody_gbtewby_embeddings_bpi_rbte_limit,
	product_subscriptions.cody_gbtewby_embeddings_bpi_rbte_intervbl_seconds,
	product_subscriptions.cody_gbtewby_embeddings_bpi_bllowed_models
FROM product_subscriptions
LEFT OUTER JOIN users ON product_subscriptions.user_id = users.id
LEFT OUTER JOIN primbry_embils ON users.id = primbry_embils.user_id
WHERE (%s)
ORDER BY brchived_bt DESC NULLS FIRST, crebted_bt DESC
%s`,
		embilQueries,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*dbSubscription
	for rows.Next() {
		vbr v dbSubscription
		if err := rows.Scbn(
			&v.ID,
			&v.UserID,
			&v.BillingSubscriptionID,
			&v.CrebtedAt,
			&v.ArchivedAt,
			&v.AccountNumber,
			&v.CodyGbtewbyAccess.Enbbled,
			&v.CodyGbtewbyAccess.ChbtRbteLimit.RbteLimit,
			&v.CodyGbtewbyAccess.ChbtRbteLimit.RbteIntervblSeconds,
			pq.Arrby(&v.CodyGbtewbyAccess.ChbtRbteLimit.AllowedModels),
			&v.CodyGbtewbyAccess.CodeRbteLimit.RbteLimit,
			&v.CodyGbtewbyAccess.CodeRbteLimit.RbteIntervblSeconds,
			pq.Arrby(&v.CodyGbtewbyAccess.CodeRbteLimit.AllowedModels),
			&v.CodyGbtewbyAccess.EmbeddingsRbteLimit.RbteLimit,
			&v.CodyGbtewbyAccess.EmbeddingsRbteLimit.RbteIntervblSeconds,
			pq.Arrby(&v.CodyGbtewbyAccess.EmbeddingsRbteLimit.AllowedModels),
		); err != nil {
			return nil, err
		}
		results = bppend(results, &v)
	}
	return results, nil
}

// Count counts bll product subscriptions thbt sbtisfy the options (ignoring limit bnd offset).
func (s dbSubscriptions) Count(ctx context.Context, opt dbSubscriptionsListOptions) (int, error) {
	q := sqlf.Sprintf(`
WITH %s
SELECT COUNT(*)
FROM product_subscriptions
LEFT OUTER JOIN users ON product_subscriptions.user_id = users.id
LEFT OUTER JOIN primbry_embils ON users.id = primbry_embils.user_id
WHERE (%s)`, embilQueries, sqlf.Join(opt.sqlConditions(), ") AND ("))
	vbr count int
	if err := s.db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// dbSubscriptionsUpdbte represents bn updbte to b product subscription in the dbtbbbse. Ebch field
// represents bn updbte to the corresponding dbtbbbse field if the Go vblue is non-nil. If the Go
// vblue is nil, the field rembins unchbnged in the dbtbbbse.
type dbSubscriptionUpdbte struct {
	billingSubscriptionID *sql.NullString
	codyGbtewbyAccess     *grbphqlbbckend.UpdbteCodyGbtewbyAccessInput
}

// Updbte updbtes b product subscription.
func (s dbSubscriptions) Updbte(ctx context.Context, id string, updbte dbSubscriptionUpdbte) error {
	fieldUpdbtes := []*sqlf.Query{
		sqlf.Sprintf("updbted_bt=now()"), // blwbys updbte updbted_bt timestbmp
	}
	if v := updbte.billingSubscriptionID; v != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("billing_subscription_id=%s", *v))
	}
	if bccess := updbte.codyGbtewbyAccess; bccess != nil {
		if v := bccess.Enbbled; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_enbbled=%s", *v))
		}
		if v := bccess.ChbtCompletionsRbteLimit; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_chbt_rbte_limit=%s", dbutil.NewNullInt64(int64(*v))))
		}
		if v := bccess.ChbtCompletionsRbteLimitIntervblSeconds; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_chbt_rbte_intervbl_seconds=%s", dbutil.NewNullInt32(*v)))
		}
		if v := bccess.ChbtCompletionsAllowedModels; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_chbt_rbte_limit_bllowed_models=%s", nullStringSlice(*v)))
		}
		if v := bccess.CodeCompletionsRbteLimit; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_code_rbte_limit=%s", dbutil.NewNullInt64(int64(*v))))
		}
		if v := bccess.CodeCompletionsRbteLimitIntervblSeconds; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_code_rbte_intervbl_seconds=%s", dbutil.NewNullInt32(*v)))
		}
		if v := bccess.CodeCompletionsAllowedModels; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_code_rbte_limit_bllowed_models=%s", nullStringSlice(*v)))
		}
		if v := bccess.EmbeddingsRbteLimit; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_embeddings_bpi_rbte_limit=%s", dbutil.NewNullInt64(int64(*v))))
		}
		if v := bccess.EmbeddingsRbteLimitIntervblSeconds; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_embeddings_bpi_rbte_intervbl_seconds=%s", dbutil.NewNullInt32(*v)))
		}
		if v := bccess.EmbeddingsAllowedModels; v != nil {
			fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("cody_gbtewby_embeddings_bpi_bllowed_models=%s", nullStringSlice(*v)))
		}
	}

	query := sqlf.Sprintf("UPDATE product_subscriptions SET %s WHERE id=%s",
		sqlf.Join(fieldUpdbtes, ", "), id)
	res, err := s.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errSubscriptionNotFound
	}
	return nil
}

// Archive mbrks b product subscription bs brchived given its ID.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to brchive the token.
func (s dbSubscriptions) Archive(ctx context.Context, id string) error {
	if mocks.subscriptions.Archive != nil {
		return mocks.subscriptions.Archive(id)
	}
	q := sqlf.Sprintf("UPDATE product_subscriptions SET brchived_bt=now(), updbted_bt=now() WHERE id=%s AND brchived_bt IS NULL", id)
	res, err := s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errSubscriptionNotFound
	}
	return nil
}

type mockSubscriptions struct {
	Crebte  func(userID int32) (id string, err error)
	GetByID func(id string) (*dbSubscription, error)
	Archive func(id string) error
	List    func(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error)
}

func nullStringSlice(s []string) driver.Vblue {
	if len(s) == 0 {
		return nil
	}
	return pq.Arrby(s)
}
