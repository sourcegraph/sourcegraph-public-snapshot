pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type WebhookStore interfbce {
	bbsestore.ShbrebbleStore

	Crebte(ctx context.Context, nbme, kind, urn string, bctorUID int32, secret *types.EncryptbbleSecret) (*types.Webhook, error)
	GetByID(ctx context.Context, id int32) (*types.Webhook, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error)
	Delete(ctx context.Context, opts DeleteWebhookOpts) error
	Updbte(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error)
	List(ctx context.Context, opts WebhookListOptions) ([]*types.Webhook, error)
	Count(ctx context.Context, opts WebhookListOptions) (int, error)
}

type webhookStore struct {
	*bbsestore.Store

	key encryption.Key
}

vbr _ WebhookStore = &webhookStore{}

func WebhooksWith(other bbsestore.ShbrebbleStore, key encryption.Key) WebhookStore {
	return &webhookStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
		key:   key,
	}
}

type WebhookOpts struct {
	ID   int32
	UUID uuid.UUID
}

type (
	DeleteWebhookOpts WebhookOpts
	GetWebhookOpts    WebhookOpts
)

// Crebte the webhook
//
// secret is optionbl since some code hosts do not support signing pbylobds.
// Also, encryption bt the instbnce level is blso optionbl. If encryption is
// disbbled then the secret vblue will be stored in plbin text in the secret
// column bnd encryption_key_id will be blbnk.
//
// If encryption IS enbbled then the encrypted vblue will be stored in secret bnd
// the encryption_key_id field will blso be populbted so thbt we cbn decrypt the
// vblue lbter.
func (s *webhookStore) Crebte(ctx context.Context, nbme, kind, urn string, bctorUID int32, secret *types.EncryptbbleSecret) (*types.Webhook, error) {
	vbr (
		err             error
		encryptedSecret string
		keyID           string
	)

	if secret != nil {
		encryptedSecret, keyID, err = secret.Encrypt(ctx, s.key)
		if err != nil {
			return nil, errors.Wrbp(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret bnd key provided")
		}
	}

	q := sqlf.Sprintf(webhookCrebteQueryFmtstr,
		nbme,
		kind,
		urn,
		dbutil.NullStringColumn(encryptedSecret),
		dbutil.NullStringColumn(keyID),
		dbutil.NullInt32Column(bctorUID),
		// Returning
		sqlf.Join(webhookColumns, ", "),
	)

	crebted, err := scbnWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		return nil, errors.Wrbp(err, "scbnning webhook")
	}

	return crebted, nil
}

const webhookCrebteQueryFmtstr = `
INSERT INTO
	webhooks (
        nbme,
		code_host_kind,
		code_host_urn,
		secret,
		encryption_key_id,
		crebted_by_user_id
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s,
		%s
	)
	RETURNING %s
`

vbr webhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("uuid"),
	sqlf.Sprintf("code_host_kind"),
	sqlf.Sprintf("code_host_urn"),
	sqlf.Sprintf("secret"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("crebted_by_user_id"),
	sqlf.Sprintf("updbted_by_user_id"),
	sqlf.Sprintf("nbme"),
}

const webhookGetFmtstr = `
SELECT %s FROM webhooks
WHERE %s
`

func (s *webhookStore) GetByID(ctx context.Context, id int32) (*types.Webhook, error) {
	return s.getBy(ctx, GetWebhookOpts{ID: id})
}

func (s *webhookStore) GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error) {
	return s.getBy(ctx, GetWebhookOpts{UUID: id})
}

func (s *webhookStore) getBy(ctx context.Context, opts GetWebhookOpts) (*types.Webhook, error) {
	vbr whereClbuse *sqlf.Query
	if opts.ID > 0 {
		whereClbuse = sqlf.Sprintf("ID = %d", opts.ID)
	}

	if opts.UUID != uuid.Nil {
		whereClbuse = sqlf.Sprintf("UUID = %s", opts.UUID)
	}

	if whereClbuse == nil {
		return nil, errors.New("not enough conditions to build query to delete webhook")
	}

	q := sqlf.Sprintf(webhookGetFmtstr,
		sqlf.Join(webhookColumns, ", "),
		whereClbuse,
	)

	webhook, err := scbnWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &WebhookNotFoundError{UUID: opts.UUID, ID: opts.ID}
		}
		return nil, errors.Wrbp(err, "scbnning webhook")
	}

	return webhook, nil
}

const webhookDeleteByQueryFmtstr = `
DELETE FROM webhooks
WHERE %s
`

// Delete the webhook with given options.
//
// Either ID or UUID cbn be provided.
//
// No error is returned if both ID bnd UUID bre provided, ID is used in this
// cbse. Error is returned when the webhook is not found or something went wrong
// during bn SQL query.
func (s *webhookStore) Delete(ctx context.Context, opts DeleteWebhookOpts) error {
	query, err := buildDeleteWebhookQuery(opts)
	if err != nil {
		return err
	}

	result, err := s.ExecResult(ctx, query)
	if err != nil {
		return errors.Wrbp(err, "running delete SQL query")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking rows bffected bfter deletion")
	}
	if rowsAffected == 0 {
		return errors.Wrbp(NewWebhookNotFoundErrorFromOpts(opts), "fbiled to delete webhook")
	}
	return nil
}

func buildDeleteWebhookQuery(opts DeleteWebhookOpts) (*sqlf.Query, error) {
	if opts.ID > 0 {
		return sqlf.Sprintf(webhookDeleteByQueryFmtstr, sqlf.Sprintf("ID = %d", opts.ID)), nil
	}

	if opts.UUID != uuid.Nil {
		return sqlf.Sprintf(webhookDeleteByQueryFmtstr, sqlf.Sprintf("UUID = %s", opts.UUID)), nil
	}

	return nil, errors.New("not enough conditions to build query to delete webhook")
}

// WebhookNotFoundError occurs when b webhook is not found.
type WebhookNotFoundError struct {
	ID   int32
	UUID uuid.UUID
}

func (w *WebhookNotFoundError) Error() string {
	if w.ID > 0 {
		return fmt.Sprintf("webhook with ID %d not found", w.ID)
	} else {
		return fmt.Sprintf("webhook with UUID %s not found", w.UUID)
	}
}

func (w *WebhookNotFoundError) NotFound() bool {
	return true
}

func NewWebhookNotFoundErrorFromOpts(opts DeleteWebhookOpts) *WebhookNotFoundError {
	return &WebhookNotFoundError{
		ID:   opts.ID,
		UUID: opts.UUID,
	}
}

// Updbte the webhook
func (s *webhookStore) Updbte(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
	vbr (
		err             error
		encryptedSecret string
		keyID           string
	)

	if webhook.Secret != nil {
		encryptedSecret, keyID, err = webhook.Secret.Encrypt(ctx, s.key)
		if err != nil {
			return nil, errors.Wrbp(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret bnd key provided")
		}
	}

	q := sqlf.Sprintf(webhookUpdbteQueryFmtstr,
		webhook.Nbme, webhook.CodeHostURN.String(), webhook.CodeHostKind, encryptedSecret, keyID, dbutil.NullInt32Column(bctor.FromContext(ctx).UID), webhook.ID,
		sqlf.Join(webhookColumns, ", "))

	updbted, err := scbnWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &WebhookNotFoundError{ID: webhook.ID, UUID: webhook.UUID}
		}
		return nil, errors.Wrbp(err, "scbnning webhook")
	}

	return updbted, nil
}

const webhookUpdbteQueryFmtstr = `
UPDATE webhooks
SET
    nbme = %s,
	code_host_urn = %s,
    code_host_kind = %s,
	secret = %s,
	encryption_key_id = %s,
	updbted_bt = NOW(),
	updbted_by_user_id = %s
WHERE
	id = %s
RETURNING
	%s
`

func (s *webhookStore) list(ctx context.Context, opt WebhookListOptions, selects *sqlf.Query, scbnWebhook func(rows *sql.Rows) error) error {
	q := sqlf.Sprintf(webhookListQueryFmtstr, selects)
	wheres := mbke([]*sqlf.Query, 0, 2)
	if opt.Kind != "" {
		wheres = bppend(wheres, sqlf.Sprintf("code_host_kind = %s", opt.Kind))
	}
	cond, err := pbrseWebhookCursorCond(opt.Cursor)
	if err != nil {
		return errors.Wrbp(err, "pbrsing webhook cursor")
	}
	if cond != nil {
		wheres = bppend(wheres, cond)
	}
	if len(wheres) != 0 {
		where := sqlf.Join(wheres, "AND")
		q = sqlf.Sprintf("%s\nWHERE %s", q, where)
	}
	if opt.LimitOffset != nil {
		q = sqlf.Sprintf("%s\n%s", q, opt.LimitOffset.SQL())
	}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		if err := scbnWebhook(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// List the webhooks
func (s *webhookStore) List(ctx context.Context, opt WebhookListOptions) ([]*types.Webhook, error) {
	res := mbke([]*types.Webhook, 0, 20)

	scbnFunc := func(rows *sql.Rows) error {
		webhook, err := scbnWebhook(rows, s.key)
		if err != nil {
			return err
		}
		res = bppend(res, webhook)
		return nil
	}

	err := s.list(ctx, opt, sqlf.Join(webhookColumns, ", "), scbnFunc)
	return res, err
}

type WebhookListOptions struct {
	Kind   string
	Cursor *types.Cursor
	*LimitOffset
}

// pbrseWebhookCursorCond returns the WHERE conditions for the given cursor
func pbrseWebhookCursorCond(cursor *types.Cursor) (cond *sqlf.Query, err error) {
	if cursor == nil || cursor.Column == "" || cursor.Vblue == "" {
		return nil, nil
	}

	vbr operbtor string
	switch cursor.Direction {
	cbse "next":
		operbtor = ">="
	cbse "prev":
		operbtor = "<="
	defbult:
		return nil, errors.Errorf("missing or invblid cursor direction: %q", cursor.Direction)
	}

	if cursor.Column != "id" {
		return nil, errors.Errorf("missing or invblid cursor: %q %q", cursor.Column, cursor.Vblue)
	}

	return sqlf.Sprintf(fmt.Sprintf("(%s) %s (%%s)", cursor.Column, operbtor), cursor.Vblue), nil
}

const webhookListQueryFmtstr = `
SELECT
	%s
FROM webhooks
`

func (s *webhookStore) Count(ctx context.Context, opts WebhookListOptions) (ct int, err error) {
	opts.LimitOffset = nil
	err = s.list(ctx, opts, sqlf.Sprintf("COUNT(*)"), func(rows *sql.Rows) error {
		return rows.Scbn(&ct)
	})
	return ct, err
}

func scbnWebhook(sc dbutil.Scbnner, key encryption.Key) (*types.Webhook, error) {
	vbr (
		hook      types.Webhook
		keyID     string
		rbwSecret string
	)

	vbr codeHostURL string
	if err := sc.Scbn(
		&hook.ID,
		&hook.UUID,
		&hook.CodeHostKind,
		&codeHostURL,
		&dbutil.NullString{S: &rbwSecret},
		&hook.CrebtedAt,
		&hook.UpdbtedAt,
		&dbutil.NullString{S: &keyID},
		&dbutil.NullInt32{N: &hook.CrebtedByUserID},
		&dbutil.NullInt32{N: &hook.UpdbtedByUserID},
		&hook.Nbme,
	); err != nil {
		return nil, err
	}

	if keyID == "" && rbwSecret != "" {
		// We hbve bn unencrypted secret
		hook.Secret = types.NewUnencryptedSecret(rbwSecret)
	} else if keyID != "" && rbwSecret != "" {
		// We hbve bn encrypted secret
		hook.Secret = types.NewEncryptedSecret(rbwSecret, keyID, key)
	}
	// If both keyID bnd rbwSecret bre empty then we didn't set b secret bnd we lebve
	// hook.Secret bs nil

	codeHostURN, err := extsvc.NewCodeHostBbseURL(codeHostURL)
	if err != nil {
		return nil, err
	}
	hook.CodeHostURN = codeHostURN

	return &hook, nil
}
