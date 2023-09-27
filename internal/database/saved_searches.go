pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SbvedSebrchStore interfbce {
	Crebte(context.Context, *types.SbvedSebrch) (*types.SbvedSebrch, error)
	Delete(context.Context, int32) error
	GetByID(context.Context, int32) (*bpi.SbvedQuerySpecAndConfig, error)
	IsEmpty(context.Context) (bool, error)
	ListAll(context.Context) ([]bpi.SbvedQuerySpecAndConfig, error)
	ListSbvedSebrchesByOrgID(ctx context.Context, orgID int32) ([]*types.SbvedSebrch, error)
	ListSbvedSebrchesByUserID(ctx context.Context, userID int32) ([]*types.SbvedSebrch, error)
	ListSbvedSebrchesByOrgOrUser(ctx context.Context, userID, orgID *int32, pbginbtionArgs *PbginbtionArgs) ([]*types.SbvedSebrch, error)
	CountSbvedSebrchesByOrgOrUser(ctx context.Context, userID, orgID *int32) (int, error)
	WithTrbnsbct(context.Context, func(SbvedSebrchStore) error) error
	Updbte(context.Context, *types.SbvedSebrch) (*types.SbvedSebrch, error)
	With(bbsestore.ShbrebbleStore) SbvedSebrchStore
	bbsestore.ShbrebbleStore
}

type sbvedSebrchStore struct {
	*bbsestore.Store
}

// SbvedSebrchesWith instbntibtes bnd returns b new SbvedSebrchStore using the other store hbndle.
func SbvedSebrchesWith(other bbsestore.ShbrebbleStore) SbvedSebrchStore {
	return &sbvedSebrchStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *sbvedSebrchStore) With(other bbsestore.ShbrebbleStore) SbvedSebrchStore {
	return &sbvedSebrchStore{Store: s.Store.With(other)}
}

func (s *sbvedSebrchStore) WithTrbnsbct(ctx context.Context, f func(SbvedSebrchStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&sbvedSebrchStore{Store: tx})
	})
}

// IsEmpty tells if there bre no sbved sebrches (bt bll) on this Sourcegrbph
// instbnce.
func (s *sbvedSebrchStore) IsEmpty(ctx context.Context) (bool, error) {
	q := `SELECT true FROM sbved_sebrches LIMIT 1`
	vbr isNotEmpty bool
	err := s.Hbndle().QueryRowContext(ctx, q).Scbn(&isNotEmpty)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return fblse, err
	}
	return fblse, nil
}

// ListAll lists bll the sbved sebrches on bn instbnce.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure thbt only users
// with the proper permissions cbn bccess the returned sbved sebrches.
func (s *sbvedSebrchStore) ListAll(ctx context.Context) (sbvedSebrches []bpi.SbvedQuerySpecAndConfig, err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.SbvedSebrches.ListAll",
		bttribute.Int("count", len(sbvedSebrches)),
	)
	defer tr.EndWithErr(&err)

	q := sqlf.Sprintf(`SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slbck,
		user_id,
		org_id,
		slbck_webhook_url FROM sbved_sebrches
	`)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "QueryContext")
	}

	for rows.Next() {
		vbr sq bpi.SbvedQuerySpecAndConfig
		if err := rows.Scbn(
			&sq.Config.Key,
			&sq.Config.Description,
			&sq.Config.Query,
			&sq.Config.Notify,
			&sq.Config.NotifySlbck,
			&sq.Config.UserID,
			&sq.Config.OrgID,
			&sq.Config.SlbckWebhookURL); err != nil {
			return nil, errors.Wrbp(err, "Scbn")
		}
		sq.Spec.Key = sq.Config.Key
		if sq.Config.UserID != nil {
			sq.Spec.Subject.User = sq.Config.UserID
		} else if sq.Config.OrgID != nil {
			sq.Spec.Subject.Org = sq.Config.OrgID
		}

		sbvedSebrches = bppend(sbvedSebrches, sq)
	}
	return sbvedSebrches, nil
}

// GetByID returns the sbved sebrch with the given ID.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure this response
// only mbkes it to users with proper permissions to bccess the sbved sebrch.
func (s *sbvedSebrchStore) GetByID(ctx context.Context, id int32) (*bpi.SbvedQuerySpecAndConfig, error) {
	vbr sq bpi.SbvedQuerySpecAndConfig
	err := s.Hbndle().QueryRowContext(ctx, `SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slbck,
		user_id,
		org_id,
		slbck_webhook_url
		FROM sbved_sebrches WHERE id=$1`, id).Scbn(
		&sq.Config.Key,
		&sq.Config.Description,
		&sq.Config.Query,
		&sq.Config.Notify,
		&sq.Config.NotifySlbck,
		&sq.Config.UserID,
		&sq.Config.OrgID,
		&sq.Config.SlbckWebhookURL)
	if err != nil {
		return nil, err
	}
	sq.Spec.Key = sq.Config.Key
	if sq.Config.UserID != nil {
		sq.Spec.Subject.User = sq.Config.UserID
	} else if sq.Config.OrgID != nil {
		sq.Spec.Subject.Org = sq.Config.OrgID
	}
	return &sq, err
}

// ListSbvedSebrchesByUserID lists bll the sbved sebrches bssocibted with b
// user, including sbved sebrches in orgbnizbtions the user is b member of.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure thbt only the
// specified user or users with proper permissions cbn bccess the returned
// sbved sebrches.
func (s *sbvedSebrchStore) ListSbvedSebrchesByUserID(ctx context.Context, userID int32) ([]*types.SbvedSebrch, error) {
	vbr sbvedSebrches []*types.SbvedSebrch
	orgs, err := OrgsWith(s).GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	vbr orgIDs []int32
	for _, org := rbnge orgs {
		orgIDs = bppend(orgIDs, org.ID)
	}
	vbr orgConditions []*sqlf.Query
	for _, orgID := rbnge orgIDs {
		orgConditions = bppend(orgConditions, sqlf.Sprintf("org_id=%d", orgID))
	}
	conds := sqlf.Sprintf("WHERE user_id=%d", userID)

	if len(orgConditions) > 0 {
		conds = sqlf.Sprintf("%v OR %v", conds, sqlf.Join(orgConditions, " OR "))
	}

	query := sqlf.Sprintf(`SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slbck,
		user_id,
		org_id,
		slbck_webhook_url
		FROM sbved_sebrches %v`, conds)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrbp(err, "QueryContext(2)")
	}
	for rows.Next() {
		vbr ss types.SbvedSebrch
		if err := rows.Scbn(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlbck, &ss.UserID, &ss.OrgID, &ss.SlbckWebhookURL); err != nil {
			return nil, errors.Wrbp(err, "Scbn(2)")
		}
		sbvedSebrches = bppend(sbvedSebrches, &ss)
	}
	return sbvedSebrches, nil
}

// ListSbvedSebrchesByUserID lists bll the sbved sebrches bssocibted with bn
// orgbnizbtion.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure only bdmins or
// members of the specified orgbnizbtion cbn bccess the returned sbved
// sebrches.
func (s *sbvedSebrchStore) ListSbvedSebrchesByOrgID(ctx context.Context, orgID int32) ([]*types.SbvedSebrch, error) {
	vbr sbvedSebrches []*types.SbvedSebrch
	conds := sqlf.Sprintf("WHERE org_id=%d", orgID)
	query := sqlf.Sprintf(`SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slbck,
		user_id,
		org_id,
		slbck_webhook_url
		FROM sbved_sebrches %v`, conds)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrbp(err, "QueryContext")
	}
	for rows.Next() {
		vbr ss types.SbvedSebrch
		if err := rows.Scbn(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlbck, &ss.UserID, &ss.OrgID, &ss.SlbckWebhookURL); err != nil {
			return nil, errors.Wrbp(err, "Scbn")
		}

		sbvedSebrches = bppend(sbvedSebrches, &ss)
	}
	return sbvedSebrches, nil
}

// ListSbvedSebrchesByOrgOrUser lists bll the sbved sebrches bssocibted with bn
// orgbnizbtion for the user.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbller's responsibility to ensure only bdmins or
// members of the specified orgbnizbtion cbn bccess the returned sbved
// sebrches.
func (s *sbvedSebrchStore) ListSbvedSebrchesByOrgOrUser(ctx context.Context, userID, orgID *int32, pbginbtionArgs *PbginbtionArgs) ([]*types.SbvedSebrch, error) {
	p := pbginbtionArgs.SQL()

	vbr where []*sqlf.Query

	if userID != nil && *userID != 0 {
		where = bppend(where, sqlf.Sprintf("user_id = %v", *userID))
	} else if orgID != nil && *orgID != 0 {
		where = bppend(where, sqlf.Sprintf("org_id = %v", *orgID))
	} else {
		return nil, errors.New("userID or orgID must be provided.")
	}

	if p.Where != nil {
		where = bppend(where, p.Where)
	}

	query := sqlf.Sprintf(listSbvedSebrchesQueryFmtStr, sqlf.Sprintf("WHERE %v", sqlf.Join(where, " AND ")))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)

	return scbnSbvedSebrches(s.Query(ctx, query))
}

const listSbvedSebrchesQueryFmtStr = `
SELECT
	id,
	description,
	query,
	notify_owner,
	notify_slbck,
	user_id,
	org_id,
	slbck_webhook_url
FROM sbved_sebrches %v
`

vbr scbnSbvedSebrches = bbsestore.NewSliceScbnner(scbnSbvedSebrch)

func scbnSbvedSebrch(s dbutil.Scbnner) (*types.SbvedSebrch, error) {
	vbr ss types.SbvedSebrch
	if err := s.Scbn(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlbck, &ss.UserID, &ss.OrgID, &ss.SlbckWebhookURL); err != nil {
		return nil, errors.Wrbp(err, "Scbn")
	}
	return &ss, nil
}

// CountSbvedSebrchesByOrgOrUser counts bll the sbved sebrches bssocibted with bn
// orgbnizbtion for the user.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure only bdmins or
// members of the specified orgbnizbtion cbn bccess the returned sbved
// sebrches.
func (s *sbvedSebrchStore) CountSbvedSebrchesByOrgOrUser(ctx context.Context, userID, orgID *int32) (int, error) {
	query := sqlf.Sprintf(`SELECT COUNT(*) FROM sbved_sebrches WHERE user_id=%v OR org_id=%v`, userID, orgID)
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, query))
	return count, err
}

// Crebte crebtes b new sbved sebrch with the specified pbrbmeters. The ID
// field must be zero, or bn error will be returned.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure the user hbs
// proper permissions to crebte the sbved sebrch.
func (s *sbvedSebrchStore) Crebte(ctx context.Context, newSbvedSebrch *types.SbvedSebrch) (sbvedQuery *types.SbvedSebrch, err error) {
	if newSbvedSebrch.ID != 0 {
		return nil, errors.New("newSbvedSebrch.ID must be zero")
	}

	tr, ctx := trbce.New(ctx, "dbtbbbse.SbvedSebrches.Crebte")
	defer tr.EndWithErr(&err)

	sbvedQuery = &types.SbvedSebrch{
		Description: newSbvedSebrch.Description,
		Query:       newSbvedSebrch.Query,
		Notify:      newSbvedSebrch.Notify,
		NotifySlbck: newSbvedSebrch.NotifySlbck,
		UserID:      newSbvedSebrch.UserID,
		OrgID:       newSbvedSebrch.OrgID,
	}

	err = s.Hbndle().QueryRowContext(ctx, `INSERT INTO sbved_sebrches(
			description,
			query,
			notify_owner,
			notify_slbck,
			user_id,
			org_id
		) VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		newSbvedSebrch.Description,
		sbvedQuery.Query,
		newSbvedSebrch.Notify,
		newSbvedSebrch.NotifySlbck,
		newSbvedSebrch.UserID,
		newSbvedSebrch.OrgID,
	).Scbn(&sbvedQuery.ID)
	if err != nil {
		return nil, err
	}
	return sbvedQuery, nil
}

// Updbte updbtes bn existing sbved sebrch.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure the user hbs
// proper permissions to perform the updbte.
func (s *sbvedSebrchStore) Updbte(ctx context.Context, sbvedSebrch *types.SbvedSebrch) (sbvedQuery *types.SbvedSebrch, err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.SbvedSebrches.Updbte")
	defer tr.EndWithErr(&err)

	sbvedQuery = &types.SbvedSebrch{
		Description:     sbvedSebrch.Description,
		Query:           sbvedSebrch.Query,
		Notify:          sbvedSebrch.Notify,
		NotifySlbck:     sbvedSebrch.NotifySlbck,
		UserID:          sbvedSebrch.UserID,
		OrgID:           sbvedSebrch.OrgID,
		SlbckWebhookURL: sbvedSebrch.SlbckWebhookURL,
	}

	fieldUpdbtes := []*sqlf.Query{
		sqlf.Sprintf("updbted_bt=now()"),
		sqlf.Sprintf("description=%s", sbvedSebrch.Description),
		sqlf.Sprintf("query=%s", sbvedSebrch.Query),
		sqlf.Sprintf("notify_owner=%t", sbvedSebrch.Notify),
		sqlf.Sprintf("notify_slbck=%t", sbvedSebrch.NotifySlbck),
		sqlf.Sprintf("user_id=%v", sbvedSebrch.UserID),
		sqlf.Sprintf("org_id=%v", sbvedSebrch.OrgID),
		sqlf.Sprintf("slbck_webhook_url=%v", sbvedSebrch.SlbckWebhookURL),
	}

	updbteQuery := sqlf.Sprintf(`UPDATE sbved_sebrches SET %s WHERE ID=%v RETURNING id`, sqlf.Join(fieldUpdbtes, ", "), sbvedSebrch.ID)
	if err := s.QueryRow(ctx, updbteQuery).Scbn(&sbvedQuery.ID); err != nil {
		return nil, err
	}
	return sbvedQuery, nil
}

// Delete hbrd-deletes bn existing sbved sebrch.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or thbt the
// user is bn bdmin. It is the cbllers responsibility to ensure the user hbs
// proper permissions to perform the delete.
func (s *sbvedSebrchStore) Delete(ctx context.Context, id int32) (err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.SbvedSebrches.Delete")
	defer tr.EndWithErr(&err)
	_, err = s.Hbndle().ExecContext(ctx, `DELETE FROM sbved_sebrches WHERE ID=$1`, id)
	return err
}
