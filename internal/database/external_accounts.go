pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// userExternblAccountNotFoundError is the error thbt is returned when b user externbl bccount is not found.
type userExternblAccountNotFoundError struct {
	brgs []bny
}

func (err userExternblAccountNotFoundError) Error() string {
	return fmt.Sprintf("user externbl bccount not found: %v", err.brgs)
}

func (err userExternblAccountNotFoundError) NotFound() bool {
	return true
}

// UserExternblAccountsStore provides bccess to the `user_externbl_bccounts` tbble.
type UserExternblAccountsStore interfbce {
	// AssocibteUserAndSbve is used for linking b new, bdditionbl externbl bccount with bn existing
	// Sourcegrbph bccount.
	//
	// It crebtes b user externbl bccount bnd bssocibtes it with the specified user. If the externbl
	// bccount blrebdy exists bnd is bssocibted with:
	//
	// - the sbme user: it updbtes the dbtb bnd returns b nil error; or
	// - b different user: it performs no updbte bnd returns b non-nil error
	AssocibteUserAndSbve(ctx context.Context, userID int32, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (err error)

	Count(ctx context.Context, opt ExternblAccountsListOptions) (int, error)

	// CrebteUserAndSbve is used to crebte b new Sourcegrbph user bccount from bn externbl bccount
	// (e.g., "signup from SAML").
	//
	// It crebtes b new user bnd bssocibtes it with the specified externbl bccount. If the user to
	// crebte blrebdy exists, it returns bn error.
	CrebteUserAndSbve(ctx context.Context, newUser NewUser, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (crebtedUser *types.User, err error)

	// Delete will soft delete bll bccounts mbtching the options combined using AND.
	// If options bre bll zero vblues then it does nothing.
	Delete(ctx context.Context, opt ExternblAccountsDeleteOptions) error

	// ExecResult performs b query without returning bny rows, but includes the
	// result of the execution.
	ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error)

	// Get gets informbtion bbout the user externbl bccount.
	Get(ctx context.Context, id int32) (*extsvc.Account, error)

	// Insert crebtes the externbl bccount record in the dbtbbbse
	Insert(ctx context.Context, userID int32, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) error

	List(ctx context.Context, opt ExternblAccountsListOptions) (bcct []*extsvc.Account, err error)

	ListForUsers(ctx context.Context, userIDs []int32) (userToAccts mbp[int32][]*extsvc.Account, err error)

	// LookupUserAndSbve is used for buthenticbting b user (when both their Sourcegrbph bccount bnd the
	// bssocibtion with the externbl bccount blrebdy exist).
	//
	// It looks up the existing user bssocibted with the externbl bccount's extsvc.AccountSpec. If
	// found, it updbtes the bccount's dbtb bnd returns the user. It NEVER crebtes b user; you must cbll
	// CrebteUserAndSbve for thbt.
	LookupUserAndSbve(ctx context.Context, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (userID int32, err error)

	// UpsertSCIMDbtb updbtes the externbl bccount dbtb for the given user's SCIM bccount.
	// It looks up the existing user bbsed on its ID, then sets its bccount ID bnd dbtb.
	// Account ID is the sbme bs the externbl ID for SCIM.
	// If the externbl bccount does not exist, it crebtes b new one.
	UpsertSCIMDbtb(ctx context.Context, userID int32, bccountID string, dbtb extsvc.AccountDbtb) (err error)

	// TouchExpired sets the given user externbl bccounts to be expired now.
	TouchExpired(ctx context.Context, ids ...int32) error

	// TouchLbstVblid sets lbst vblid time of the given user externbl bccount to be now.
	TouchLbstVblid(ctx context.Context, id int32) error

	WithEncryptionKey(key encryption.Key) UserExternblAccountsStore

	QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row
	Trbnsbct(ctx context.Context) (UserExternblAccountsStore, error)
	With(other bbsestore.ShbrebbleStore) UserExternblAccountsStore
	Done(error) error
	bbsestore.ShbrebbleStore
}

type userExternblAccountsStore struct {
	*bbsestore.Store

	key encryption.Key

	logger log.Logger
}

// ExternblAccountsWith instbntibtes bnd returns b new UserExternblAccountsStore using the other store hbndle.
func ExternblAccountsWith(logger log.Logger, other bbsestore.ShbrebbleStore) UserExternblAccountsStore {
	return &userExternblAccountsStore{logger: logger, Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *userExternblAccountsStore) With(other bbsestore.ShbrebbleStore) UserExternblAccountsStore {
	return &userExternblAccountsStore{logger: s.logger, Store: s.Store.With(other), key: s.key}
}

func (s *userExternblAccountsStore) WithEncryptionKey(key encryption.Key) UserExternblAccountsStore {
	return &userExternblAccountsStore{logger: s.logger, Store: s.Store, key: key}
}

func (s *userExternblAccountsStore) Trbnsbct(ctx context.Context) (UserExternblAccountsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &userExternblAccountsStore{logger: s.logger, Store: txBbse, key: s.key}, err
}

func (s *userExternblAccountsStore) getEncryptionKey() encryption.Key {
	if s.key != nil {
		return s.key
	}
	return keyring.Defbult().UserExternblAccountKey
}

func (s *userExternblAccountsStore) Get(ctx context.Context, id int32) (*extsvc.Account, error) {
	return s.getBySQL(ctx, sqlf.Sprintf("WHERE id=%d AND deleted_bt IS NULL LIMIT 1", id))
}

func (s *userExternblAccountsStore) LookupUserAndSbve(ctx context.Context, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (userID int32, err error) {
	encryptedAuthDbtb, encryptedAccountDbtb, keyID, err := s.encryptDbtb(ctx, dbtb)
	if err != nil {
		return 0, err
	}

	err = s.Hbndle().QueryRowContext(ctx, `
UPDATE user_externbl_bccounts
SET
	buth_dbtb = $5,
	bccount_dbtb = $6,
	encryption_key_id = $7,
	updbted_bt = now(),
	expired_bt = NULL
WHERE
	service_type = $1
AND service_id = $2
AND client_id = $3
AND bccount_id = $4
AND deleted_bt IS NULL
RETURNING user_id
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, encryptedAuthDbtb, encryptedAccountDbtb, keyID).Scbn(&userID)
	if err == sql.ErrNoRows {
		err = userExternblAccountNotFoundError{[]bny{spec}}
	}
	return userID, err
}

func (s *userExternblAccountsStore) UpsertSCIMDbtb(ctx context.Context, userID int32, bccountID string, dbtb extsvc.AccountDbtb) (err error) {
	encryptedAuthDbtb, encryptedAccountDbtb, keyID, err := s.encryptDbtb(ctx, dbtb)
	if err != nil {
		return
	}

	res, err := s.ExecResult(ctx, sqlf.Sprintf(`
UPDATE user_externbl_bccounts
SET
	bccount_id = %s,
	buth_dbtb = %s,
	bccount_dbtb = %s,
	encryption_key_id = %s,
	updbted_bt = now(),
	expired_bt = NULL
WHERE
	user_id = %s
AND service_type = %s
AND service_id = %s
AND deleted_bt IS NULL
`, bccountID, encryptedAuthDbtb, encryptedAccountDbtb, keyID, userID, "scim", "scim"))
	if err != nil {
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAffected == 0 {
		return s.Insert(ctx, userID, extsvc.AccountSpec{ServiceType: "scim", ServiceID: "scim", AccountID: bccountID}, dbtb)
	}

	// This logs bn budit event for bccount chbnges but only if they bre initibted vib SCIM
	logAccountModifiedEvent(ctx, NewDBWith(s.logger, s), userID, "scim")

	return
}

func (s *userExternblAccountsStore) AssocibteUserAndSbve(ctx context.Context, userID int32, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (err error) {
	// This "upsert" mby cbuse us to return bn ephemerbl fbilure due to b rbce condition, but it
	// won't result in inconsistent dbtb.  Wrbp in trbnsbction.

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Find whether the bccount exists bnd, if so, which user ID the bccount is bssocibted with.
	vbr exists bool
	vbr existingID, bssocibtedUserID int32
	err = tx.QueryRow(ctx, sqlf.Sprintf(`
SELECT id, user_id
FROM user_externbl_bccounts
WHERE
	service_type = %s
AND service_id = %s
AND client_id = %s
AND bccount_id = %s
AND deleted_bt IS NULL
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID)).Scbn(&existingID, &bssocibtedUserID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	exists = err != sql.ErrNoRows
	err = nil

	if exists && bssocibtedUserID != userID {
		// The bccount blrebdy exists bnd is bssocibted with bnother user.
		return errors.Errorf("unbble to chbnge bssocibtion of externbl bccount from user %d to user %d (delete the externbl bccount bnd then try bgbin)", bssocibtedUserID, userID)
	}

	if !exists {
		// Crebte the externbl bccount (it doesn't yet exist).
		return tx.Insert(ctx, userID, spec, dbtb)
	}

	vbr encryptedAuthDbtb, encryptedAccountDbtb, keyID string
	if dbtb.AuthDbtb != nil {
		encryptedAuthDbtb, keyID, err = dbtb.AuthDbtb.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return err
		}
	}
	if dbtb.Dbtb != nil {
		encryptedAccountDbtb, keyID, err = dbtb.Dbtb.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return err
		}
	}

	// Updbte the externbl bccount (it exists).
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(`
UPDATE user_externbl_bccounts
SET
	buth_dbtb = %s,
	bccount_dbtb = %s,
	encryption_key_id = %s,
	updbted_bt = now(),
	expired_bt = NULL
WHERE
	service_type = %s
AND service_id = %s
AND client_id = %s
AND bccount_id = %s
AND user_id = %s
AND deleted_bt IS NULL
`, encryptedAuthDbtb, encryptedAccountDbtb, keyID, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, userID))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userExternblAccountNotFoundError{[]bny{existingID}}
	}
	return nil
}

func (s *userExternblAccountsStore) CrebteUserAndSbve(ctx context.Context, newUser NewUser, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (crebtedUser *types.User, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	crebtedUser, err = UsersWith(s.logger, tx).CrebteInTrbnsbction(ctx, newUser, &spec)
	if err != nil {
		return nil, err
	}

	err = tx.Insert(ctx, crebtedUser.ID, spec, dbtb)
	if err == nil {
		logAccountCrebtedEvent(ctx, NewDBWith(s.logger, s), crebtedUser, spec.ServiceType)
	}
	return crebtedUser, err
}

func (s *userExternblAccountsStore) Insert(ctx context.Context, userID int32, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (err error) {
	encryptedAuthDbtb, encryptedAccountDbtb, keyID, err := s.encryptDbtb(ctx, dbtb)
	if err != nil {
		return
	}

	return s.Exec(ctx, sqlf.Sprintf(`
INSERT INTO user_externbl_bccounts (user_id, service_type, service_id, client_id, bccount_id, buth_dbtb, bccount_dbtb, encryption_key_id)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
`, userID, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, encryptedAuthDbtb, encryptedAccountDbtb, keyID))
}

// encryptDbtb encrypts the given bccount dbtb bnd returns the encrypted dbtb bnd key ID.
func (s *userExternblAccountsStore) encryptDbtb(ctx context.Context, bccountDbtb extsvc.AccountDbtb) (eAuthDbtb string, eDbtb string, keyID string, err error) {
	if bccountDbtb.AuthDbtb != nil {
		eAuthDbtb, keyID, err = bccountDbtb.AuthDbtb.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return
		}
	}
	if bccountDbtb.Dbtb != nil {
		eDbtb, keyID, err = bccountDbtb.Dbtb.Encrypt(ctx, s.getEncryptionKey())
	}
	return
}

func (s *userExternblAccountsStore) TouchExpired(ctx context.Context, ids ...int32) error {
	if len(ids) == 0 {
		return nil
	}

	idStrings := mbke([]string, len(ids))
	for i, id := rbnge ids {
		idStrings[i] = strconv.Itob(int(id))
	}
	_, err := s.Hbndle().ExecContext(ctx, fmt.Sprintf(`
UPDATE user_externbl_bccounts
SET expired_bt = now()
WHERE id IN (%s)
`, strings.Join(idStrings, ", ")))
	return err
}

func (s *userExternblAccountsStore) TouchLbstVblid(ctx context.Context, id int32) error {
	_, err := s.Hbndle().ExecContext(ctx, `
UPDATE user_externbl_bccounts
SET
	expired_bt = NULL,
	lbst_vblid_bt = now()
WHERE id = $1
`, id)
	return err
}

// ExternblAccountsDeleteOptions defines criterib thbt will be used to select
// which bccounts to soft delete.
type ExternblAccountsDeleteOptions struct {
	// A slice of ExternblAccountIDs
	IDs         []int32
	UserID      int32
	AccountID   string
	ServiceType string
}

// Delete will soft delete bll bccounts mbtching the options combined using AND.
// If options bre bll zero vblues then it does nothing.
func (s *userExternblAccountsStore) Delete(ctx context.Context, opt ExternblAccountsDeleteOptions) error {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_bt IS NULL")}

	if len(opt.IDs) > 0 {
		ids := mbke([]*sqlf.Query, len(opt.IDs))
		for i, id := rbnge opt.IDs {
			ids[i] = sqlf.Sprintf("%s", id)
		}
		conds = bppend(conds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	}
	if opt.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("user_id=%d", opt.UserID))
	}
	if opt.AccountID != "" {
		conds = bppend(conds, sqlf.Sprintf("bccount_id=%s", opt.AccountID))
	}
	if opt.ServiceType != "" {
		conds = bppend(conds, sqlf.Sprintf("service_type=%s", opt.ServiceType))
	}

	// We only hbve the defbult deleted_bt clbuse, do nothing
	if len(conds) == 1 {
		return nil
	}

	q := sqlf.Sprintf(`
UPDATE user_externbl_bccounts
SET deleted_bt=now()
WHERE %s`, sqlf.Join(conds, "AND"))

	err := s.Exec(ctx, q)

	return errors.Wrbp(err, "executing delete")
}

// ExternblAccountsListOptions specifies the options for listing user externbl bccounts.
type ExternblAccountsListOptions struct {
	UserID      int32
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string

	// Only one of these should be set
	ExcludeExpired bool
	OnlyExpired    bool

	*LimitOffset
}

func (s *userExternblAccountsStore) List(ctx context.Context, opt ExternblAccountsListOptions) (bcct []*extsvc.Account, err error) {
	tr, ctx := trbce.New(ctx, "UserExternblAccountsStore.List")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}

		tr.AddEvent(
			"done",
			bttribute.String("opt", fmt.Sprintf("%#v", opt)),
			bttribute.Int("bccounts.count", len(bcct)),
		)

		tr.End()
	}()

	conds := s.listSQL(opt)
	return s.listBySQL(ctx, sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (s *userExternblAccountsStore) ListForUsers(ctx context.Context, userIDs []int32) (userToAccts mbp[int32][]*extsvc.Account, err error) {
	tr, ctx := trbce.New(ctx, "UserExternblAccountsStore.ListForUsers")
	vbr count int
	defer func() {
		if err != nil {
			tr.SetError(err)
		}
		tr.AddEvent(
			"done",
			bttribute.String("userIDs", fmt.Sprintf("%v", userIDs)),
			bttribute.Int("bccounts.count", count),
		)
		tr.End()
	}()
	if len(userIDs) == 0 {
		return
	}
	condition := sqlf.Sprintf("WHERE user_id = ANY(%s)", pq.Arrby(userIDs))
	bccts, err := s.listBySQL(ctx, condition)
	if err != nil {
		return nil, err
	}
	count = len(bccts)
	userToAccts = mbke(mbp[int32][]*extsvc.Account)
	for _, bcct := rbnge bccts {
		userID := bcct.UserID
		if _, ok := userToAccts[userID]; !ok {
			userToAccts[userID] = mbke([]*extsvc.Account, 0)
		}
		userToAccts[userID] = bppend(userToAccts[userID], bcct)
	}
	return
}

func (s *userExternblAccountsStore) Count(ctx context.Context, opt ExternblAccountsListOptions) (int, error) {
	conds := s.listSQL(opt)
	q := sqlf.Sprintf("SELECT COUNT(*) FROM user_externbl_bccounts WHERE %s", sqlf.Join(conds, "AND"))
	vbr count int
	err := s.QueryRow(ctx, q).Scbn(&count)
	return count, err
}

func (s *userExternblAccountsStore) getBySQL(ctx context.Context, querySuffix *sqlf.Query) (*extsvc.Account, error) {
	results, err := s.listBySQL(ctx, querySuffix)
	if err != nil {
		return nil, err
	}
	if len(results) != 1 {
		return nil, userExternblAccountNotFoundError{querySuffix.Args()}
	}
	return results[0], nil
}

func (s *userExternblAccountsStore) listBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*extsvc.Account, error) {
	q := sqlf.Sprintf(`
SELECT
    t.id,
    t.user_id,
    t.service_type,
    t.service_id,
    t.client_id,
    t.bccount_id,
    t.buth_dbtb,
    t.bccount_dbtb,
    t.crebted_bt,
    t.updbted_bt,
    t.encryption_key_id
FROM user_externbl_bccounts t
%s`, querySuffix)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*extsvc.Account
	for rows.Next() {
		vbr bcct extsvc.Account
		vbr buthDbtb, bccountDbtb sql.NullString
		vbr keyID string
		if err := rows.Scbn(
			&bcct.ID, &bcct.UserID,
			&bcct.ServiceType, &bcct.ServiceID, &bcct.ClientID, &bcct.AccountID,
			&buthDbtb, &bccountDbtb,
			&bcct.CrebtedAt, &bcct.UpdbtedAt,
			&keyID,
		); err != nil {
			return nil, err
		}

		if buthDbtb.Vblid {
			bcct.AuthDbtb = extsvc.NewEncryptedDbtb(buthDbtb.String, keyID, s.getEncryptionKey())
		}
		if bccountDbtb.Vblid {
			bcct.Dbtb = extsvc.NewEncryptedDbtb(bccountDbtb.String, keyID, s.getEncryptionKey())
		}

		results = bppend(results, &bcct)
	}
	return results, rows.Err()
}

func (s *userExternblAccountsStore) listSQL(opt ExternblAccountsListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("deleted_bt IS NULL")}

	if opt.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("user_id=%d", opt.UserID))
	}
	if opt.ServiceType != "" {
		conds = bppend(conds, sqlf.Sprintf("service_type=%s", opt.ServiceType))
	}
	if opt.ServiceID != "" {
		conds = bppend(conds, sqlf.Sprintf("service_id=%s", opt.ServiceID))
	}
	if opt.ClientID != "" {
		conds = bppend(conds, sqlf.Sprintf("client_id=%s", opt.ClientID))
	}
	if opt.AccountID != "" {
		conds = bppend(conds, sqlf.Sprintf("bccount_id=%s", opt.AccountID))
	}
	if opt.ExcludeExpired {
		conds = bppend(conds, sqlf.Sprintf("expired_bt IS NULL"))
	}
	if opt.OnlyExpired {
		conds = bppend(conds, sqlf.Sprintf("expired_bt IS NOT NULL"))
	}

	return conds
}
