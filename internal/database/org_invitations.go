package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var timeNow = time.Now

// An OrgInvitation is an invitation for a user to join an organization as a member.
type OrgInvitation struct {
	ID              int64
	OrgID           int32
	SenderUserID    int32  // the sender of the invitation
	RecipientUserID int32  // the recipient of the invitation
	RecipientEmail  string // the email of the recipient in case we do not have a user id
	CreatedAt       time.Time
	NotifiedAt      *time.Time
	RespondedAt     *time.Time
	ResponseType    *bool // accepted (true), rejected (false), no response (nil)
	RevokedAt       *time.Time
	ExpiresAt       *time.Time
	IsVerifiedEmail bool // returns true if the current user has verified email that matches the invite
}

// Pending reports whether the invitation is pending (i.e., can be responded to by the recipient
// because it has not been revoked or responded to yet).
func (oi *OrgInvitation) Pending() bool {
	return oi.RespondedAt == nil && oi.RevokedAt == nil && !oi.Expired()
}

// Expired reports whether the invitation is expired (i.e., can be responded to by the recipient
// because it has not been expired yet).
func (oi *OrgInvitation) Expired() bool {
	return oi.ExpiresAt != nil && timeNow().After(*oi.ExpiresAt)
}

type OrgInvitationExpiredErr struct {
	id int64
}

func (e OrgInvitationExpiredErr) Error() string {
	return fmt.Sprintf("invitation with id %d is expired", e.id)
}

func NewOrgInvitationExpiredErr(id int64) OrgInvitationExpiredErr {
	return OrgInvitationExpiredErr{id: id}
}

type OrgInvitationStore interface {
	basestore.ShareableStore
	With(basestore.ShareableStore) OrgInvitationStore
	Transact(context.Context) (OrgInvitationStore, error)
	Create(ctx context.Context, orgID, senderUserID, recipientUserID int32, email string, expiryTime time.Time) (*OrgInvitation, error)
	GetByID(context.Context, int64) (*OrgInvitation, error)
	GetPending(ctx context.Context, orgID, recipientUserID int32) (*OrgInvitation, error)
	GetPendingByID(ctx context.Context, id int64) (*OrgInvitation, error)
	GetPendingByOrgID(ctx context.Context, orgID int32) ([]*OrgInvitation, error)
	List(context.Context, OrgInvitationsListOptions) ([]*OrgInvitation, error)
	Count(context.Context, OrgInvitationsListOptions) (int, error)
	UpdateEmailSentTimestamp(ctx context.Context, id int64) error
	Respond(ctx context.Context, id int64, recipientUserID int32, accept bool) (orgID int32, err error)
	Revoke(ctx context.Context, id int64) error
	UpdateExpiryTime(ctx context.Context, id int64, expiresAt time.Time) error
}

type orgInvitationStore struct {
	*basestore.Store
}

// OrgInvitationsWith instantiates and returns a new OrgInvitationStore using the other store handle.
func OrgInvitationsWith(other basestore.ShareableStore) OrgInvitationStore {
	return &orgInvitationStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *orgInvitationStore) With(other basestore.ShareableStore) OrgInvitationStore {
	return &orgInvitationStore{Store: s.Store.With(other)}
}

func (s *orgInvitationStore) Transact(ctx context.Context) (OrgInvitationStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &orgInvitationStore{Store: txBase}, err
}

// OrgInvitationNotFoundError occurs when an org invitation is not found.
type OrgInvitationNotFoundError struct {
	args []any
}

func NewOrgInvitationNotFoundError(id int64) OrgInvitationNotFoundError {
	return OrgInvitationNotFoundError{[]any{id}}
}

// NotFound implements errcode.NotFounder.
func (err OrgInvitationNotFoundError) NotFound() bool { return true }

func (err OrgInvitationNotFoundError) Error() string {
	return fmt.Sprintf("org invitation not found: %v", err.args)
}

func (s *orgInvitationStore) Create(ctx context.Context, orgID, senderUserID, recipientUserID int32, email string, expiryTime time.Time) (*OrgInvitation, error) {
	t := &OrgInvitation{
		OrgID:           orgID,
		SenderUserID:    senderUserID,
		RecipientUserID: recipientUserID,
		RecipientEmail:  email,
		ExpiresAt:       &expiryTime,
	}

	var column, value string
	if recipientUserID > 0 {
		column = "recipient_user_id"
		value = strconv.FormatInt(int64(recipientUserID), 10)
	} else if email != "" {
		column = "recipient_email"
		value = email
	}

	// check if the invitation exists first and return that
	q := sqlf.Sprintf(fmt.Sprintf("org_id=%%d AND %s=%%s AND responded_at IS NULL AND revoked_at IS NULL AND expires_at > now()", column), orgID, value)
	results, err := s.list(ctx, []*sqlf.Query{
		q,
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results[len(results)-1], nil
	}

	if err := s.Handle().QueryRowContext(
		ctx,
		fmt.Sprintf("INSERT INTO org_invitations(org_id, sender_user_id, %s, expires_at) VALUES($1, $2, $3, $4) RETURNING id, created_at", column),
		orgID, senderUserID, value, expiryTime,
	).Scan(&t.ID, &t.CreatedAt); err != nil {
		return nil, err
	}
	return t, nil
}

// GetPendingByOrgID retrieves the pending invitations for the given organization.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *orgInvitationStore) GetPendingByOrgID(ctx context.Context, orgID int32) ([]*OrgInvitation, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("org_id=%d AND responded_at IS NULL AND revoked_at IS NULL AND expires_at > now()", orgID),
	}, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetByID retrieves the org invitation (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *orgInvitationStore) GetByID(ctx context.Context, id int64) (*OrgInvitation, error) {
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, NewOrgInvitationNotFoundError(id)
	}
	return results[0], nil
}

// GetPending retrieves the pending invitation (if any) for the recipient to join the org. At most
// one invitation may be pending for an (org,recipient).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *orgInvitationStore) GetPending(ctx context.Context, orgID, recipientUserID int32) (*OrgInvitation, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("org_id=%d AND recipient_user_id=%d AND responded_at IS NULL AND revoked_at IS NULL", orgID, recipientUserID),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, OrgInvitationNotFoundError{[]any{fmt.Sprintf("pending for org %d recipient %d", orgID, recipientUserID)}}
	}
	lastInvitation := results[len(results)-1]
	if lastInvitation.Expired() {
		return nil, NewOrgInvitationExpiredErr(lastInvitation.ID)
	}
	return lastInvitation, nil
}

// GetPendingByID retrieves the pending invitation (if any) based on the invitation ID
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *orgInvitationStore) GetPendingByID(ctx context.Context, id int64) (*OrgInvitation, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("id=%d AND responded_at IS NULL AND revoked_at IS NULL", id),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, NewOrgInvitationNotFoundError(id)
	}
	if results[0].Expired() {
		return nil, NewOrgInvitationExpiredErr(results[0].ID)
	}
	return results[0], nil
}

// OrgInvitationsListOptions contains options for listing org invitations.
type OrgInvitationsListOptions struct {
	OrgID           int32 // only list org invitations for this org
	RecipientUserID int32 // only list org invitations with this user as the recipient
	*LimitOffset
}

func (o OrgInvitationsListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if o.OrgID != 0 {
		conds = append(conds, sqlf.Sprintf("org_id=%d", o.OrgID))
	}
	if o.RecipientUserID != 0 {
		conds = append(conds, sqlf.Sprintf("recipient_user_id=%d", o.RecipientUserID))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

// List lists all access tokens that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s *orgInvitationStore) List(ctx context.Context, opt OrgInvitationsListOptions) ([]*OrgInvitation, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *orgInvitationStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*OrgInvitation, error) {
	q := sqlf.Sprintf(`
SELECT id, org_id, sender_user_id, recipient_user_id, recipient_email, created_at, notified_at, responded_at, response_type, revoked_at, expires_at FROM org_invitations
WHERE (%s) AND deleted_at IS NULL
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

	var results []*OrgInvitation
	for rows.Next() {
		var t OrgInvitation
		if err := rows.Scan(&t.ID, &t.OrgID, &t.SenderUserID, &dbutil.NullInt32{N: &t.RecipientUserID}, &dbutil.NullString{S: &t.RecipientEmail}, &t.CreatedAt, &t.NotifiedAt, &t.RespondedAt, &t.ResponseType, &t.RevokedAt, &t.ExpiresAt); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all org invitations that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the invitations.
func (s *orgInvitationStore) Count(ctx context.Context, opt OrgInvitationsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM org_invitations WHERE (%s) AND deleted_at IS NULL", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateEmailSentTimestamp updates the email-sent timestam[ for the org invitation to the current
// time.
func (s *orgInvitationStore) UpdateEmailSentTimestamp(ctx context.Context, id int64) error {
	res, err := s.Handle().ExecContext(ctx, "UPDATE org_invitations SET notified_at=now() WHERE id=$1 AND revoked_at IS NULL AND deleted_at IS NULL AND expires_at > now()", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitationNotFoundError(id)
	}
	return nil
}

// Respond sets the recipient's response to the org invitation and returns the organization's ID to
// which the recipient was invited. If the recipient user ID given is incorrect, an
// OrgInvitationNotFoundError error is returned.
func (s *orgInvitationStore) Respond(ctx context.Context, id int64, recipientUserID int32, accept bool) (orgID int32, err error) {
	if err := s.Handle().QueryRowContext(ctx, "UPDATE org_invitations SET responded_at=now(), response_type=$2 WHERE id=$1 AND responded_at IS NULL AND revoked_at IS NULL AND deleted_at IS NULL AND expires_at > now() RETURNING org_id", id, accept).Scan(&orgID); err == sql.ErrNoRows {
		return 0, OrgInvitationNotFoundError{[]any{fmt.Sprintf("id %d recipient %d", id, recipientUserID)}}
	} else if err != nil {
		return 0, err
	}
	return orgID, nil
}

// Revoke marks an org invitation as revoked. The recipient is forbidden from responding to it after
// it has been revoked.
func (s *orgInvitationStore) Revoke(ctx context.Context, id int64) error {
	res, err := s.Handle().ExecContext(ctx, "UPDATE org_invitations SET revoked_at=now() WHERE id=$1 AND revoked_at IS NULL AND deleted_at IS NULL AND expires_at > now()", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitationNotFoundError(id)
	}
	return nil
}

// UpdateExpiryTime updates the expiry time of the invitation.
func (s *orgInvitationStore) UpdateExpiryTime(ctx context.Context, id int64, expiresAt time.Time) error {
	res, err := s.Handle().ExecContext(ctx, "UPDATE org_invitations SET expires_at=$2 WHERE id=$1 AND revoked_at IS NULL AND deleted_at IS NULL AND expires_at > now()", id, expiresAt)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return NewOrgInvitationNotFoundError(id)
	}
	return nil
}
