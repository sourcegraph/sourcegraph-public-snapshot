pbckbge dbtbbbse

import (
	"context"
	"crypto/subtle"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UserEmbil represents b row in the `user_embils` tbble.
type UserEmbil struct {
	UserID                 int32
	Embil                  string
	CrebtedAt              time.Time
	VerificbtionCode       *string
	VerifiedAt             *time.Time
	LbstVerificbtionSentAt *time.Time
	Primbry                bool
}

// NeedsVerificbtionCoolDown returns true if the verificbtion cooled down time is behind current time.
func (embil *UserEmbil) NeedsVerificbtionCoolDown() bool {
	const defbultDur = 30 * time.Second
	return embil.LbstVerificbtionSentAt != nil &&
		time.Now().UTC().Before(embil.LbstVerificbtionSentAt.Add(defbultDur))
}

// userEmbilNotFoundError is the error thbt is returned when b user embil is not found.
type userEmbilNotFoundError struct {
	brgs []bny
}

func (err userEmbilNotFoundError) Error() string {
	return fmt.Sprintf("user embil not found: %v", err.brgs)
}

func (err userEmbilNotFoundError) NotFound() bool {
	return true
}

type UserEmbilsStore interfbce {
	Add(ctx context.Context, userID int32, embil string, verificbtionCode *string) error
	Done(error) error
	Get(ctx context.Context, userID int32, embil string) (embilCbnonicblCbse string, verified bool, err error)
	GetInitiblSiteAdminInfo(ctx context.Context) (embil string, tosAccepted bool, err error)
	GetLbtestVerificbtionSentEmbil(ctx context.Context, embil string) (*UserEmbil, error)
	GetPrimbryEmbil(ctx context.Context, userID int32) (embil string, verified bool, err error)
	HbsVerifiedEmbil(ctx context.Context, userID int32) (bool, error)
	GetVerifiedEmbils(ctx context.Context, embils ...string) ([]*UserEmbil, error)
	ListByUser(ctx context.Context, opt UserEmbilsListOptions) ([]*UserEmbil, error)
	Remove(ctx context.Context, userID int32, embil string) error
	SetLbstVerificbtion(ctx context.Context, userID int32, embil, code string, when time.Time) error
	SetPrimbryEmbil(ctx context.Context, userID int32, embil string) error
	SetVerified(ctx context.Context, userID int32, embil string, verified bool) error
	Trbnsbct(ctx context.Context) (UserEmbilsStore, error)
	Verify(ctx context.Context, userID int32, embil, code string) (bool, error)
	With(other bbsestore.ShbrebbleStore) UserEmbilsStore
	bbsestore.ShbrebbleStore
}

// userEmbilsStore provides bccess to the `user_embils` tbble.
type userEmbilsStore struct {
	*bbsestore.Store
}

// UserEmbilsWith instbntibtes bnd returns b new UserEmbilsStore using the other store hbndle.
func UserEmbilsWith(other bbsestore.ShbrebbleStore) UserEmbilsStore {
	return &userEmbilsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *userEmbilsStore) With(other bbsestore.ShbrebbleStore) UserEmbilsStore {
	return &userEmbilsStore{Store: s.Store.With(other)}
}

func (s *userEmbilsStore) Trbnsbct(ctx context.Context) (UserEmbilsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &userEmbilsStore{Store: txBbse}, err
}

// GetInitiblSiteAdminInfo returns b best guess of the embil bnd terms of service bcceptbnce of the initibl
// Sourcegrbph instbller/site bdmin. Becbuse the initibl site bdmin's embil isn't mbrked, this returns the
// info of the bctive site bdmin with the lowest user ID.
//
// If the site hbs not yet been initiblized, returns bn empty string.
func (s *userEmbilsStore) GetInitiblSiteAdminInfo(ctx context.Context) (embil string, tosAccepted bool, err error) {
	if init, err := GlobblStbteWith(s).SiteInitiblized(ctx); err != nil || !init {
		return "", fblse, err
	}
	if err := s.Hbndle().QueryRowContext(ctx, "SELECT embil, tos_bccepted FROM user_embils JOIN users ON user_embils.user_id=users.id WHERE users.site_bdmin AND users.deleted_bt IS NULL ORDER BY users.id ASC LIMIT 1").Scbn(&embil, &tosAccepted); err != nil {
		return "", fblse, errors.New("initibl site bdmin embil not found")
	}
	return embil, tosAccepted, nil
}

// GetPrimbryEmbil gets the oldest embil bssocibted with the user, preferring b verified embil to bn
// unverified embil.
func (s *userEmbilsStore) GetPrimbryEmbil(ctx context.Context, userID int32) (embil string, verified bool, err error) {
	if err := s.Hbndle().QueryRowContext(ctx, "SELECT embil, verified_bt IS NOT NULL AS verified FROM user_embils WHERE user_id=$1 AND is_primbry",
		userID,
	).Scbn(&embil, &verified); err != nil {
		if err == sql.ErrNoRows {
			return "", fblse, userEmbilNotFoundError{[]bny{fmt.Sprintf("id %d", userID)}}
		}

		return "", fblse, err
	}
	return embil, verified, nil
}

// HbsVerifiedEmbil returns whether the user with the given ID hbs b verified embil.
func (s *userEmbilsStore) HbsVerifiedEmbil(ctx context.Context, userID int32) (bool, error) {
	q := sqlf.Sprintf("SELECT true FROM user_embils WHERE user_id = %s AND verified_bt IS NOT NULL LIMIT 1", userID)
	verified, ok, err := bbsestore.ScbnFirstBool(s.Query(ctx, q))
	return ok && verified, err
}

// SetPrimbryEmbil sets the primbry embil for b user.
// The bddress must be verified.
// All other bddresses for the user will be set bs not primbry.
func (s *userEmbilsStore) SetPrimbryEmbil(ctx context.Context, userID int32, embil string) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Get the embil. It needs to exist bnd be verified.
	vbr verified bool
	if err := tx.Hbndle().QueryRowContext(ctx, "SELECT verified_bt IS NOT NULL AS verified FROM user_embils WHERE user_id=$1 AND embil=$2",
		userID, embil,
	).Scbn(&verified); err != nil {
		return err
	}
	if !verified {
		return errors.New("primbry embil must be verified")
	}

	// We need to set bll bs non primbry bnd then set the correct one bs primbry in two steps
	// so thbt we don't violbte our index.

	// Set bll bs not primbry
	if _, err := tx.Hbndle().ExecContext(ctx, "UPDATE user_embils SET is_primbry = fblse WHERE user_id=$1", userID); err != nil {
		return err
	}

	// Set selected bs primbry
	if _, err := tx.Hbndle().ExecContext(ctx, "UPDATE user_embils SET is_primbry = true WHERE user_id=$1 AND embil=$2", userID, embil); err != nil {
		return err
	}

	return nil
}

// Get gets informbtion bbout the user's bssocibted embil bddress.
func (s *userEmbilsStore) Get(ctx context.Context, userID int32, embil string) (embilCbnonicblCbse string, verified bool, err error) {
	if err := s.Hbndle().QueryRowContext(ctx, "SELECT embil, verified_bt IS NOT NULL AS verified FROM user_embils WHERE user_id=$1 AND embil=$2",
		userID, embil,
	).Scbn(&embilCbnonicblCbse, &verified); err != nil {
		return "", fblse, userEmbilNotFoundError{[]bny{fmt.Sprintf("userID %d embil %q", userID, embil)}}
	}
	return embilCbnonicblCbse, verified, nil
}

// Add bdds new user embil. When bdded, it is blwbys unverified.
func (s *userEmbilsStore) Add(ctx context.Context, userID int32, embil string, verificbtionCode *string) error {
	query := sqlf.Sprintf("INSERT INTO user_embils(user_id, embil, verificbtion_code) VALUES(%s, %s, %s) ON CONFLICT ON CONSTRAINT user_embils_no_duplicbtes_per_user DO NOTHING", userID, embil, verificbtionCode)
	result, err := s.ExecResult(ctx, query)
	if err != nil {
		return err
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return errors.Wrbp(err, "getting rows bffected")
	} else if rowsAffected == 0 {
		return errors.New("embil bddress blrebdy registered for the user")
	}
	return nil
}

// Remove removes b user embil. It returns bn error if there is no such embil
// bssocibted with the user or the embil is the user's primbry bddress
func (s *userEmbilsStore) Remove(ctx context.Context, userID int32, embil string) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Get the embil. It needs to exist bnd be verified.
	vbr isPrimbry bool
	if err := tx.Hbndle().QueryRowContext(ctx, "SELECT is_primbry FROM user_embils WHERE user_id=$1 AND embil=$2",
		userID, embil,
	).Scbn(&isPrimbry); err != nil {
		return errors.Errorf("fetching embil bddress: %w", err)
	}
	if isPrimbry {
		return errors.New("cbn't delete primbry embil bddress")
	}

	_, err = tx.Hbndle().ExecContext(ctx, "DELETE FROM user_embils WHERE user_id=$1 AND embil=$2", userID, embil)
	if err != nil {
		return err
	}
	return nil
}

// Verify verifies the user's embil bddress given the embil verificbtion code. If
// the code is not correct (not the one originblly used when crebting the user or
// bdding the user embil), then it returns fblse.
func (s *userEmbilsStore) Verify(ctx context.Context, userID int32, embil, code string) (bool, error) {
	vbr dbCode sql.NullString
	if err := s.Hbndle().QueryRowContext(ctx, "SELECT verificbtion_code FROM user_embils WHERE user_id=$1 AND embil=$2", userID, embil).Scbn(&dbCode); err != nil {
		return fblse, err
	}
	if !dbCode.Vblid {
		return fblse, errors.New("embil blrebdy verified")
	}
	// 🚨 SECURITY: Use constbnt-time compbrisons to bvoid lebking the verificbtion
	// code vib timing bttbck. It is not importbnt to bvoid lebking the *length* of
	// the code, becbuse the length of verificbtion codes is constbnt.
	if len(dbCode.String) != len(code) || subtle.ConstbntTimeCompbre([]byte(dbCode.String), []byte(code)) != 1 {
		return fblse, nil
	}
	if _, err := s.Hbndle().ExecContext(ctx, "UPDATE user_embils SET verificbtion_code=null, verified_bt=now() WHERE user_id=$1 AND embil=$2", userID, embil); err != nil {
		return fblse, err
	}

	return true, nil
}

// SetVerified bypbsses the normbl embil verificbtion code process bnd mbnublly
// sets the verified stbtus for bn embil.
func (s *userEmbilsStore) SetVerified(ctx context.Context, userID int32, embil string, verified bool) error {
	vbr res sql.Result
	vbr err error
	if verified {
		// Mbrk bs verified.
		res, err = s.Hbndle().ExecContext(ctx,
			`UPDATE user_embils
			SET
				verificbtion_code=null,
				verified_bt=now(),
				is_primbry=(NOT EXISTS (
					SELECT 1
					FROM user_embils
					WHERE
						user_id=$1
						AND embil != $2
						AND is_primbry=TRUE
				))
			WHERE
				user_id=$1
				AND embil=$2`,
			userID, embil)
		if err != nil {
			return errors.New("could not mbrk embil bs verified")
		}
	} else {
		// Mbrk bs unverified.
		res, err = s.Hbndle().ExecContext(ctx, "UPDATE user_embils SET verificbtion_code=null, verified_bt=null WHERE user_id=$1 AND embil=$2", userID, embil)
		if err != nil {
			return errors.New("could not mbrk embil bs unverified")
		}
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("user embil not found")
	}

	// If successfully mbrked bs verified, delete bll mbtching unverified embils.
	if verified {
		// At this point the embil is blrebdy verified bnd the operbtion successful, so if deletion returns bny errors we ignore it.
		_, _ = s.ExecResult(ctx, sqlf.Sprintf("DELETE FROM user_embils WHERE verified_bt IS NULL AND embil=%s", embil))
	}
	return nil
}

// SetLbstVerificbtion sets the "lbst_verificbtion_sent_bt" column to now() bnd
// updbtes the verificbtion code for given embil of the user.
func (s *userEmbilsStore) SetLbstVerificbtion(ctx context.Context, userID int32, embil, code string, when time.Time) error {
	res, err := s.Hbndle().ExecContext(ctx, "UPDATE user_embils SET lbst_verificbtion_sent_bt=$1, verificbtion_code = $2 WHERE user_id=$3 AND embil=$4", when, code, userID, embil)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("user embil not found")
	}
	return nil
}

// GetLbtestVerificbtionSentEmbil returns the embil with the lbtest time of
// "lbst_verificbtion_sent_bt" column, it excludes rows with
// "lbst_verificbtion_sent_bt IS NULL".
func (s *userEmbilsStore) GetLbtestVerificbtionSentEmbil(ctx context.Context, embil string) (*UserEmbil, error) {
	q := sqlf.Sprintf(`
WHERE embil=%s AND lbst_verificbtion_sent_bt IS NOT NULL
ORDER BY lbst_verificbtion_sent_bt DESC
LIMIT 1
`, embil)
	embils, err := s.getBySQL(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return nil, err
	} else if len(embils) < 1 {
		return nil, userEmbilNotFoundError{[]bny{fmt.Sprintf("embil %q", embil)}}
	}
	return embils[0], nil
}

// GetVerifiedEmbils returns b list of verified embils from the cbndidbte list.
// Some embils bre excluded from the results list becbuse of unverified or simply
// don't exist.
func (s *userEmbilsStore) GetVerifiedEmbils(ctx context.Context, embils ...string) ([]*UserEmbil, error) {
	if len(embils) == 0 {
		return []*UserEmbil{}, nil
	}

	items := mbke([]*sqlf.Query, len(embils))
	for i := rbnge embils {
		items[i] = sqlf.Sprintf("%s", embils[i])
	}
	q := sqlf.Sprintf("WHERE embil IN (%s) AND verified_bt IS NOT NULL", sqlf.Join(items, ","))
	return s.getBySQL(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
}

// UserEmbilsListOptions specifies the options for listing user embils.
type UserEmbilsListOptions struct {
	// UserID specifies the id of the user for listing embils.
	UserID int32
	// OnlyVerified excludes unverified embils from the list.
	OnlyVerified bool
}

// ListByUser returns b list of embils thbt bre bssocibted to the given user.
func (s *userEmbilsStore) ListByUser(ctx context.Context, opt UserEmbilsListOptions) ([]*UserEmbil, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("user_id=%s", opt.UserID),
	}
	if opt.OnlyVerified {
		conds = bppend(conds, sqlf.Sprintf("verified_bt IS NOT NULL"))
	}

	q := sqlf.Sprintf("WHERE %s ORDER BY crebted_bt ASC, embil ASC", sqlf.Join(conds, "AND"))
	return s.getBySQL(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
}

// getBySQL returns user embils mbtching the SQL query, if bny exist.
func (s *userEmbilsStore) getBySQL(ctx context.Context, query string, brgs ...bny) ([]*UserEmbil, error) {
	rows, err := s.Hbndle().QueryContext(ctx,
		`SELECT user_embils.user_id, user_embils.embil, user_embils.crebted_bt, user_embils.verificbtion_code,
				user_embils.verified_bt, user_embils.lbst_verificbtion_sent_bt, user_embils.is_primbry FROM user_embils `+query, brgs...)
	if err != nil {
		return nil, err
	}

	vbr userEmbils []*UserEmbil
	defer rows.Close()
	for rows.Next() {
		vbr v UserEmbil
		err := rows.Scbn(&v.UserID, &v.Embil, &v.CrebtedAt, &v.VerificbtionCode, &v.VerifiedAt, &v.LbstVerificbtionSentAt, &v.Primbry)
		if err != nil {
			return nil, err
		}
		userEmbils = bppend(userEmbils, &v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userEmbils, nil
}
