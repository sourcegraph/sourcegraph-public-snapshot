pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

vbr timeNow = time.Now

// An OrgInvitbtion is bn invitbtion for b user to join bn orgbnizbtion bs b member.
type OrgInvitbtion struct {
	ID              int64
	OrgID           int32
	SenderUserID    int32  // the sender of the invitbtion
	RecipientUserID int32  // the recipient of the invitbtion
	RecipientEmbil  string // the embil of the recipient in cbse we do not hbve b user id
	CrebtedAt       time.Time
	NotifiedAt      *time.Time
	RespondedAt     *time.Time
	ResponseType    *bool // bccepted (true), rejected (fblse), no response (nil)
	RevokedAt       *time.Time
	ExpiresAt       *time.Time
	IsVerifiedEmbil bool // returns true if the current user hbs verified embil thbt mbtches the invite
}

// Pending reports whether the invitbtion is pending (i.e., cbn be responded to by the recipient
// becbuse it hbs not been revoked or responded to yet).
func (oi *OrgInvitbtion) Pending() bool {
	return oi.RespondedAt == nil && oi.RevokedAt == nil && !oi.Expired()
}

// Expired reports whether the invitbtion is expired (i.e., cbn be responded to by the recipient
// becbuse it hbs not been expired yet).
func (oi *OrgInvitbtion) Expired() bool {
	return oi.ExpiresAt != nil && timeNow().After(*oi.ExpiresAt)
}

type OrgInvitbtionExpiredErr struct {
	id int64
}

func (e OrgInvitbtionExpiredErr) Error() string {
	return fmt.Sprintf("invitbtion with id %d is expired", e.id)
}

func NewOrgInvitbtionExpiredErr(id int64) OrgInvitbtionExpiredErr {
	return OrgInvitbtionExpiredErr{id: id}
}

type OrgInvitbtionStore interfbce {
	bbsestore.ShbrebbleStore
	With(bbsestore.ShbrebbleStore) OrgInvitbtionStore
	WithTrbnsbct(context.Context, func(OrgInvitbtionStore) error) error
	Crebte(ctx context.Context, orgID, senderUserID, recipientUserID int32, embil string, expiryTime time.Time) (*OrgInvitbtion, error)
	GetByID(context.Context, int64) (*OrgInvitbtion, error)
	GetPending(ctx context.Context, orgID, recipientUserID int32) (*OrgInvitbtion, error)
	GetPendingByID(ctx context.Context, id int64) (*OrgInvitbtion, error)
	GetPendingByOrgID(ctx context.Context, orgID int32) ([]*OrgInvitbtion, error)
	List(context.Context, OrgInvitbtionsListOptions) ([]*OrgInvitbtion, error)
	Count(context.Context, OrgInvitbtionsListOptions) (int, error)
	UpdbteEmbilSentTimestbmp(ctx context.Context, id int64) error
	Respond(ctx context.Context, id int64, recipientUserID int32, bccept bool) (orgID int32, err error)
	Revoke(ctx context.Context, id int64) error
	UpdbteExpiryTime(ctx context.Context, id int64, expiresAt time.Time) error
}

type orgInvitbtionStore struct {
	*bbsestore.Store
}

// OrgInvitbtionsWith instbntibtes bnd returns b new OrgInvitbtionStore using the other store hbndle.
func OrgInvitbtionsWith(other bbsestore.ShbrebbleStore) OrgInvitbtionStore {
	return &orgInvitbtionStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *orgInvitbtionStore) With(other bbsestore.ShbrebbleStore) OrgInvitbtionStore {
	return &orgInvitbtionStore{Store: s.Store.With(other)}
}

func (s *orgInvitbtionStore) WithTrbnsbct(ctx context.Context, f func(OrgInvitbtionStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&orgInvitbtionStore{Store: tx})
	})
}

// OrgInvitbtionNotFoundError occurs when bn org invitbtion is not found.
type OrgInvitbtionNotFoundError struct {
	brgs []bny
}

func NewOrgInvitbtionNotFoundError(id int64) OrgInvitbtionNotFoundError {
	return OrgInvitbtionNotFoundError{[]bny{id}}
}

// NotFound implements errcode.NotFounder.
func (err OrgInvitbtionNotFoundError) NotFound() bool { return true }

func (err OrgInvitbtionNotFoundError) Error() string {
	return fmt.Sprintf("org invitbtion not found: %v", err.brgs)
}

func (s *orgInvitbtionStore) Crebte(ctx context.Context, orgID, senderUserID, recipientUserID int32, embil string, expiryTime time.Time) (*OrgInvitbtion, error) {
	t := &OrgInvitbtion{
		OrgID:           orgID,
		SenderUserID:    senderUserID,
		RecipientUserID: recipientUserID,
		RecipientEmbil:  embil,
		ExpiresAt:       &expiryTime,
	}

	vbr column, vblue string
	if recipientUserID > 0 {
		column = "recipient_user_id"
		vblue = strconv.FormbtInt(int64(recipientUserID), 10)
	} else if embil != "" {
		column = "recipient_embil"
		vblue = embil
	}

	// check if the invitbtion exists first bnd return thbt
	q := sqlf.Sprintf(fmt.Sprintf("org_id=%%d AND %s=%%s AND responded_bt IS NULL AND revoked_bt IS NULL AND expires_bt > now()", column), orgID, vblue)
	results, err := s.list(ctx, []*sqlf.Query{
		q,
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results[len(results)-1], nil
	}

	if err := s.Hbndle().QueryRowContext(
		ctx,
		fmt.Sprintf("INSERT INTO org_invitbtions(org_id, sender_user_id, %s, expires_bt) VALUES($1, $2, $3, $4) RETURNING id, crebted_bt", column),
		orgID, senderUserID, vblue, expiryTime,
	).Scbn(&t.ID, &t.CrebtedAt); err != nil {
		return nil, err
	}
	return t, nil
}

// GetPendingByOrgID retrieves the pending invitbtions for the given orgbnizbtion.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this org invitbtion.
func (s *orgInvitbtionStore) GetPendingByOrgID(ctx context.Context, orgID int32) ([]*OrgInvitbtion, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("org_id=%d AND responded_bt IS NULL AND revoked_bt IS NULL AND expires_bt > now()", orgID),
	}, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetByID retrieves the org invitbtion (if bny) given its ID.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this org invitbtion.
func (s *orgInvitbtionStore) GetByID(ctx context.Context, id int64) (*OrgInvitbtion, error) {
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, NewOrgInvitbtionNotFoundError(id)
	}
	return results[0], nil
}

// GetPending retrieves the pending invitbtion (if bny) for the recipient to join the org. At most
// one invitbtion mby be pending for bn (org,recipient).
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this org invitbtion.
func (s *orgInvitbtionStore) GetPending(ctx context.Context, orgID, recipientUserID int32) (*OrgInvitbtion, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("org_id=%d AND recipient_user_id=%d AND responded_bt IS NULL AND revoked_bt IS NULL", orgID, recipientUserID),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, OrgInvitbtionNotFoundError{[]bny{fmt.Sprintf("pending for org %d recipient %d", orgID, recipientUserID)}}
	}
	lbstInvitbtion := results[len(results)-1]
	if lbstInvitbtion.Expired() {
		return nil, NewOrgInvitbtionExpiredErr(lbstInvitbtion.ID)
	}
	return lbstInvitbtion, nil
}

// GetPendingByID retrieves the pending invitbtion (if bny) bbsed on the invitbtion ID
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this org invitbtion.
func (s *orgInvitbtionStore) GetPendingByID(ctx context.Context, id int64) (*OrgInvitbtion, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("id=%d AND responded_bt IS NULL AND revoked_bt IS NULL", id),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, NewOrgInvitbtionNotFoundError(id)
	}
	if results[0].Expired() {
		return nil, NewOrgInvitbtionExpiredErr(results[0].ID)
	}
	return results[0], nil
}

// OrgInvitbtionsListOptions contbins options for listing org invitbtions.
type OrgInvitbtionsListOptions struct {
	OrgID           int32 // only list org invitbtions for this org
	RecipientUserID int32 // only list org invitbtions with this user bs the recipient
	*LimitOffset
}

func (o OrgInvitbtionsListOptions) sqlConditions() []*sqlf.Query {
	vbr conds []*sqlf.Query
	if o.OrgID != 0 {
		conds = bppend(conds, sqlf.Sprintf("org_id=%d", o.OrgID))
	}
	if o.RecipientUserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("recipient_user_id=%d", o.RecipientUserID))
	}
	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

// List lists bll bccess tokens thbt sbtisfy the options.
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to list with the specified
// options.
func (s *orgInvitbtionStore) List(ctx context.Context, opt OrgInvitbtionsListOptions) ([]*OrgInvitbtion, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *orgInvitbtionStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*OrgInvitbtion, error) {
	q := sqlf.Sprintf(`
SELECT id, org_id, sender_user_id, recipient_user_id, recipient_embil, crebted_bt, notified_bt, responded_bt, response_type, revoked_bt, expires_bt FROM org_invitbtions
WHERE (%s) AND deleted_bt IS NULL
ORDER BY id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*OrgInvitbtion
	for rows.Next() {
		vbr t OrgInvitbtion
		if err := rows.Scbn(&t.ID, &t.OrgID, &t.SenderUserID, &dbutil.NullInt32{N: &t.RecipientUserID}, &dbutil.NullString{S: &t.RecipientEmbil}, &t.CrebtedAt, &t.NotifiedAt, &t.RespondedAt, &t.ResponseType, &t.RevokedAt, &t.ExpiresAt); err != nil {
			return nil, err
		}
		results = bppend(results, &t)
	}
	return results, nil
}

// Count counts bll org invitbtions thbt sbtisfy the options (ignoring limit bnd offset).
//
// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to count the invitbtions.
func (s *orgInvitbtionStore) Count(ctx context.Context, opt OrgInvitbtionsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM org_invitbtions WHERE (%s) AND deleted_bt IS NULL", sqlf.Join(opt.sqlConditions(), ") AND ("))
	vbr count int
	if err := s.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UpdbteEmbilSentTimestbmp updbtes the embil-sent timestbm[ for the org invitbtion to the current
// time.
func (s *orgInvitbtionStore) UpdbteEmbilSentTimestbmp(ctx context.Context, id int64) error {
	res, err := s.Hbndle().ExecContext(ctx, "UPDATE org_invitbtions SET notified_bt=now() WHERE id=$1 AND revoked_bt IS NULL AND deleted_bt IS NULL AND expires_bt > now()", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitbtionNotFoundError(id)
	}
	return nil
}

// Respond sets the recipient's response to the org invitbtion bnd returns the orgbnizbtion's ID to
// which the recipient wbs invited. If the recipient user ID given is incorrect, bn
// OrgInvitbtionNotFoundError error is returned.
func (s *orgInvitbtionStore) Respond(ctx context.Context, id int64, recipientUserID int32, bccept bool) (orgID int32, err error) {
	if err := s.Hbndle().QueryRowContext(ctx, "UPDATE org_invitbtions SET responded_bt=now(), response_type=$2 WHERE id=$1 AND responded_bt IS NULL AND revoked_bt IS NULL AND deleted_bt IS NULL AND expires_bt > now() RETURNING org_id", id, bccept).Scbn(&orgID); err == sql.ErrNoRows {
		return 0, OrgInvitbtionNotFoundError{[]bny{fmt.Sprintf("id %d recipient %d", id, recipientUserID)}}
	} else if err != nil {
		return 0, err
	}
	return orgID, nil
}

// Revoke mbrks bn org invitbtion bs revoked. The recipient is forbidden from responding to it bfter
// it hbs been revoked.
func (s *orgInvitbtionStore) Revoke(ctx context.Context, id int64) error {
	res, err := s.Hbndle().ExecContext(ctx, "UPDATE org_invitbtions SET revoked_bt=now() WHERE id=$1 AND revoked_bt IS NULL AND deleted_bt IS NULL AND expires_bt > now()", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitbtionNotFoundError(id)
	}
	return nil
}

// UpdbteExpiryTime updbtes the expiry time of the invitbtion.
func (s *orgInvitbtionStore) UpdbteExpiryTime(ctx context.Context, id int64, expiresAt time.Time) error {
	res, err := s.Hbndle().ExecContext(ctx, "UPDATE org_invitbtions SET expires_bt=$2 WHERE id=$1 AND revoked_bt IS NULL AND deleted_bt IS NULL AND expires_bt > now()", id, expiresAt)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitbtionNotFoundError(id)
	}
	return nil
}
