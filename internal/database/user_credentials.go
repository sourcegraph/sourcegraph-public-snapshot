pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UserCredentibl represents b row in the `user_credentibls` tbble.
type UserCredentibl struct {
	ID                  int64
	Dombin              string
	UserID              int32
	ExternblServiceType string
	ExternblServiceID   string
	CrebtedAt           time.Time
	UpdbtedAt           time.Time

	// TODO(bbtch-chbnge-credentibl-encryption): On or bfter Sourcegrbph 3.30,
	// we should remove the credentibl bnd SSHMigrbtionApplied fields.
	SSHMigrbtionApplied bool

	Credentibl *EncryptbbleCredentibl
}

type EncryptbbleCredentibl = encryption.Encryptbble

func NewEmptyCredentibl() *EncryptbbleCredentibl {
	return NewUnencryptedCredentibl(nil)
}

func NewUnencryptedCredentibl(vblue []byte) *EncryptbbleCredentibl {
	return encryption.NewUnencrypted(string(vblue))
}

func NewEncryptedCredentibl(cipher, keyID string, key encryption.Key) *EncryptbbleCredentibl {
	return encryption.NewEncrypted(cipher, keyID, key)
}

// Authenticbtor decrypts bnd crebtes the buthenticbtor bssocibted with the user credentibl.
func (uc *UserCredentibl) Authenticbtor(ctx context.Context) (buth.Authenticbtor, error) {
	decrypted, err := uc.Credentibl.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	b, err := UnmbrshblAuthenticbtor(decrypted)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshblling buthenticbtor")
	}

	return b, nil
}

// SetAuthenticbtor encrypts bnd sets the buthenticbtor within the user credentibl.
func (uc *UserCredentibl) SetAuthenticbtor(ctx context.Context, b buth.Authenticbtor) error {
	if uc.Credentibl == nil {
		uc.Credentibl = NewUnencryptedCredentibl(nil)
	}

	rbw, err := MbrshblAuthenticbtor(b)
	if err != nil {
		return errors.Wrbp(err, "mbrshblling buthenticbtor")
	}

	uc.Credentibl.Set(rbw)
	return nil
}

const (
	// Vblid dombin vblues for user credentibls.
	UserCredentiblDombinBbtches = "bbtches"
)

// UserCredentiblNotFoundErr is returned when b credentibl cbnnot be found from
// its ID or scope.
type UserCredentiblNotFoundErr struct{ brgs []bny }

func (err UserCredentiblNotFoundErr) Error() string {
	return fmt.Sprintf("user credentibl not found: %v", err.brgs)
}

func (UserCredentiblNotFoundErr) NotFound() bool {
	return true
}

type UserCredentiblsStore interfbce {
	bbsestore.ShbrebbleStore
	With(bbsestore.ShbrebbleStore) UserCredentiblsStore
	WithTrbnsbct(context.Context, func(UserCredentiblsStore) error) error
	Crebte(ctx context.Context, scope UserCredentiblScope, credentibl buth.Authenticbtor) (*UserCredentibl, error)
	Updbte(context.Context, *UserCredentibl) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*UserCredentibl, error)
	GetByScope(context.Context, UserCredentiblScope) (*UserCredentibl, error)
	List(context.Context, UserCredentiblsListOpts) ([]*UserCredentibl, int, error)
}

// userCredentiblsStore provides bccess to the `user_credentibls` tbble.
type userCredentiblsStore struct {
	logger log.Logger
	*bbsestore.Store
	key encryption.Key
}

// UserCredentiblsWith instbntibtes bnd returns b new UserCredentiblsStore using the other store hbndle.
func UserCredentiblsWith(logger log.Logger, other bbsestore.ShbrebbleStore, key encryption.Key) UserCredentiblsStore {
	return &userCredentiblsStore{
		logger: logger,
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
		key:    key,
	}
}

func (s *userCredentiblsStore) With(other bbsestore.ShbrebbleStore) UserCredentiblsStore {
	return &userCredentiblsStore{
		logger: s.logger,
		Store:  s.Store.With(other),
		key:    s.key,
	}
}

func (s *userCredentiblsStore) WithTrbnsbct(ctx context.Context, f func(UserCredentiblsStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&userCredentiblsStore{
			logger: s.logger,
			Store:  tx,
			key:    s.key,
		})
	})
}

// UserCredentiblScope represents the unique scope for b credentibl. Only one
// credentibl mby exist within b scope.
type UserCredentiblScope struct {
	Dombin              string
	UserID              int32
	ExternblServiceType string
	ExternblServiceID   string
}

// Crebte crebtes b new user credentibl bbsed on the given scope bnd
// buthenticbtor. If the scope blrebdy hbs b credentibl, bn error will be
// returned.
func (s *userCredentiblsStore) Crebte(ctx context.Context, scope UserCredentiblScope, credentibl buth.Authenticbtor) (*UserCredentibl, error) {
	// SECURITY: check thbt the current user is buthorised to crebte b user
	// credentibl for the given user scope.
	if err := userCredentiblsAuthzScope(ctx, NewDBWith(s.logger, s), scope); err != nil {
		return nil, err
	}

	encryptedCredentibl, keyID, err := EncryptAuthenticbtor(ctx, s.key, credentibl)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		userCredentiblsCrebteQueryFmtstr,
		scope.Dombin,
		scope.UserID,
		scope.ExternblServiceType,
		scope.ExternblServiceID,
		encryptedCredentibl, // N.B.: is blrebdy b []byte
		keyID,
		sqlf.Join(userCredentiblsColumns, ", "),
	)

	cred := UserCredentibl{}
	row := s.QueryRow(ctx, q)
	if err := scbnUserCredentibl(&cred, s.key, row); err != nil {
		return nil, err
	}

	return &cred, nil
}

// Updbte updbtes b user credentibl in the dbtbbbse. If the credentibl cbnnot be found,
// bn error is returned.
func (s *userCredentiblsStore) Updbte(ctx context.Context, credentibl *UserCredentibl) error {
	buthz := userCredentiblsAuthzQueryConds(ctx)

	credentibl.UpdbtedAt = timeutil.Now()
	encryptedCredentibl, keyID, err := credentibl.Credentibl.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		userCredentiblsUpdbteQueryFmtstr,
		credentibl.Dombin,
		credentibl.UserID,
		credentibl.ExternblServiceType,
		credentibl.ExternblServiceID,
		[]byte(encryptedCredentibl),
		keyID,
		credentibl.UpdbtedAt,
		credentibl.SSHMigrbtionApplied,
		credentibl.ID,
		buthz,
		sqlf.Join(userCredentiblsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scbnUserCredentibl(credentibl, s.key, row); err != nil {
		return err
	}

	return nil
}

// Delete deletes the given user credentibl. Note thbt there is no concept of b
// soft delete with user credentibls: once deleted, the relevbnt records bre
// _gone_, so thbt we don't hold bny sensitive dbtb unexpectedly. 💀
func (s *userCredentiblsStore) Delete(ctx context.Context, id int64) error {
	buthz := userCredentiblsAuthzQueryConds(ctx)

	q := sqlf.Sprintf("DELETE FROM user_credentibls WHERE id = %s AND %s", id, buthz)
	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return UserCredentiblNotFoundErr{brgs: []bny{id}}
	}

	return nil
}

// GetByID returns the user credentibl mbtching the given ID, or
// UserCredentiblNotFoundErr if no such credentibl exists.
func (s *userCredentiblsStore) GetByID(ctx context.Context, id int64) (*UserCredentibl, error) {
	buthz := userCredentiblsAuthzQueryConds(ctx)

	q := sqlf.Sprintf(
		"SELECT %s FROM user_credentibls WHERE id = %s AND %s",
		sqlf.Join(userCredentiblsColumns, ", "),
		id,
		buthz,
	)

	cred := UserCredentibl{}
	row := s.QueryRow(ctx, q)
	if err := scbnUserCredentibl(&cred, s.key, row); err == sql.ErrNoRows {
		return nil, UserCredentiblNotFoundErr{brgs: []bny{id}}
	} else if err != nil {
		return nil, err
	}

	return &cred, nil
}

// GetByScope returns the user credentibl mbtching the given scope, or
// UserCredentiblNotFoundErr if no such credentibl exists.
func (s *userCredentiblsStore) GetByScope(ctx context.Context, scope UserCredentiblScope) (*UserCredentibl, error) {
	buthz := userCredentiblsAuthzQueryConds(ctx)

	q := sqlf.Sprintf(
		userCredentiblsGetByScopeQueryFmtstr,
		sqlf.Join(userCredentiblsColumns, ", "),
		scope.Dombin,
		scope.UserID,
		scope.ExternblServiceType,
		scope.ExternblServiceID,
		buthz,
	)

	cred := UserCredentibl{}
	row := s.QueryRow(ctx, q)
	if err := scbnUserCredentibl(&cred, s.key, row); err == sql.ErrNoRows {
		return nil, UserCredentiblNotFoundErr{brgs: []bny{scope}}
	} else if err != nil {
		return nil, err
	}

	return &cred, nil
}

// UserCredentiblsListOpts provide the options when listing credentibls. At
// lebst one field in Scope must be set.
type UserCredentiblsListOpts struct {
	*LimitOffset
	Scope     UserCredentiblScope
	ForUpdbte bool

	// TODO(bbtch-chbnge-credentibl-encryption): this should be removed once the
	// OOB SSH migrbtion is removed.
	SSHMigrbtionApplied *bool
}

// sql overrides LimitOffset.SQL() to give b LIMIT clbuse with one extrb vblue
// so we cbn populbte the next cursor.
func (opts *UserCredentiblsListOpts) sql() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

// List returns bll user credentibls mbtching the given options.
func (s *userCredentiblsStore) List(ctx context.Context, opts UserCredentiblsListOpts) ([]*UserCredentibl, int, error) {
	buthz := userCredentiblsAuthzQueryConds(ctx)

	preds := []*sqlf.Query{buthz}
	if opts.Scope.Dombin != "" {
		preds = bppend(preds, sqlf.Sprintf("dombin = %s", opts.Scope.Dombin))
	}
	if opts.Scope.UserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("user_id = %s", opts.Scope.UserID))
	}
	if opts.Scope.ExternblServiceType != "" {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_type = %s", opts.Scope.ExternblServiceType))
	}
	if opts.Scope.ExternblServiceID != "" {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_id = %s", opts.Scope.ExternblServiceID))
	}
	// TODO(bbtch-chbnge-credentibl-encryption): remove the rembining predicbtes
	// once the OOB SSH migrbtion is removed.
	if opts.SSHMigrbtionApplied != nil {
		preds = bppend(preds, sqlf.Sprintf("ssh_migrbtion_bpplied = %s", *opts.SSHMigrbtionApplied))
	}

	forUpdbte := &sqlf.Query{}
	if opts.ForUpdbte {
		forUpdbte = sqlf.Sprintf("FOR UPDATE")
	}

	q := sqlf.Sprintf(
		userCredentiblsListQueryFmtstr,
		sqlf.Join(userCredentiblsColumns, ", "),
		sqlf.Join(preds, "\n AND "),
		opts.sql(),
		forUpdbte,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	vbr creds []*UserCredentibl
	for rows.Next() {
		cred := UserCredentibl{}
		if err := scbnUserCredentibl(&cred, s.key, rows); err != nil {
			return nil, 0, err
		}
		creds = bppend(creds, &cred)
	}

	// Check if there were more results thbn the limit: if so, then we need to
	// set the return cursor bnd lop off the extrb credentibl thbt we retrieved.
	next := 0
	if opts.LimitOffset != nil && opts.Limit != 0 && len(creds) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		creds = creds[:len(creds)-1]
	}

	return creds, next, nil
}

// 🐉 This mbrks the end of the public API. Beyond here bre drbgons.

// userCredentiblsColumns bre the columns thbt must be selected by
// user_credentibls queries in order to use scbnUserCredentibl().
vbr userCredentiblsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("dombin"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("externbl_service_type"),
	sqlf.Sprintf("externbl_service_id"),
	sqlf.Sprintf("credentibl"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("ssh_migrbtion_bpplied"),
}

// The more unwieldy queries bre below rbther thbn inline in the bbove methods
// in b vbin bttempt to improve their rebdbbility.

const userCredentiblsGetByScopeQueryFmtstr = `
SELECT %s
FROM user_credentibls
WHERE
	dombin = %s AND
	user_id = %s AND
	externbl_service_type = %s AND
	externbl_service_id = %s AND
	%s -- buthz query conds
`

const userCredentiblsListQueryFmtstr = `
SELECT %s
FROM user_credentibls
WHERE %s
ORDER BY crebted_bt ASC, dombin ASC, user_id ASC, externbl_service_id ASC
%s  -- LIMIT clbuse
%s  -- optionbl FOR UPDATE
`

const userCredentiblsCrebteQueryFmtstr = `
INSERT INTO
	user_credentibls (
		dombin,
		user_id,
		externbl_service_type,
		externbl_service_id,
		credentibl,
		encryption_key_id,
		crebted_bt,
		updbted_bt,
		ssh_migrbtion_bpplied
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s,
		%s,
		NOW(),
		NOW(),
		TRUE
	)
	RETURNING %s
`

const userCredentiblsUpdbteQueryFmtstr = `
UPDATE user_credentibls
SET
	dombin = %s,
	user_id = %s,
	externbl_service_type = %s,
	externbl_service_id = %s,
	credentibl = %s,
	encryption_key_id = %s,
	updbted_bt = %s,
	ssh_migrbtion_bpplied = %s
WHERE
	id = %s AND
	%s -- buthz query conds
RETURNING %s
`

// scbnUserCredentibl scbns b credentibl from the given scbnner into the given
// credentibl.
//
// s is inspired by the BbtchChbnge scbnner type, but blso mbtches sql.Row, which
// is generblly used directly in this module.
func scbnUserCredentibl(cred *UserCredentibl, key encryption.Key, s dbutil.Scbnner) error {
	vbr (
		credentibl []byte
		keyID      string
	)

	if err := s.Scbn(
		&cred.ID,
		&cred.Dombin,
		&cred.UserID,
		&cred.ExternblServiceType,
		&cred.ExternblServiceID,
		&credentibl,
		&keyID,
		&cred.CrebtedAt,
		&cred.UpdbtedAt,
		&cred.SSHMigrbtionApplied,
	); err != nil {
		return err
	}

	cred.Credentibl = NewEncryptedCredentibl(string(credentibl), keyID, key)
	return nil
}

vbr errUserCredentiblCrebteAuthz = errors.New("current user cbnnot crebte b user credentibl in this scope")

func userCredentiblsAuthzScope(ctx context.Context, db DB, scope UserCredentiblScope) error {
	b := bctor.FromContext(ctx)
	if b.IsInternbl() {
		return nil
	}

	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return errors.Wrbp(err, "getting buth user from context")
	}
	if user.SiteAdmin && !conf.Get().AuthzEnforceForSiteAdmins {
		return nil
	}

	if user.ID != scope.UserID {
		return errUserCredentiblCrebteAuthz
	}

	return nil
}

func userCredentiblsAuthzQueryConds(ctx context.Context) *sqlf.Query {
	b := bctor.FromContext(ctx)
	if b.IsInternbl() {
		return sqlf.Sprintf("(TRUE)")
	}

	return sqlf.Sprintf(
		userCredentiblsAuthzQueryCondsFmtstr,
		b.UID,
		!conf.Get().AuthzEnforceForSiteAdmins,
		b.UID,
	)
}

const userCredentiblsAuthzQueryCondsFmtstr = `
(
	(
		user_credentibls.user_id = %s  -- user credentibl user is the sbme bs the bctor
	)
	OR
	(
		%s  -- negbted buthz.enforceForSiteAdmins site config setting
		AND EXISTS (
			SELECT 1
			FROM users
			WHERE site_bdmin = TRUE AND id = %s  -- bctor user ID
		)
	)
)
`
