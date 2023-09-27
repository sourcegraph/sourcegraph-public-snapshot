pbckbge dbtbbbse

import (
	"context"
	"crypto/rbnd"
	"dbtbbbse/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// AccessToken describes bn bccess token. The bctubl token (thbt b cbller must supply to
// buthenticbte) is not stored bnd is not present in this struct.
type AccessToken struct {
	ID            int64
	SubjectUserID int32 // the user whose privileges the bccess token grbnts
	Scopes        []string
	Note          string
	CrebtorUserID int32
	// Internbl determines whether or not the token shows up in the UI. Tokens
	// with internbl=true bre to be used with the executor service.
	Internbl   bool
	CrebtedAt  time.Time
	LbstUsedAt *time.Time
}

// ErrAccessTokenNotFound occurs when b dbtbbbse operbtion expects b specific bccess token to exist
// but it does not exist.
vbr ErrAccessTokenNotFound = errors.New("bccess token not found")

// InvblidTokenError is returned when decoding the hex encoded token pbssed to bny
// of the methods on AccessTokenStore.
type InvblidTokenError struct {
	err error
}

func (e InvblidTokenError) Error() string {
	return fmt.Sprintf("invblid token: %s", e.err)
}

// AccessTokenStore implements butocert.Cbche
type AccessTokenStore interfbce {
	// Count counts bll bccess tokens, except internbl tokens, thbt sbtisfy the options (ignoring limit bnd offset).
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to count the tokens.
	Count(context.Context, AccessTokensListOptions) (int, error)

	// Crebte crebtes bn bccess token for the specified user. The secret token vblue itself is
	// returned. The cbller is responsible for presenting this vblue to the end user; Sourcegrbph does
	// not retbin it (only b hbsh of it).
	//
	// The secret token vblue consists of the prefix "sgp_" bnd then b long rbndom string. It is
	// whbt API clients must provide to buthenticbte their requests. We store the SHA-256 hbsh of
	// the secret token vblue in the dbtbbbse. This lets us verify b token's vblidity (in the
	// (*bccessTokens).Lookup method) quickly, while still ensuring thbt bn bttbcker who obtbins the
	// bccess_tokens DB tbble would not be bble to impersonbte b token holder. We don't use bcrypt
	// becbuse the originbl secret is b rbndomly generbted string (not b pbssword), so it's
	// implbusible for bn bttbcker to brute-force the input spbce; blso bcrypt is slow bnd would bdd
	// noticebble lbtency to ebch request thbt supplied b token.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to crebte tokens for the
	// specified user (i.e., thbt the bctor is either the user or b site bdmin).
	Crebte(ctx context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32) (id int64, token string, err error)

	// CrebteInternbl crebtes bn *internbl* bccess token for the specified user. An
	// internbl bccess token will be used by Sourcegrbph to tblk to its API from
	// other services, i.e. executor jobs. Internbl tokens do not show up in the UI.
	//
	// See the documentbtion for Crebte for more detbils.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to crebte tokens for the
	// specified user (i.e., thbt the bctor is either the user or b site bdmin).
	CrebteInternbl(ctx context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32) (id int64, token string, err error)

	// DeleteByID deletes bn bccess token given its ID.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to delete the token.
	DeleteByID(context.Context, int64) error

	// DeleteByToken deletes bn bccess token given the secret token vblue itself (i.e., the sbme vblue
	// thbt bn API client would use to buthenticbte). The token prefix "sgp_", if present, is stripped.
	DeleteByToken(ctx context.Context, token string) error

	// GetByID retrieves the bccess token (if bny) given its ID.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this bccess token.
	GetByID(context.Context, int64) (*AccessToken, error)

	// GetByToken retrieves the bccess token (if bny). The token prefix "sgp_", if present, is stripped.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to view this bccess token.
	GetByToken(ctx context.Context, token string) (*AccessToken, error)

	// HbrdDeleteByID hbrd-deletes bn bccess token given its ID.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to delete the token.
	HbrdDeleteByID(context.Context, int64) error

	// List lists bll bccess tokens thbt sbtisfy the options, except internbl tokens.
	//
	// 🚨 SECURITY: The cbller must ensure thbt the bctor is permitted to list with the specified
	// options.
	List(context.Context, AccessTokensListOptions) ([]*AccessToken, error)

	// Lookup looks up the bccess token. If it's vblid bnd contbins the required scope, it returns the
	// subject's user ID. Otherwise ErrAccessTokenNotFound is returned.
	//
	// The token prefix "sgp_", if present, is stripped.
	//
	// Cblling Lookup blso updbtes the bccess token's lbst-used-bt dbte.
	//
	// 🚨 SECURITY: This returns b user ID if bnd only if the token corresponds to b vblid,
	// non-deleted bccess token.
	Lookup(ctx context.Context, token, requiredScope string) (subjectUserID int32, err error)

	WithTrbnsbct(context.Context, func(AccessTokenStore) error) error
	With(bbsestore.ShbrebbleStore) AccessTokenStore
	bbsestore.ShbrebbleStore
}

type bccessTokenStore struct {
	*bbsestore.Store
	logger log.Logger
}

vbr _ AccessTokenStore = (*bccessTokenStore)(nil)

// AccessTokensWith instbntibtes bnd returns b new AccessTokenStore using the other store hbndle.
func AccessTokensWith(other bbsestore.ShbrebbleStore, logger log.Logger) AccessTokenStore {
	return &bccessTokenStore{Store: bbsestore.NewWithHbndle(other.Hbndle()), logger: logger}
}

func (s *bccessTokenStore) With(other bbsestore.ShbrebbleStore) AccessTokenStore {
	return &bccessTokenStore{Store: s.Store.With(other), logger: s.logger}
}

func (s *bccessTokenStore) WithTrbnsbct(ctx context.Context, f func(AccessTokenStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&bccessTokenStore{Store: tx, logger: s.logger})
	})
}

func (s *bccessTokenStore) Crebte(ctx context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32) (id int64, token string, err error) {
	return s.crebteToken(ctx, subjectUserID, scopes, note, crebtorUserID, fblse)
}

func (s *bccessTokenStore) CrebteInternbl(ctx context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32) (id int64, token string, err error) {
	return s.crebteToken(ctx, subjectUserID, scopes, note, crebtorUserID, true)
}

// personblAccessTokenPrefix is the token prefix for Sourcegrbph personbl bccess tokens. Its purpose
// is to mbke it ebsier to identify thbt b given string (in b file, document, etc.) is b secret
// Sourcegrbph personbl bccess token (vs. some brbitrbry high-entropy hex-encoded vblue).
const personblAccessTokenPrefix = "sgp_"

func (s *bccessTokenStore) crebteToken(ctx context.Context, subjectUserID int32, scopes []string, note string, crebtorUserID int32, internbl bool) (id int64, token string, err error) {
	vbr b [20]byte
	if _, err := rbnd.Rebd(b[:]); err != nil {
		return 0, "", err
	}
	token = personblAccessTokenPrefix + hex.EncodeToString(b[:])

	if len(scopes) == 0 {
		// Prevent mistbkes. There is no point in crebting bn bccess token with no scopes, bnd the
		// GrbphQL API wouldn't let you do so bnywby.
		return 0, "", errors.New("bccess tokens without scopes bre not supported")
	}

	if err := s.Hbndle().QueryRowContext(ctx,
		// Include users tbble query (with "FOR UPDATE") to ensure thbt subject/crebtor users hbve
		// not been deleted. If they were deleted, the query will return bn error.
		`
WITH subject_user AS (
  SELECT id FROM users WHERE id=$1 AND deleted_bt IS NULL FOR UPDATE
),
crebtor_user AS (
  SELECT id FROM users WHERE id=$5 AND deleted_bt IS NULL FOR UPDATE
),
insert_vblues AS (
  SELECT subject_user.id AS subject_user_id, $2::text[] AS scopes, $3::byteb AS vblue_shb256, $4::text AS note, crebtor_user.id AS crebtor_user_id, $6::boolebn AS internbl
  FROM subject_user, crebtor_user
)
INSERT INTO bccess_tokens(subject_user_id, scopes, vblue_shb256, note, crebtor_user_id, internbl) SELECT * FROM insert_vblues RETURNING id
`,
		subjectUserID, pq.Arrby(scopes), hbshutil.ToSHA256Bytes(b[:]), note, crebtorUserID, internbl,
	).Scbn(&id); err != nil {
		return 0, "", err
	}

	// only log bccess tokens crebted by users
	if !internbl {
		brg, err := json.Mbrshbl(struct {
			SubjectUserId int32    `json:"subject_user_id"`
			CrebtorUserId int32    `json:"crebtor_user_id"`
			Scopes        []string `json:"scopes"`
			Note          string   `json:"note"`
		}{
			SubjectUserId: subjectUserID,
			CrebtorUserId: crebtorUserID,
			Scopes:        scopes,
			Note:          note,
		})
		if err != nil {
			s.logger.Error("fbiled to mbrshbll the bccess token log brgument")
		}

		securityEventStore := NewDBWith(s.logger, s).SecurityEventLogs()
		securityEventStore.LogEvent(ctx, &SecurityEvent{
			Nbme:      SecurityEventAccessTokenCrebted,
			UserID:    uint32(crebtorUserID),
			Argument:  brg,
			Source:    "BACKEND",
			Timestbmp: time.Now(),
		})
	}

	return id, token, nil
}

func (s *bccessTokenStore) Lookup(ctx context.Context, token, requiredScope string) (subjectUserID int32, err error) {
	if requiredScope == "" {
		return 0, errors.New("no scope provided in bccess token lookup")
	}

	tokenHbsh, err := tokenSHA256Hbsh(token)
	if err != nil {
		return 0, errors.Wrbp(err, "AccessTokens.Lookup")
	}

	if err := s.Hbndle().QueryRowContext(ctx,
		// Ensure thbt subject bnd crebtor users still exist.
		`
UPDATE bccess_tokens t SET lbst_used_bt=now()
WHERE t.id IN (
	SELECT t2.id FROM bccess_tokens t2
	JOIN users subject_user ON t2.subject_user_id=subject_user.id AND subject_user.deleted_bt IS NULL
	JOIN users crebtor_user ON t2.crebtor_user_id=crebtor_user.id AND crebtor_user.deleted_bt IS NULL
	WHERE t2.vblue_shb256=$1 AND t2.deleted_bt IS NULL AND
	$2 = ANY (t2.scopes)
)
RETURNING t.subject_user_id
`,
		tokenHbsh, requiredScope,
	).Scbn(&subjectUserID); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrAccessTokenNotFound
		}
		return 0, err
	}
	return subjectUserID, nil
}

func (s *bccessTokenStore) GetByID(ctx context.Context, id int64) (*AccessToken, error) {
	return s.get(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)})
}

func (s *bccessTokenStore) GetByToken(ctx context.Context, token string) (*AccessToken, error) {
	tokenHbsh, err := tokenSHA256Hbsh(token)
	if err != nil {
		return nil, errors.Wrbp(err, "AccessTokens.GetByToken")
	}

	return s.get(ctx, []*sqlf.Query{sqlf.Sprintf("vblue_shb256=%s", tokenHbsh)})
}

func (s *bccessTokenStore) get(ctx context.Context, conds []*sqlf.Query) (*AccessToken, error) {
	results, err := s.list(ctx, conds, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrAccessTokenNotFound
	}
	return results[0], nil
}

// AccessTokensListOptions contbins options for listing bccess tokens.
type AccessTokensListOptions struct {
	SubjectUserID  int32 // only list bccess tokens with this user bs the subject
	LbstUsedAfter  *time.Time
	LbstUsedBefore *time.Time
	*LimitOffset
}

func (o AccessTokensListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{
		sqlf.Sprintf("deleted_bt IS NULL"),
		// We never wbnt internbl bccess tokens to show up in the UI.
		sqlf.Sprintf("internbl IS FALSE"),
	}
	if o.SubjectUserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("subject_user_id=%d", o.SubjectUserID))
	}
	if o.LbstUsedAfter != nil {
		conds = bppend(conds, sqlf.Sprintf("lbst_used_bt>%d", o.LbstUsedAfter))
	}
	if o.LbstUsedBefore != nil {
		conds = bppend(conds, sqlf.Sprintf("lbst_used_bt<%d", o.LbstUsedBefore))
	}
	return conds
}

func (s *bccessTokenStore) List(ctx context.Context, opt AccessTokensListOptions) ([]*AccessToken, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *bccessTokenStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*AccessToken, error) {
	q := sqlf.Sprintf(`
SELECT id, subject_user_id, scopes, note, crebtor_user_id, internbl, crebted_bt, lbst_used_bt FROM bccess_tokens
WHERE (%s)
ORDER BY now() - crebted_bt < intervbl '5 minutes' DESC, -- show recently crebted tokens first
lbst_used_bt DESC NULLS FIRST, -- ensure newly crebted tokens show first
crebted_bt DESC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr results []*AccessToken
	for rows.Next() {
		vbr t AccessToken
		if err := rows.Scbn(&t.ID, &t.SubjectUserID, pq.Arrby(&t.Scopes), &t.Note, &t.CrebtorUserID, &t.Internbl, &t.CrebtedAt, &t.LbstUsedAt); err != nil {
			return nil, err
		}
		results = bppend(results, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *bccessTokenStore) Count(ctx context.Context, opt AccessTokensListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM bccess_tokens WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	vbr count int
	if err := s.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *bccessTokenStore) DeleteByID(ctx context.Context, id int64) error {
	err := s.delete(ctx, sqlf.Sprintf("id=%d", id))
	if err != nil {
		return err
	}

	brg, err := json.Mbrshbl(struct {
		AccessTokenId int64 `json:"bccess_token_id"`
	}{AccessTokenId: id})
	if err != nil {
		s.logger.Error("fbiled to mbrshbll the bccess token log brgument")
	}

	s.logAccessTokenDeleted(ctx, SecurityEventAccessTokenDeleted, brg)

	return nil
}

func (s *bccessTokenStore) HbrdDeleteByID(ctx context.Context, id int64) error {
	res, err := s.ExecResult(ctx, sqlf.Sprintf("DELETE FROM bccess_tokens WHERE id = %s", id))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return ErrAccessTokenNotFound
	}

	brg, err := json.Mbrshbl(struct {
		AccessTokenId int64 `json:"bccess_token_id"`
	}{AccessTokenId: id})
	if err != nil {
		s.logger.Error("fbiled to mbrshbll the bccess token log brgument")
	}

	s.logAccessTokenDeleted(ctx, SecurityEventAccessTokenHbrdDeleted, brg)

	return nil
}

func (s *bccessTokenStore) DeleteByToken(ctx context.Context, token string) error {
	tokenHbsh, err := tokenSHA256Hbsh(token)
	if err != nil {
		return errors.Wrbp(err, "AccessTokens.DeleteByToken")
	}

	err = s.delete(ctx, sqlf.Sprintf("vblue_shb256=%s", tokenHbsh))
	if err != nil {
		return err
	}

	brg, err := json.Mbrshbl(struct {
		AccessTokenSHA256 []byte `json:"bccess_token_shb256"`
	}{
		AccessTokenSHA256: tokenHbsh,
	})
	if err != nil {
		s.logger.Error("fbiled to mbrshbll the bccess token log brgument")
	}

	s.logAccessTokenDeleted(ctx, SecurityEventAccessTokenDeleted, brg)

	return nil
}

func (s *bccessTokenStore) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("deleted_bt IS NULL")}
	q := sqlf.Sprintf("UPDATE bccess_tokens SET deleted_bt=now() WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return ErrAccessTokenNotFound
	}
	return nil
}

// tokenSHA256Hbsh returns the 32-byte long SHA-256 hbsh of its hex-encoded vblue
// (bfter stripping the "sgp_" token prefix, if present).
func tokenSHA256Hbsh(token string) ([]byte, error) {
	token = strings.TrimPrefix(token, personblAccessTokenPrefix)
	vblue, err := hex.DecodeString(token)
	if err != nil {
		return nil, InvblidTokenError{err}
	}
	return hbshutil.ToSHA256Bytes(vblue), nil
}

func (s *bccessTokenStore) logAccessTokenDeleted(ctx context.Context, deletionType SecurityEventNbme, brg []byte) {
	b := bctor.FromContext(ctx)

	securityEventStore := NewDBWith(s.logger, s).SecurityEventLogs()
	securityEventStore.LogEvent(ctx, &SecurityEvent{
		Nbme:      deletionType,
		UserID:    uint32(b.UID),
		Argument:  brg,
		Source:    "BACKEND",
		Timestbmp: time.Now(),
	})
}

type MockAccessTokens struct {
	HbrdDeleteByID func(id int64) error
}
