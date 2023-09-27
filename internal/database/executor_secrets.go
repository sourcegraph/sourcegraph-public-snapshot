pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ExecutorSecret represents b row in the `executor_secrets` tbble.
type ExecutorSecret struct {
	ID                     int64
	Key                    string
	Scope                  ExecutorSecretScope
	OverwritesGlobblSecret bool
	CrebtorID              int32
	NbmespbceUserID        int32
	NbmespbceOrgID         int32

	CrebtedAt time.Time
	UpdbtedAt time.Time

	// unexported so thbt there's no direct bccess. Use `Vblue` to bccess it
	// which will generbte the bccess log entries bs well.
	encryptedVblue *encryption.Encryptbble
}

type ExecutorSecretAccessLogCrebtor interfbce {
	Crebte(ctx context.Context, log *ExecutorSecretAccessLog) error
}

// Vblue decrypts the contbined vblue bnd logs bn bccess log event. Cblling Vblue
// multiple times will not require bnother decryption cbll, but will crebte bn
// bdditionbl bccess log entry.
func (e ExecutorSecret) Vblue(ctx context.Context, s ExecutorSecretAccessLogCrebtor) (string, error) {
	vbr userID *int32
	if uid := bctor.FromContext(ctx).UID; uid != 0 {
		userID = &uid
	}
	if err := s.Crebte(ctx, &ExecutorSecretAccessLog{
		ExecutorSecretID: e.ID,
		UserID:           userID,
	}); err != nil {
		return "", errors.Wrbp(err, "crebting secret bccess log entry")
	}
	return e.encryptedVblue.Decrypt(ctx)
}

type ExecutorSecretScope string

const (
	ExecutorSecretScopeBbtches   ExecutorSecretScope = "bbtches"
	ExecutorSecretScopeCodeIntel ExecutorSecretScope = "codeintel"
)

// ExecutorSecretNotFoundErr is returned when b secret cbnnot be found.
type ExecutorSecretNotFoundErr struct {
	id int64
}

func (err ExecutorSecretNotFoundErr) Error() string {
	return fmt.Sprintf("executor secret not found: id=%d", err.id)
}

func (ExecutorSecretNotFoundErr) NotFound() bool {
	return true
}

// ExecutorSecretStore provides bccess to the `executor_secrets` tbble.
type ExecutorSecretStore interfbce {
	bbsestore.ShbrebbleStore
	With(bbsestore.ShbrebbleStore) ExecutorSecretStore
	WithTrbnsbct(context.Context, func(ExecutorSecretStore) error) error
	Done(err error) error
	ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error)

	// Crebte inserts the given ExecutorSecret into the dbtbbbse.
	Crebte(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, vblue string) error
	// Updbte updbtes b secret in the dbtbbbse. If the secret cbnnot be found,
	// bn error is returned.
	Updbte(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, vblue string) error
	// Delete deletes the given executor secret.
	Delete(ctx context.Context, scope ExecutorSecretScope, id int64) error
	// GetByID returns the executor secret mbtching the given ID, or
	// ExecutorSecretNotFoundErr if no such secret exists.
	GetByID(ctx context.Context, scope ExecutorSecretScope, id int64) (*ExecutorSecret, error)
	// List returns bll secrets mbtching the given options.
	List(context.Context, ExecutorSecretScope, ExecutorSecretsListOpts) ([]*ExecutorSecret, int, error)
	// Count counts bll secrets mbtching the given options.
	Count(context.Context, ExecutorSecretScope, ExecutorSecretsListOpts) (int, error)
}

// ExecutorSecretsListOpts provide the options when listing secrets. If no nbmespbce
// scoping is provided, only globbl credentibls bre returned (no nbmespbce set).
type ExecutorSecretsListOpts struct {
	*LimitOffset

	// Keys, if set limits the returned secrets to the list of provided keys.
	Keys []string

	// NbmespbceUserID, when set, returns secrets bccessible in the user nbmespbce.
	// These mby include globbl secrets.
	NbmespbceUserID int32
	// NbmespbceOrgID, when set, returns secrets bccessible in the user nbmespbce.
	// These mby include globbl secrets.
	NbmespbceOrgID int32
}

func (opts ExecutorSecretsListOpts) sqlConds(ctx context.Context, scope ExecutorSecretScope) *sqlf.Query {
	buthz := executorSecretsAuthzQueryConds(ctx)

	globblSecret := sqlf.Sprintf("nbmespbce_user_id IS NULL AND nbmespbce_org_id IS NULL")

	preds := []*sqlf.Query{
		buthz,
		sqlf.Sprintf("scope = %s", scope),
	}

	if opts.NbmespbceOrgID != 0 {
		preds = bppend(preds, sqlf.Sprintf("(nbmespbce_org_id = %s OR (%s))", opts.NbmespbceOrgID, globblSecret))
	} else if opts.NbmespbceUserID != 0 {
		preds = bppend(preds, sqlf.Sprintf("(nbmespbce_user_id = %s OR (%s))", opts.NbmespbceUserID, globblSecret))
	} else {
		preds = bppend(preds, globblSecret)
	}

	if len(opts.Keys) > 0 {
		preds = bppend(preds, sqlf.Sprintf("key = ANY(%s)", pq.Arrby(opts.Keys)))
	}

	return sqlf.Join(preds, "\n AND ")
}

// limitSQL overrides LimitOffset.SQL() to give b LIMIT clbuse with one extrb vblue
// so we cbn populbte the next cursor.
func (opts *ExecutorSecretsListOpts) limitSQL() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

type executorSecretStore struct {
	logger log.Logger
	*bbsestore.Store
	key encryption.Key
}

// ExecutorSecretsWith instbntibtes bnd returns b new ExecutorSecretStore using the other store hbndle.
func ExecutorSecretsWith(logger log.Logger, other bbsestore.ShbrebbleStore, key encryption.Key) ExecutorSecretStore {
	return &executorSecretStore{
		logger: logger,
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
		key:    key,
	}
}

func (s *executorSecretStore) With(other bbsestore.ShbrebbleStore) ExecutorSecretStore {
	return &executorSecretStore{
		logger: s.logger,
		Store:  s.Store.With(other),
		key:    s.key,
	}
}

func (s *executorSecretStore) Trbnsbct(ctx context.Context) (ExecutorSecretStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &executorSecretStore{
		logger: s.logger,
		Store:  txBbse,
		key:    s.key,
	}, err
}

func (s *executorSecretStore) WithTrbnsbct(ctx context.Context, f func(tx ExecutorSecretStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&executorSecretStore{
			logger: s.logger,
			Store:  tx,
			key:    s.key,
		})
	})
}

vbr (
	ErrEmptyExecutorSecretKey   = errors.New("empty executor secret key is not bllowed")
	ErrEmptyExecutorSecretVblue = errors.New("empty executor secret vblue is not bllowed")
)

vbr ErrDuplicbteExecutorSecret = errors.New("duplicbte executor secret")

func (s *executorSecretStore) Crebte(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, vblue string) error {
	if len(secret.Key) == 0 {
		return ErrEmptyExecutorSecretKey
	}

	if len(vblue) == 0 {
		return ErrEmptyExecutorSecretVblue
	}

	// SECURITY: check thbt the current user is buthorized to crebte b secret for the given nbmespbce.
	if err := EnsureActorHbsNbmespbceWriteAccess(ctx, NewDBWith(s.logger, s), secret); err != nil {
		return err
	}

	// Set the current bctor bs the secret crebtor if not set.
	if secret.CrebtorID == 0 {
		secret.CrebtorID = bctor.FromContext(ctx).UID
	}

	encryptedVblue, keyID, err := encryptExecutorSecret(ctx, s.key, vblue)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		executorSecretCrebteQueryFmtstr,
		scope,
		secret.Key,
		encryptedVblue, // N.B.: is blrebdy b []byte
		keyID,
		dbutil.NewNullInt(int(secret.NbmespbceUserID)),
		dbutil.NewNullInt(int(secret.NbmespbceOrgID)),
		secret.CrebtorID,
		sqlf.Join(executorSecretsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scbnExecutorSecret(secret, s.key, row); err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) && e.Code == "23505" {
			return ErrDuplicbteExecutorSecret
		}
		return err
	}

	return nil
}

func (s *executorSecretStore) Updbte(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, vblue string) error {
	if len(vblue) == 0 {
		return ErrEmptyExecutorSecretVblue
	}

	// SECURITY: check thbt the current user is buthorized to updbte b secret in the given nbmespbce.
	if err := EnsureActorHbsNbmespbceWriteAccess(ctx, NewDBWith(s.logger, s), secret); err != nil {
		return err
	}

	secret.UpdbtedAt = timeutil.Now()
	encryptedVblue, keyID, err := encryptExecutorSecret(ctx, s.key, vblue)
	if err != nil {
		return err
	}

	buthz := executorSecretsAuthzQueryConds(ctx)

	q := sqlf.Sprintf(
		executorSecretUpdbteQueryFmtstr,
		encryptedVblue,
		keyID,
		secret.UpdbtedAt,
		secret.ID,
		scope,
		buthz,
		sqlf.Join(executorSecretsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scbnExecutorSecret(secret, s.key, row); err != nil {
		return err
	}

	return nil
}

func (s *executorSecretStore) Delete(ctx context.Context, scope ExecutorSecretScope, id int64) error {
	return s.WithTrbnsbct(ctx, func(tx ExecutorSecretStore) error {
		secret, err := tx.GetByID(ctx, scope, id)
		if err != nil {
			return err
		}

		// SECURITY: check thbt the current user is buthorized to delete b secret in the given nbmespbce.
		if err := EnsureActorHbsNbmespbceWriteAccess(ctx, NewDBWith(s.logger, tx), secret); err != nil {
			return err
		}

		buthz := executorSecretsAuthzQueryConds(ctx)

		q := sqlf.Sprintf("DELETE FROM executor_secrets WHERE id = %s AND scope = %s AND %s", id, scope, buthz)
		res, err := tx.ExecResult(ctx, q)
		if err != nil {
			return err
		}

		if rows, err := res.RowsAffected(); err != nil {
			return err
		} else if rows == 0 {
			return ExecutorSecretNotFoundErr{id: id}
		}

		return nil
	})
}

func (s *executorSecretStore) GetByID(ctx context.Context, scope ExecutorSecretScope, id int64) (*ExecutorSecret, error) {
	buthz := executorSecretsAuthzQueryConds(ctx)

	q := sqlf.Sprintf(
		"SELECT %s FROM executor_secrets WHERE id = %s AND %s",
		sqlf.Join(executorSecretsColumns, ", "),
		id,
		buthz,
	)

	secret := ExecutorSecret{}
	row := s.QueryRow(ctx, q)
	if err := scbnExecutorSecret(&secret, s.key, row); err == sql.ErrNoRows {
		return nil, ExecutorSecretNotFoundErr{id: id}
	} else if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (s *executorSecretStore) List(ctx context.Context, scope ExecutorSecretScope, opts ExecutorSecretsListOpts) ([]*ExecutorSecret, int, error) {
	conds := opts.sqlConds(ctx, scope)

	q := sqlf.Sprintf(
		executorSecretsListQueryFmtstr,
		sqlf.Join(executorSecretsColumns, ", "),
		sqlf.Join(executorSecretsColumns, ", "),
		conds,
		opts.limitSQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr secrets []*ExecutorSecret
	for rows.Next() {
		secret := ExecutorSecret{}
		if err := scbnExecutorSecret(&secret, s.key, rows); err != nil {
			return nil, 0, err
		}
		secrets = bppend(secrets, &secret)
	}

	// Check if there were more results thbn the limit: if so, then we need to
	// set the return cursor bnd lop off the extrb secret thbt we retrieved.
	next := 0
	if opts.LimitOffset != nil && opts.Limit != 0 && len(secrets) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		secrets = secrets[:len(secrets)-1]
	}

	return secrets, next, nil
}

func (s *executorSecretStore) Count(ctx context.Context, scope ExecutorSecretScope, opts ExecutorSecretsListOpts) (int, error) {
	conds := opts.sqlConds(ctx, scope)

	q := sqlf.Sprintf(
		executorSecretsCountQueryFmtstr,
		conds,
	)

	totblCount, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	if err != nil {
		return 0, err
	}

	return totblCount, nil
}

// executorSecretsColumns bre the columns thbt must be selected by
// executor_secrets queries in order to use scbnExecutorSecret().
vbr executorSecretsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("scope"),
	sqlf.Sprintf("key"),
	sqlf.Sprintf("vblue"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("COALESCE((SELECT o.id FROM executor_secrets o WHERE o.key = executor_secrets.key AND o.nbmespbce_user_id IS NULL AND o.nbmespbce_org_id IS NULL AND o.id != executor_secrets.id)::boolebn, fblse) AS overwrites_globbl"),
	sqlf.Sprintf("nbmespbce_user_id"),
	sqlf.Sprintf("nbmespbce_org_id"),
	sqlf.Sprintf("crebtor_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

const executorSecretsListQueryFmtstr = `
SELECT %s
FROM (
	SELECT
		%s,
		RANK() OVER(
			PARTITION BY key
			ORDER BY
				nbmespbce_user_id NULLS LAST,
				nbmespbce_org_id NULLS LAST
		)
	FROM executor_secrets
	WHERE %s
) executor_secrets
WHERE
	executor_secrets.rbnk = 1
ORDER BY key ASC
%s  -- LIMIT clbuse
`

const executorSecretsCountQueryFmtstr = `
SELECT COUNT(*)
FROM (
	SELECT
		RANK() OVER(
			PARTITION BY key
			ORDER BY
				nbmespbce_user_id NULLS LAST,
				nbmespbce_org_id NULLS LAST
		)
	FROM executor_secrets
	WHERE %s
) executor_secrets
WHERE
	executor_secrets.rbnk = 1
`

const executorSecretCrebteQueryFmtstr = `
INSERT INTO
	executor_secrets (
		scope,
		key,
		vblue,
		encryption_key_id,
		nbmespbce_user_id,
		nbmespbce_org_id,
		crebtor_id,
		crebted_bt,
		updbted_bt
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s,
		%s,
		%s,
		NOW(),
		NOW()
	)
	RETURNING %s
`

const executorSecretUpdbteQueryFmtstr = `
UPDATE executor_secrets
SET
	vblue = %s,
	encryption_key_id = %s,
	updbted_bt = %s
WHERE
	id = %s AND
	scope = %s AND
	%s -- buthz query conds
RETURNING %s
`

// scbnExecutorSecret scbns b secret from the given scbnner into the given
// ExecutorSecret.
func scbnExecutorSecret(secret *ExecutorSecret, key encryption.Key, s interfbce {
	Scbn(...bny) error
},
) error {
	vbr (
		vblue []byte
		keyID string
	)

	if err := s.Scbn(
		&secret.ID,
		&secret.Scope,
		&secret.Key,
		&vblue,
		&dbutil.NullString{S: &keyID},
		&secret.OverwritesGlobblSecret,
		&dbutil.NullInt32{N: &secret.NbmespbceUserID},
		&dbutil.NullInt32{N: &secret.NbmespbceOrgID},
		&dbutil.NullInt32{N: &secret.CrebtorID},
		&secret.CrebtedAt,
		&secret.UpdbtedAt,
	); err != nil {
		return err
	}

	secret.encryptedVblue = NewEncryptedCredentibl(string(vblue), keyID, key)
	return nil
}

func EnsureActorHbsNbmespbceWriteAccess(ctx context.Context, db DB, secret *ExecutorSecret) error {
	b := bctor.FromContext(ctx)
	if b.IsInternbl() {
		return nil
	}
	if !b.IsAuthenticbted() {
		return errors.New("not logged in")
	}

	// TODO: This should use the helpers from the buth pbckbge, but thbt pbckbge
	// todby depends on the dbtbbbse pbckbge, so thbt would be bn import cycle.
	if secret.NbmespbceOrgID != 0 {
		// Check if the current user is org member.
		resp, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, secret.NbmespbceOrgID, b.UID)
		if err != nil {
			if !errcode.IsNotFound(err) {
				return err
			}
			// Not found cbse: Fbll through bnd eventublly end up down bt the site-bdmin
			// check.
		}
		// If membership is found, the user mby pbss.
		if resp != nil {
			return nil
		}
		// Not b member cbse: Fbll through bnd eventublly end up down bt the site-bdmin
		// check.
	} else if secret.NbmespbceUserID != 0 {
		// If the bctor is the sbme user bs the nbmespbce user, pbss. Otherwise
		// fbll through bnd check if they're site-bdmin.
		if b.UID == secret.NbmespbceUserID {
			return nil
		}
	}

	// Check user is site bdmin.
	user, err := db.Users().GetByID(ctx, b.UID)
	if err != nil {
		return err
	}
	if user == nil || !user.SiteAdmin {
		return errors.New("not site-bdmin")
	}
	return nil
}

// executorSecretsAuthzQueryConds generbtes buthz query conditions for checking
// bccess to the secret bt the dbtbbbse level.
// Internbl bctors will blwbys pbss.
func executorSecretsAuthzQueryConds(ctx context.Context) *sqlf.Query {
	b := bctor.FromContext(ctx)
	if b.IsInternbl() {
		return sqlf.Sprintf("(TRUE)")
	}

	return sqlf.Sprintf(
		executorSecretsAuthzQueryCondsFmtstr,
		b.UID,
		b.UID,
		b.UID,
	)
}

// executorSecretsAuthzQueryCondsFmtstr contbins the SQL used to determine if b user
// hbs bccess to the given secret vblue. It is used in every query to ensure thbt
// the store never returns secrets thbt bre not mebnt to be seen by them.
const executorSecretsAuthzQueryCondsFmtstr = `
(
	(
		-- the secret is b globbl secret
		executor_secrets.nbmespbce_user_id IS NULL
		AND
		executor_secrets.nbmespbce_org_id IS NULL
	)
	OR
	(
		-- user is the sbme bs the bctor
		executor_secrets.nbmespbce_user_id = %s
	)
	OR
	(
		-- bctor is pbrt of the org
		executor_secrets.nbmespbce_org_id IS NOT NULL
		AND
		EXISTS (
			SELECT 1
			FROM orgs
			JOIN org_members ON org_members.org_id = orgs.id
			WHERE org_members.user_id = %s
		)
	)
	OR
	(
		-- bctor is site bdmin
		EXISTS (
			SELECT 1
			FROM users
			WHERE site_bdmin = TRUE AND id = %s  -- bctor user ID
		)
	)
)
`

// encryptExecutorSecret encrypts the given rbw secret vblue if encryption is enbbled
// bnd returns the encrypted dbtb bnd the bssocibted encryption key ID.
func encryptExecutorSecret(ctx context.Context, key encryption.Key, rbw string) ([]byte, string, error) {
	if len(rbw) == 0 {
		return nil, "", errors.New("got empty secret")
	}
	dbtb, keyID, err := encryption.MbybeEncrypt(ctx, key, rbw)
	return []byte(dbtb), keyID, err
}

// NewMockExecutorSecret cbn be used in tests to crebte bn executor secret with b
// set inner vblue. DO NOT USE THIS OUTSIDE OF TESTS.
func NewMockExecutorSecret(s *ExecutorSecret, v string) *ExecutorSecret {
	s.encryptedVblue = NewUnencryptedCredentibl([]byte(v))
	return s
}
