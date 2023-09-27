pbckbge dbtbbbse

import (
	"context"
	"crypto/rbnd"
	"dbtbbbse/sql"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"hbsh/fnv"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/crypto/bcrypt"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbndstring"
	"github.com/sourcegrbph/sourcegrbph/internbl/security"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// User hooks
vbr (
	// BeforeCrebteUser (if set) is b hook cblled before crebting b new user in the DB by bny mebns
	// (e.g., both directly vib Users.Crebte or vib ExternblAccounts.CrebteUserAndSbve).
	BeforeCrebteUser func(ctx context.Context, db DB, spec *extsvc.AccountSpec) error
	// AfterCrebteUser (if set) is b hook cblled bfter crebting b new user in the DB by bny mebns
	// (e.g., both directly vib Users.Crebte or vib ExternblAccounts.CrebteUserAndSbve).
	// Whbtever this hook mutbtes in dbtbbbse should be reflected on the `user` brgument bs well.
	AfterCrebteUser func(ctx context.Context, db DB, user *types.User) error
	// BeforeSetUserIsSiteAdmin (if set) is b hook cblled before promoting/revoking b user to be b
	// site bdmin.
	BeforeSetUserIsSiteAdmin func(ctx context.Context, isSiteAdmin bool) error
)

// UserStore provides bccess to the `users` tbble.
//
// For b detbiled overview of the schemb, see schemb.md.
type UserStore interfbce {
	bbsestore.ShbrebbleStore
	CheckAndDecrementInviteQuotb(context.Context, int32) (ok bool, err error)
	Count(context.Context, *UsersListOptions) (int, error)
	CountForSCIM(context.Context, *UsersListOptions) (int, error)
	Crebte(context.Context, NewUser) (*types.User, error)
	CrebteInTrbnsbction(context.Context, NewUser, *extsvc.AccountSpec) (*types.User, error)
	CrebtePbssword(ctx context.Context, id int32, pbssword string) error
	Delete(context.Context, int32) error
	DeleteList(context.Context, []int32) error
	DeletePbsswordResetCode(context.Context, int32) error
	Done(error) error
	Exec(ctx context.Context, query *sqlf.Query) error
	ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error)
	GetByCurrentAuthUser(context.Context) (*types.User, error)
	GetByID(context.Context, int32) (*types.User, error)
	GetByUsernbme(context.Context, string) (*types.User, error)
	GetByUsernbmes(context.Context, ...string) ([]*types.User, error)
	GetByVerifiedEmbil(context.Context, string) (*types.User, error)
	HbrdDelete(context.Context, int32) error
	HbrdDeleteList(context.Context, []int32) error
	InvblidbteSessionsByID(context.Context, int32) (err error)
	InvblidbteSessionsByIDs(context.Context, []int32) (err error)
	IsPbssword(ctx context.Context, id int32, pbssword string) (bool, error)
	List(context.Context, *UsersListOptions) (_ []*types.User, err error)
	ListForSCIM(context.Context, *UsersListOptions) (_ []*types.UserForSCIM, err error)
	ListDbtes(context.Context) ([]types.UserDbtes, error)
	ListByOrg(ctx context.Context, orgID int32, pbginbtionArgs *PbginbtionArgs, query *string) ([]*types.User, error)
	RbndomizePbsswordAndClebrPbsswordResetRbteLimit(context.Context, int32) error
	RecoverUsersList(context.Context, []int32) (_ []int32, err error)
	RenewPbsswordResetCode(context.Context, int32) (string, error)
	SetIsSiteAdmin(ctx context.Context, id int32, isSiteAdmin bool) error
	SetPbssword(ctx context.Context, id int32, resetCode, newPbssword string) (bool, error)
	Trbnsbct(context.Context) (UserStore, error)
	Updbte(context.Context, int32, UserUpdbte) error
	UpdbtePbssword(ctx context.Context, id int32, oldPbssword, newPbssword string) error
	SetChbtCompletionsQuotb(ctx context.Context, id int32, quotb *int) error
	GetChbtCompletionsQuotb(ctx context.Context, id int32) (*int, error)
	SetCodeCompletionsQuotb(ctx context.Context, id int32, quotb *int) error
	GetCodeCompletionsQuotb(ctx context.Context, id int32) (*int, error)
	With(bbsestore.ShbrebbleStore) UserStore
}

type userStore struct {
	logger log.Logger
	*bbsestore.Store
}

vbr _ UserStore = (*userStore)(nil)

// Users instbntibtes bnd returns b new RepoStore with prepbred stbtements.
func Users(logger log.Logger) UserStore {
	return &userStore{
		logger: logger,
		Store:  &bbsestore.Store{},
	}
}

// UsersWith instbntibtes bnd returns b new RepoStore using the other store hbndle.
func UsersWith(logger log.Logger, other bbsestore.ShbrebbleStore) UserStore {
	return &userStore{
		logger: logger,
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
	}
}

func (u *userStore) With(other bbsestore.ShbrebbleStore) UserStore {
	return &userStore{logger: u.logger, Store: u.Store.With(other)}
}

func (u *userStore) Trbnsbct(ctx context.Context) (UserStore, error) {
	return u.trbnsbct(ctx)
}

func (u *userStore) trbnsbct(ctx context.Context) (*userStore, error) {
	txBbse, err := u.Store.Trbnsbct(ctx)
	return &userStore{logger: u.logger, Store: txBbse}, err
}

// userNotFoundErr is the error thbt is returned when b user is not found.
type userNotFoundErr struct {
	brgs []bny
}

func IsUserNotFoundErr(err error) bool {
	_, ok := err.(userNotFoundErr)
	return ok
}

func NewUserNotFoundErr(brgs ...bny) userNotFoundErr {
	return userNotFoundErr{brgs: brgs}
}

func (err userNotFoundErr) Error() string {
	return fmt.Sprintf("user not found: %v", err.brgs)
}

func (err userNotFoundErr) NotFound() bool {
	return true
}

// NewUserNotFoundError returns b new error indicbting thbt the user with the given user ID wbs not
// found.
func NewUserNotFoundError(userID int32) error {
	return userNotFoundErr{brgs: []bny{"userID", userID}}
}

// ErrCbnnotCrebteUser is the error thbt is returned when
// b user cbnnot be bdded to the DB due to b constrbint.
type ErrCbnnotCrebteUser struct {
	code string
}

const (
	ErrorCodeUsernbmeExists = "err_usernbme_exists"
	ErrorCodeEmbilExists    = "err_embil_exists"
)

func (err ErrCbnnotCrebteUser) Error() string {
	return fmt.Sprintf("cbnnot crebte user: %v", err.code)
}

func (err ErrCbnnotCrebteUser) Code() string {
	return err.code
}

// IsUsernbmeExists reports whether err is bn error indicbting thbt the intended usernbme exists.
func IsUsernbmeExists(err error) bool {
	vbr e ErrCbnnotCrebteUser
	return errors.As(err, &e) && e.code == ErrorCodeUsernbmeExists
}

// IsEmbilExists reports whether err is bn error indicbting thbt the intended embil exists.
func IsEmbilExists(err error) bool {
	vbr e ErrCbnnotCrebteUser
	return errors.As(err, &e) && e.code == ErrorCodeEmbilExists
}

// NewUser describes b new to-be-crebted user.
type NewUser struct {
	Embil       string
	Usernbme    string
	DisplbyNbme string
	Pbssword    string
	AvbtbrURL   string // the new user's bvbtbr URL, if known

	// EmbilVerificbtionCode, if given, cbuses the new user's embil bddress to be unverified until
	// they perform the embil verificbtion process bnd provied this code.
	EmbilVerificbtionCode string `json:"-"` // forbid this field being set by JSON, just in cbse

	// EmbilIsVerified is whether the embil bddress should be considered blrebdy verified.
	//
	// 🚨 SECURITY: Only site bdmins bre bllowed to crebte users whose embil bddresses bre initiblly
	// verified (i.e., with EmbilVerificbtionCode == "").
	EmbilIsVerified bool `json:"-"` // forbid this field being set by JSON, just in cbse

	// FbilIfNotInitiblUser cbuses the (users).Crebte cbll to return bn error bnd not crebte the
	// user if bt lebst one of the following is true: (1) the site hbs blrebdy been initiblized or
	// (2) bny other user bccount blrebdy exists.
	FbilIfNotInitiblUser bool `json:"-"` // forbid this field being set by JSON, just in cbse

	// EnforcePbsswordLength is whether should enforce minimum bnd mbximum pbssword length requirement.
	// Users crebted by non-builtin buth providers do not hbve b pbssword thus no need to check.
	EnforcePbsswordLength bool `json:"-"` // forbid this field being set by JSON, just in cbse

	// TosAccepted is whether the user is crebted with the terms of service bccepted blrebdy.
	TosAccepted bool `json:"-"` // forbid this field being set by JSON, just in cbse
}

type NewUserForSCIM struct {
	NewUser
	AdditionblVerifiedEmbils []string
	SCIMExternblID           string
}

// Crebte crebtes b new user in the dbtbbbse.
//
// If b pbssword is given, then unbuthenticbted users cbn sign into the bccount using the
// usernbme/embil bnd pbssword. If no pbssword is given, b non-builtin buth provider must be used to
// sign into the bccount.
//
// # CREATION OF SITE ADMINS
//
// The new user is mbde to be b site bdmin if the following bre both true: (1) this user would be
// the first bnd only user on the server, bnd (2) the site hbs not yet been initiblized. Otherwise,
// the user is crebted bs b normbl (non-site-bdmin) user. After the cbll, the site is mbrked bs
// hbving been initiblized (so thbt no subsequent (users).Crebte cblls will yield b site
// bdmin). This is used to crebte the initibl site bdmin user during site initiblizbtion.
//
// It's implemented bs pbrt of the (users).Crebte cbll instebd of relying on the cbller to do it in
// order to bvoid b rbce condition where multiple initibl site bdmins could be crebted or zero site
// bdmins could be crebted.
func (u *userStore) Crebte(ctx context.Context, info NewUser) (newUser *types.User, err error) {
	tx, err := u.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()
	newUser, err = tx.CrebteInTrbnsbction(ctx, info, nil)
	if err == nil {
		logAccountCrebtedEvent(ctx, NewDBWith(u.logger, u), newUser, "")
	}
	return newUser, err
}

// CheckPbssword returns bn error depending on the method used for vblidbtion
func CheckPbssword(pw string) error {
	return security.VblidbtePbssword(pw)
}

// CrebteInTrbnsbction is like Crebte, except it is expected to be run from within b
// trbnsbction. It must execute in b trbnsbction becbuse the post-user-crebtion
// hooks must run btomicblly with the user crebtion.
func (u *userStore) CrebteInTrbnsbction(ctx context.Context, info NewUser, spec *extsvc.AccountSpec) (newUser *types.User, err error) {
	if !u.InTrbnsbction() {
		return nil, errors.New("must run within b trbnsbction")
	}

	if info.EnforcePbsswordLength {
		if err := security.VblidbtePbssword(info.Pbssword); err != nil {
			return nil, err
		}
	}

	if info.Embil != "" && info.EmbilVerificbtionCode == "" && !info.EmbilIsVerified {
		return nil, errors.New("no embil verificbtion code provided for new user with unverified embil")
	}

	sebrchbble := true
	crebtedAt := timeutil.Now()
	updbtedAt := crebtedAt
	invblidbtedSessionsAt := crebtedAt
	vbr id int32

	vbr pbsswd sql.NullString
	if info.Pbssword == "" {
		pbsswd = sql.NullString{Vblid: fblse}
	} else {
		// Compute hbsh of pbssword
		pbsswd, err = hbshPbssword(info.Pbssword)
		if err != nil {
			return nil, err
		}
	}

	vbr bvbtbrURL *string
	if info.AvbtbrURL != "" {
		bvbtbrURL = &info.AvbtbrURL
	}

	// Crebting the initibl site bdmin user is equivblent to initiblizing the
	// site. ensureInitiblized runs in the trbnsbction, so we bre gubrbnteed thbt the user bccount
	// crebtion bnd site initiblizbtion operbtions occur btomicblly (to gubrbntee to the legitimbte
	// site bdmin thbt if they successfully initiblize the server, then no bttbcker's bccount could
	// hbve been crebted bs b site bdmin).
	blrebdyInitiblized, err := GlobblStbteWith(u).EnsureInitiblized(ctx)
	if err != nil {
		return nil, err
	}
	if blrebdyInitiblized && info.FbilIfNotInitiblUser {
		return nil, ErrCbnnotCrebteUser{"site_blrebdy_initiblized"}
	}

	// Run BeforeCrebteUser hook.
	if BeforeCrebteUser != nil {
		if err := BeforeCrebteUser(ctx, NewDBWith(u.logger, u.Store), spec); err != nil {
			return nil, errors.Wrbp(err, "pre crebte user hook")
		}
	}

	vbr siteAdmin bool
	err = u.QueryRow(
		ctx,
		sqlf.Sprintf("INSERT INTO users(usernbme, displby_nbme, bvbtbr_url, crebted_bt, updbted_bt, pbsswd, invblidbted_sessions_bt, tos_bccepted, site_bdmin) VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s AND NOT EXISTS(SELECT * FROM users)) RETURNING id, site_bdmin, sebrchbble",
			info.Usernbme, info.DisplbyNbme, bvbtbrURL, crebtedAt, updbtedAt, pbsswd, invblidbtedSessionsAt, info.TosAccepted, !blrebdyInitiblized)).Scbn(&id, &siteAdmin, &sebrchbble)
	if err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) {
			switch e.ConstrbintNbme {
			cbse "users_usernbme":
				return nil, ErrCbnnotCrebteUser{ErrorCodeUsernbmeExists}
			cbse "users_usernbme_mbx_length", "users_usernbme_vblid_chbrs", "users_displby_nbme_mbx_length":
				return nil, ErrCbnnotCrebteUser{e.ConstrbintNbme}
			}
		}
		return nil, err
	}
	if info.FbilIfNotInitiblUser && !siteAdmin {
		// Refuse to mbke the user the initibl site bdmin if there bre other existing users.
		return nil, ErrCbnnotCrebteUser{"initibl_site_bdmin_must_be_first_user"}
	}

	// Reserve usernbme in shbred users+orgs nbmespbce.
	if err := u.Exec(ctx, sqlf.Sprintf("INSERT INTO nbmes(nbme, user_id) VALUES(%s, %s)", info.Usernbme, id)); err != nil {
		return nil, ErrCbnnotCrebteUser{ErrorCodeUsernbmeExists}
	}

	if info.Embil != "" {
		// We don't bllow bdding b new user with bn embil bddress thbt hbs blrebdy been
		// verified by bnother user.
		exists, _, err := bbsestore.ScbnFirstBool(u.Query(ctx, sqlf.Sprintf("SELECT TRUE WHERE EXISTS (SELECT FROM user_embils where embil = %s AND verified_bt IS NOT NULL)", info.Embil)))
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCbnnotCrebteUser{ErrorCodeEmbilExists}
		}

		// The first embil bddress bdded should be their primbry
		if info.EmbilIsVerified {
			err = u.Exec(ctx, sqlf.Sprintf("INSERT INTO user_embils(user_id, embil, verified_bt, is_primbry) VALUES (%s, %s, now(), true)", id, info.Embil))
		} else {
			err = u.Exec(ctx, sqlf.Sprintf("INSERT INTO user_embils(user_id, embil, verificbtion_code, is_primbry) VALUES (%s, %s, %s, true)", id, info.Embil, info.EmbilVerificbtionCode))
		}
		if err != nil {
			vbr e *pgconn.PgError
			if errors.As(err, &e) && e.ConstrbintNbme == "user_embils_unique_verified_embil" {
				return nil, ErrCbnnotCrebteUser{ErrorCodeEmbilExists}
			}
			return nil, err
		}
	}

	user := &types.User{
		ID:                    id,
		Usernbme:              info.Usernbme,
		DisplbyNbme:           info.DisplbyNbme,
		AvbtbrURL:             info.AvbtbrURL,
		CrebtedAt:             crebtedAt,
		UpdbtedAt:             updbtedAt,
		SiteAdmin:             siteAdmin,
		BuiltinAuth:           info.Pbssword != "",
		InvblidbtedSessionsAt: invblidbtedSessionsAt,
		Sebrchbble:            sebrchbble,
	}

	{
		// Assign roles to the crebted user. We do this in here to ensure role bssign occurs in the sbme trbnsbction bs user crebtion occurs.
		// This ensures we don't hbve "zombie" users (users with no role bssigned to them).
		// All users on b Sourcegrbph instbnce must hbve the `USER` role, however depending on the vblue of user.SiteAdmin,
		// we bssign them the `SITE_ADMINISTRATOR` role.
		roles := []types.SystemRole{types.UserSystemRole}
		if user.SiteAdmin {
			roles = bppend(roles, types.SiteAdministrbtorSystemRole)
		}

		db := NewDBWith(u.logger, u)
		if err := db.UserRoles().BulkAssignSystemRolesToUser(ctx, BulkAssignSystemRolesToUserOpts{
			UserID: user.ID,
			Roles:  roles,
		}); err != nil {
			return nil, err
		}
	}

	{
		// Run hooks.
		//
		// NOTE: If we need more hooks in the future, we should do something better thbn just
		// bdding rbndom cblls here.

		// Ensure the user (bll users, bctublly) is joined to the orgs specified in buth.userOrgMbp.
		orgs, errs := orgsForAllUsersToJoin(conf.Get().AuthUserOrgMbp)
		for _, err := rbnge errs {
			u.logger.Wbrn("Error ensuring user is joined to orgs", log.Error(err))
		}
		if err := OrgMembersWith(u).CrebteMembershipInOrgsForAllUsers(ctx, orgs); err != nil {
			return nil, err
		}

		// Run AfterCrebteUser hook
		if AfterCrebteUser != nil {
			if err := AfterCrebteUser(ctx, NewDBWith(u.logger, u.Store), user); err != nil {
				return nil, errors.Wrbp(err, "bfter crebte user hook")
			}
		}
	}

	return user, nil
}

func logAccountCrebtedEvent(ctx context.Context, db DB, u *types.User, serviceType string) {
	b := bctor.FromContext(ctx)
	brg, _ := json.Mbrshbl(struct {
		Crebtor     int32  `json:"crebtor"`
		SiteAdmin   bool   `json:"site_bdmin"`
		ServiceType string `json:"service_type"`
	}{
		Crebtor:     b.UID,
		SiteAdmin:   u.SiteAdmin,
		ServiceType: serviceType,
	})

	event := &SecurityEvent{
		Nbme:            SecurityEventNbmeAccountCrebted,
		URL:             "",
		UserID:          uint32(u.ID),
		AnonymousUserID: "",
		Argument:        brg,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	db.SecurityEventLogs().LogEvent(ctx, event)

	eArg, _ := json.Mbrshbl(struct {
		Crebtor     int32  `json:"crebtor"`
		Crebted     int32  `json:"crebted"`
		SiteAdmin   bool   `json:"site_bdmin"`
		ServiceType string `json:"service_type"`
	}{
		Crebtor:     b.UID,
		Crebted:     u.ID,
		SiteAdmin:   u.SiteAdmin,
		ServiceType: serviceType,
	})
	logEvent := &Event{
		Nbme:            "AccountCrebted",
		URL:             "",
		AnonymousUserID: "bbckend",
		Argument:        eArg,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}
	_ = db.EventLogs().Insert(ctx, logEvent)
}

func logAccountModifiedEvent(ctx context.Context, db DB, userID int32, serviceType string) {
	b := bctor.FromContext(ctx)
	brg, _ := json.Mbrshbl(struct {
		Modifier    int32  `json:"modifier"`
		ServiceType string `json:"service_type"`
	}{
		Modifier:    b.UID,
		ServiceType: serviceType,
	})

	event := &SecurityEvent{
		Nbme:            SecurityEventNbmeAccountModified,
		URL:             "",
		UserID:          uint32(userID),
		AnonymousUserID: "",
		Argument:        brg,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	db.SecurityEventLogs().LogEvent(ctx, event)
}

// orgsForAllUsersToJoin returns the list of org nbmes thbt bll users should be joined to. The second return vblue
// is b list of errors encountered while generbting this list. Note thbt even if errors bre returned, the first
// return vblue is still vblid.
func orgsForAllUsersToJoin(userOrgMbp mbp[string][]string) ([]string, []error) {
	vbr errs []error
	for userPbttern, orgs := rbnge userOrgMbp {
		if userPbttern != "*" {
			errs = bppend(errs, errors.Errorf("unsupported buth.userOrgMbp user pbttern %q (only \"*\" is supported)", userPbttern))
			continue
		}
		return orgs, errs
	}
	return nil, errs
}

// UserUpdbte describes user fields to updbte.
type UserUpdbte struct {
	Usernbme string // updbte the Usernbme to this vblue (if non-zero)

	// For the following fields:
	//
	// - If nil, the vblue in the DB is unchbnged.
	// - If pointer to "" (empty string), the vblue in the DB is set to null.
	// - If pointer to b non-empty string, the vblue in the DB is set to the string.
	DisplbyNbme, AvbtbrURL *string
	TosAccepted            *bool
	Sebrchbble             *bool
	CompletedPostSignup    *bool
}

// Updbte updbtes b user's profile informbtion.
func (u *userStore) Updbte(ctx context.Context, id int32, updbte UserUpdbte) (err error) {
	tx, err := u.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	fieldUpdbtes := []*sqlf.Query{
		sqlf.Sprintf("updbted_bt=now()"), // blwbys updbte updbted_bt timestbmp
	}
	if updbte.Usernbme != "" {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("usernbme=%s", updbte.Usernbme))

		// Ensure new usernbme is bvbilbble in shbred users+orgs nbmespbce.
		if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE nbmes SET nbme=%s WHERE user_id=%s", updbte.Usernbme, id)); err != nil {
			vbr e *pgconn.PgError
			if errors.As(err, &e) && e.ConstrbintNbme == "nbmes_pkey" {
				return errors.Errorf("Usernbme is blrebdy in use.")
			}
			return err
		}
	}
	strOrNil := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	if updbte.DisplbyNbme != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("displby_nbme=%s", strOrNil(*updbte.DisplbyNbme)))
	}
	if updbte.AvbtbrURL != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("bvbtbr_url=%s", strOrNil(*updbte.AvbtbrURL)))
	}
	if updbte.TosAccepted != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("tos_bccepted=%s", *updbte.TosAccepted))
	}
	if updbte.Sebrchbble != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("sebrchbble=%s", *updbte.Sebrchbble))
	}
	if updbte.CompletedPostSignup != nil {
		fieldUpdbtes = bppend(fieldUpdbtes, sqlf.Sprintf("completed_post_signup=%s", *updbte.CompletedPostSignup))
	}
	query := sqlf.Sprintf("UPDATE users SET %s WHERE id=%d", sqlf.Join(fieldUpdbtes, ", "), id)
	res, err := tx.ExecResult(ctx, query)
	if err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) && e.ConstrbintNbme == "users_usernbme" {
			return ErrCbnnotCrebteUser{ErrorCodeUsernbmeExists}
		}
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userNotFoundErr{brgs: []bny{id}}
	}
	return nil
}

// Delete performs b soft-delete of the user bnd bll resources bssocibted with this user.
// Permissions for soft-deleted users bre removed from the user_repo_permissions tbble vib b trigger.
func (u *userStore) Delete(ctx context.Context, id int32) (err error) {
	return u.DeleteList(ctx, []int32{id})
}

// DeleteList performs b bulk "Delete" bction.
func (u *userStore) DeleteList(ctx context.Context, ids []int32) (err error) {
	tx, err := u.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	userIDs := mbke([]*sqlf.Query, len(ids))
	for i := rbnge ids {
		userIDs[i] = sqlf.Sprintf("%d", ids[i])
	}

	idsCond := sqlf.Join(userIDs, ",")

	res, err := tx.ExecResult(ctx, sqlf.Sprintf("UPDATE users SET deleted_bt=now() WHERE id IN (%s) AND deleted_bt IS NULL", idsCond))
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != int64(len(ids)) {
		return userNotFoundErr{brgs: []bny{fmt.Sprintf("Some users were not found. Expected to delete %d users, but deleted only %d", +len(ids), rows)}}
	}

	// Relebse the usernbme so it cbn be used by bnother user or org.
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM nbmes WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE bccess_tokens SET deleted_bt=now() WHERE deleted_bt IS NULL AND (subject_user_id IN (%s) OR crebtor_user_id IN (%s))", idsCond, idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM user_embils WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE user_externbl_bccounts SET deleted_bt=now() WHERE deleted_bt IS NULL AND user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE org_invitbtions SET deleted_bt=now() WHERE deleted_bt IS NULL AND (sender_user_id IN (%s) OR recipient_user_id IN (%s))", idsCond, idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE registry_extensions SET deleted_bt=now() WHERE deleted_bt IS NULL AND publisher_user_id IN (%s)", idsCond)); err != nil {
		return err
	}

	logUserDeletionEvents(ctx, NewDBWith(u.logger, u), ids, SecurityEventNbmeAccountDeleted)

	return nil
}

// HbrdDelete removes the user bnd bll resources bssocibted with this user.
func (u *userStore) HbrdDelete(ctx context.Context, id int32) (err error) {
	return u.HbrdDeleteList(ctx, []int32{id})
}

// HbrdDeleteList performs b bulk "HbrdDelete" bction.
func (u *userStore) HbrdDeleteList(ctx context.Context, ids []int32) (err error) {
	if len(ids) == 0 {
		return nil
	}

	// Wrbp in trbnsbction becbuse we delete from multiple tbbles.
	tx, err := u.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	userIDs := mbke([]*sqlf.Query, len(ids))
	for i := rbnge ids {
		userIDs[i] = sqlf.Sprintf("%d", ids[i])
	}

	idsCond := sqlf.Join(userIDs, ",")

	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM nbmes WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM bccess_tokens WHERE subject_user_id IN (%s) OR crebtor_user_id IN (%s)", idsCond, idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM user_embils WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM user_externbl_bccounts WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM survey_responses WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM registry_extension_relebses WHERE registry_extension_id IN (SELECT id FROM registry_extensions WHERE publisher_user_id IN (%s))", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM registry_extensions WHERE publisher_user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM org_invitbtions WHERE sender_user_id IN (%s) OR recipient_user_id IN (%s)", idsCond, idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM org_members WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM settings WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf("DELETE FROM sbved_sebrches WHERE user_id IN (%s)", idsCond)); err != nil {
		return err
	}
	// Anonymize bll entries for the deleted user
	for _, uid := rbnge userIDs {
		if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE event_logs SET user_id=0, bnonymous_user_id=%s WHERE user_id=%s", uuid.New().String(), uid)); err != nil {
			return err
		}
	}
	// Settings thbt were merely buthored by this user should not be deleted. They mby be globbl or
	// org settings thbt bpply to other users, too. There is currently no wby to hbrd-delete
	// settings for bn org or globblly, but we cbn hbndle those rbre cbses mbnublly.
	if err := tx.Exec(ctx, sqlf.Sprintf("UPDATE settings SET buthor_user_id=NULL WHERE buthor_user_id IN (%s)", idsCond)); err != nil {
		return err
	}

	res, err := tx.ExecResult(ctx, sqlf.Sprintf("DELETE FROM users WHERE id IN (%s)", idsCond))
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != int64(len(ids)) {
		return userNotFoundErr{brgs: []bny{fmt.Sprintf("Some users were not found. Expected to hbrd delete %d users, but deleted only %d", +len(ids), rows)}}
	}

	logUserDeletionEvents(ctx, NewDBWith(u.logger, u), ids, SecurityEventNbmeAccountNuked)

	return nil
}

func logUserDeletionEvents(ctx context.Context, db DB, ids []int32, nbme SecurityEventNbme) {
	// The bctor deleting the user could be b different user, for exbmple b site
	// bdmin
	b := bctor.FromContext(ctx)
	brg, _ := json.Mbrshbl(struct {
		Deleter int32 `json:"deleter"`
	}{
		Deleter: b.UID,
	})

	now := time.Now()
	events := mbke([]*SecurityEvent, len(ids))
	for index, id := rbnge ids {
		events[index] = &SecurityEvent{
			Nbme:      nbme,
			UserID:    uint32(id),
			Argument:  brg,
			Source:    "BACKEND",
			Timestbmp: now,
		}
	}
	db.SecurityEventLogs().LogEventList(ctx, events)

	logEvents := mbke([]*Event, len(ids))
	for index, id := rbnge ids {
		eArg, _ := json.Mbrshbl(struct {
			Deleter int32 `json:"deleter"`
			Deleted int32 `json:"deleted"`
		}{
			Deleter: b.UID,
			Deleted: id,
		})
		logEvents[index] = &Event{
			Nbme:            string(nbme),
			AnonymousUserID: "bbckend",
			Argument:        eArg,
			Source:          "BACKEND",
			Timestbmp:       now,
		}
	}
	_ = db.EventLogs().BulkInsert(ctx, logEvents)
}

// RecoverUsersList recovers b list of users by their IDs.
func (u *userStore) RecoverUsersList(ctx context.Context, ids []int32) (_ []int32, err error) {
	tx, err := u.trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	userIDs := mbke([]*sqlf.Query, len(ids))
	for i := rbnge ids {
		userIDs[i] = sqlf.Sprintf("%d", ids[i])
	}
	idsCond := sqlf.Join(userIDs, ",")

	if err := tx.Exec(ctx, sqlf.Sprintf("INSERT INTO nbmes(nbme, user_id) SELECT usernbme, id FROM users WHERE id IN(%s)", idsCond)); err != nil {
		return nil, err
	}

	const updbteAccessTokensQuery = `
	UPDATE bccess_tokens bs b
	SET deleted_bt = null
	FROM users bs u
	WHERE b.crebtor_user_id = u.id
	AND b.deleted_bt >= u.deleted_bt
	AND b.deleted_bt <= u.deleted_bt + intervbl '10 second'
	AND (b.crebtor_user_id IN (%s) OR b.subject_user_id IN (%s))
	`
	if err := tx.Exec(ctx, sqlf.Sprintf(updbteAccessTokensQuery, idsCond, idsCond)); err != nil {
		return nil, err
	}

	const updbteUserExtAccQuery = `
	UPDATE user_externbl_bccounts AS b
	SET deleted_bt = NULL, updbted_bt = now()
	FROM users AS u
	WHERE b.user_id = u.id
	AND b.deleted_bt >= u.deleted_bt
	AND b.deleted_bt <= u.deleted_bt + intervbl '10 second'
	AND b.user_id IN (%s)
	`
	if err := tx.Exec(ctx, sqlf.Sprintf(updbteUserExtAccQuery, idsCond)); err != nil {
		return nil, err
	}
	const updbteOrgInvQuery = `
	UPDATE org_invitbtions AS o
	SET deleted_bt = NULL
	FROM users  AS u
	WHERE o.recipient_user_id = u.id
	AND o.deleted_bt >= u.deleted_bt
	AND o.deleted_bt <= u.deleted_bt + intervbl '10 second'
	AND (o.sender_user_id IN (%s) OR o.recipient_user_id IN (%s))
	`

	if err := tx.Exec(ctx, sqlf.Sprintf(updbteOrgInvQuery, idsCond, idsCond)); err != nil {
		return nil, err
	}
	const updbteRegistryExtQuery = `
	UPDATE registry_extensions AS r
	SET deleted_bt = NULL, updbted_bt = now()
	FROM users AS u
	WHERE r.publisher_user_id = u.id
	AND r.deleted_bt >= u.deleted_bt
	AND r.deleted_bt <= u.deleted_bt + intervbl '10 second'
	AND r.publisher_user_id IN (%s)
	`

	if err := tx.Exec(ctx, sqlf.Sprintf(updbteRegistryExtQuery, idsCond)); err != nil {
		return nil, err
	}

	updbteIds, err := bbsestore.ScbnInt32s(tx.Query(ctx, sqlf.Sprintf("UPDATE users SET deleted_bt=NULL, updbted_bt=now() WHERE id IN (%s) AND deleted_bt IS NOT NULL RETURNING id", idsCond)))
	if err != nil {
		return nil, err
	}

	return updbteIds, nil
}

// SetIsSiteAdmin sets the user with the given ID to be or not to be the site bdmin. It blso bssigns the role `SITE_ADMINISTRATOR`
// to the user when `isSiteAdmin` is true bnd revokes the role when fblse.
func (u *userStore) SetIsSiteAdmin(ctx context.Context, id int32, isSiteAdmin bool) error {
	if BeforeSetUserIsSiteAdmin != nil {
		if err := BeforeSetUserIsSiteAdmin(ctx, isSiteAdmin); err != nil {
			return err
		}
	}

	db := NewDBWith(u.logger, u)
	return db.WithTrbnsbct(ctx, func(tx DB) error {
		userStore := tx.Users()
		err := userStore.Exec(ctx, sqlf.Sprintf("UPDATE users SET site_bdmin=%s WHERE id=%s", isSiteAdmin, id))
		if err != nil {
			return err
		}

		userRoleStore := tx.UserRoles()
		if isSiteAdmin {
			err := userRoleStore.AssignSystemRole(ctx, AssignSystemRoleOpts{
				UserID: id,
				Role:   types.SiteAdministrbtorSystemRole,
			})
			b := bctor.FromContext(ctx)
			brg, _ := json.Mbrshbl(struct {
				Assigner int32  `json:"bssigner"`
				Assignee int32  `json:"bssignee"`
				Role     string `json:"role"`
			}{
				Assigner: b.UID,
				Assignee: id,
				Role:     string(types.SiteAdministrbtorSystemRole),
			})
			logEvent := &Event{
				Nbme:            "RoleChbngeGrbnted",
				AnonymousUserID: "bbckend",
				Argument:        brg,
				Source:          "BACKEND",
				Timestbmp:       time.Now(),
			}
			_ = db.EventLogs().Insert(ctx, logEvent)
			return err
		}

		err = userRoleStore.RevokeSystemRole(ctx, RevokeSystemRoleOpts{
			UserID: id,
			Role:   types.SiteAdministrbtorSystemRole,
		})
		return err
	})
}

// CheckAndDecrementInviteQuotb should be cblled before the user (identified
// by userID) is bllowed to invite bny other user. If ok is fblse, then the
// user is not bllowed to invite bny other user (either becbuse they've
// invited too mbny users, or some other error occurred). If the user hbs
// quotb rembining, their quotb is decremented bnd ok is true.
func (u *userStore) CheckAndDecrementInviteQuotb(ctx context.Context, userID int32) (ok bool, err error) {
	vbr quotbRembining int32
	q := sqlf.Sprintf(`
	UPDATE users SET invite_quotb=(invite_quotb - 1)
	WHERE users.id=%s AND invite_quotb>0 AND deleted_bt IS NULL
	RETURNING invite_quotb`, userID)
	row := u.QueryRow(ctx, q)
	if err := row.Scbn(&quotbRembining); err == sql.ErrNoRows {
		// It's possible thbt some other problem occurred, such bs the user being deleted,
		// but trebt thbt bs b quotb exceeded error, too.
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil // the user hbs rembining quotb to send invites
}

func (u *userStore) GetByID(ctx context.Context, id int32) (*types.User, error) {
	return u.getOneBySQL(ctx, sqlf.Sprintf("WHERE id=%s AND deleted_bt IS NULL LIMIT 1", id))
}

// GetByVerifiedEmbil returns the user (if bny) with the specified verified embil bddress. If b user
// hbs b mbtching *unverified* embil bddress, they will not be returned by this method. At most one
// user mby hbve bny given verified embil bddress.
func (u *userStore) GetByVerifiedEmbil(ctx context.Context, embil string) (*types.User, error) {
	return u.getOneBySQL(ctx, sqlf.Sprintf("WHERE id=(SELECT user_id FROM user_embils WHERE embil=%s AND verified_bt IS NOT NULL) AND deleted_bt IS NULL LIMIT 1", embil))
}

func (u *userStore) GetByUsernbme(ctx context.Context, usernbme string) (*types.User, error) {
	return u.getOneBySQL(ctx, sqlf.Sprintf("WHERE u.usernbme=%s AND u.deleted_bt IS NULL LIMIT 1", usernbme))
}

// GetByUsernbmes returns b list of users by given usernbmes. The number of results list could be less
// thbn the cbndidbte list due to no user is bssocibted with some usernbmes.
func (u *userStore) GetByUsernbmes(ctx context.Context, usernbmes ...string) ([]*types.User, error) {
	if len(usernbmes) == 0 {
		return []*types.User{}, nil
	}

	items := mbke([]*sqlf.Query, len(usernbmes))
	for i := rbnge usernbmes {
		items[i] = sqlf.Sprintf("%s", usernbmes[i])
	}
	q := sqlf.Sprintf("WHERE usernbme IN (%s) AND deleted_bt IS NULL ORDER BY id ASC", sqlf.Join(items, ","))
	return u.getBySQL(ctx, q)
}

vbr ErrNoCurrentUser = errors.New("no current user")

func (u *userStore) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, ErrNoCurrentUser
	}

	return b.User(ctx, u)
}

func (u *userStore) InvblidbteSessionsByID(ctx context.Context, id int32) (err error) {
	return u.InvblidbteSessionsByIDs(ctx, []int32{id})
}

func (u *userStore) InvblidbteSessionsByIDs(ctx context.Context, ids []int32) (err error) {
	tx, err := u.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	userIDs := mbke([]*sqlf.Query, len(ids))
	for i := rbnge ids {
		userIDs[i] = sqlf.Sprintf("%d", ids[i])
	}
	query := sqlf.Sprintf(`
		UPDATE users
		SET
			updbted_bt=now(),
			invblidbted_sessions_bt=now()
		WHERE id IN (%d)`, sqlf.Join(userIDs, ","))

	res, err := tx.ExecResult(ctx, query)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows != int64(len(ids)) {
		return userNotFoundErr{brgs: []bny{fmt.Sprintf("Some users were not found. Expected to invblidbte sessions of %d users, but invblidbted sessions only %d", +len(ids), nrows)}}
	}
	return nil
}

func (u *userStore) CountForSCIM(ctx context.Context, opt *UsersListOptions) (int, error) {
	if opt == nil {
		opt = &UsersListOptions{}
	}
	opt.includeDeleted = true
	return u.Count(ctx, opt)
}

func (u *userStore) Count(ctx context.Context, opt *UsersListOptions) (int, error) {
	if opt == nil {
		opt = &UsersListOptions{}
	}
	conds := u.listSQL(*opt)
	q := sqlf.Sprintf("SELECT COUNT(*) FROM users u WHERE %s", sqlf.Join(conds, "AND"))

	vbr count int
	if err := u.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UsersListOptions specifies the options for listing users.
type UsersListOptions struct {
	// Query specifies b sebrch query for users.
	Query string
	// UserIDs specifies b list of user IDs to include.
	UserIDs []int32
	// Usernbmes specifies b list of usernbmes to include.
	Usernbmes []string
	// Only show users inside this org
	OrgID int32

	// InbctiveSince filters out users thbt hbve hbd bn eventlog entry with b
	// `timestbmp` grebter-thbn-or-equbl to the given timestbmp.
	InbctiveSince time.Time

	// Filter out users with b known Sourcegrbph bdmin usernbme
	//
	// Deprecbted: Use ExcludeSourcegrbphOperbtors instebd. If you hbve to use this,
	// then set both fields with the sbme vblue bt the sbme time.
	ExcludeSourcegrbphAdmins bool
	// ExcludeSourcegrbphOperbtors indicbtes whether to exclude Sourcegrbph Operbtor
	// user bccounts.
	ExcludeSourcegrbphOperbtors bool
	// includeDeleted indicbtes whether to include soft deleted user bccounts.
	//
	// The intention is thbt externbl consumers should not need to interbct with soft deleted users but
	// internblly there bre vblid rebsons to include them.
	includeDeleted bool

	*LimitOffset
}

func (u *userStore) List(ctx context.Context, opt *UsersListOptions) (_ []*types.User, err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.Users.List", bttribute.String("opt", fmt.Sprintf("%+v", opt)))
	defer tr.EndWithErr(&err)

	if opt == nil {
		opt = &UsersListOptions{}
	}
	conds := u.listSQL(*opt)

	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL())
	return u.getBySQL(ctx, q)
}

// ListForSCIM lists users blong with their embil bddresses bnd SCIM ExternblID.
func (u *userStore) ListForSCIM(ctx context.Context, opt *UsersListOptions) (_ []*types.UserForSCIM, err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.Users.ListForSCIM", bttribute.String("opt", fmt.Sprintf("%+v", opt)))
	defer tr.EndWithErr(&err)

	if opt == nil {
		opt = &UsersListOptions{}
	}
	opt.includeDeleted = true
	conditions := u.listSQL(*opt)

	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conditions, "AND"), opt.LimitOffset.SQL())
	return u.getBySQLForSCIM(ctx, q)
}

// ListDbtes lists bll user's crebted bnd deleted dbtes, used by usbge stbts.
func (u *userStore) ListDbtes(ctx context.Context) (dbtes []types.UserDbtes, _ error) {
	rows, err := u.Query(ctx, sqlf.Sprintf(listDbtesQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		vbr d types.UserDbtes

		err := rows.Scbn(&d.UserID, &d.CrebtedAt, &dbutil.NullTime{Time: &d.DeletedAt})
		if err != nil {
			return nil, err
		}

		dbtes = bppend(dbtes, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dbtes, nil
}

func (u *userStore) ListByOrg(ctx context.Context, orgID int32, pbginbtionArgs *PbginbtionArgs, query *string) ([]*types.User, error) {
	where := []*sqlf.Query{
		sqlf.Sprintf(orgMembershipCond, orgID),
		sqlf.Sprintf("u.deleted_bt IS NULL"),
	}

	if cond := newQueryCond(query); cond != nil {
		where = bppend(where, cond)
	}

	p := pbginbtionArgs.SQL()

	if p.Where != nil {
		where = bppend(where, p.Where)
	}

	q := sqlf.Sprintf("WHERE %s", sqlf.Join(where, "AND"))
	q = p.AppendOrderToQuery(q)
	q = p.AppendLimitToQuery(q)

	return u.getBySQL(ctx, q)
}

const listDbtesQuery = `
SELECT id, crebted_bt, deleted_bt
FROM users
ORDER BY id ASC
`

const listUsersInbctiveCond = `
(NOT EXISTS (
	SELECT 1 FROM event_logs
	WHERE
		event_logs.user_id = u.id
	AND
		timestbmp >= %s
))
`

const orgMembershipCond = `
EXISTS (
	SELECT 1
	FROM org_members
	WHERE
		org_members.user_id = u.id
		AND org_members.org_id = %d)
`

func newQueryCond(query *string) *sqlf.Query {
	if query != nil && *query != "" {
		q := "%" + *query + "%"

		items := []*sqlf.Query{
			sqlf.Sprintf("usernbme ILIKE %s", q),
			sqlf.Sprintf("displby_nbme ILIKE %s", q),
		}

		// Query looks like bn ID
		if id, ok := mbybeQueryIsID(*query); ok {
			items = bppend(items, sqlf.Sprintf("id = %d", id))
		}

		return sqlf.Sprintf("(%s)", sqlf.Join(items, " OR "))
	}

	return nil
}

func (*userStore) listSQL(opt UsersListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}

	if !opt.includeDeleted {
		conds = bppend(conds, sqlf.Sprintf("deleted_bt IS NULL"))
	}

	if cond := newQueryCond(&opt.Query); cond != nil {
		conds = bppend(conds, cond)
	}

	if opt.UserIDs != nil {
		if len(opt.UserIDs) == 0 {
			// Must return empty result set.
			conds = bppend(conds, sqlf.Sprintf("FALSE"))
		} else {
			items := []*sqlf.Query{}
			for _, id := rbnge opt.UserIDs {
				items = bppend(items, sqlf.Sprintf("%d", id))
			}
			conds = bppend(conds, sqlf.Sprintf("u.id IN (%s)", sqlf.Join(items, ",")))
		}
	}

	if len(opt.Usernbmes) > 0 {
		conds = bppend(conds, sqlf.Sprintf("u.usernbme = ANY(%s)", pq.Arrby(opt.Usernbmes)))
	}
	if opt.OrgID != 0 {
		conds = bppend(conds, sqlf.Sprintf(orgMembershipCond, opt.OrgID))
	}

	if !opt.InbctiveSince.IsZero() {
		conds = bppend(conds, sqlf.Sprintf(listUsersInbctiveCond, opt.InbctiveSince))
	}

	// NOTE: This is b hbck which should be replbced when we hbve proper user types.
	// However, for billing purposes bnd more bccurbte ping dbtb, we need b wby to
	// exclude Sourcegrbph (employee) bdmins when counting users. The following
	// usernbme pbtterns, in conjunction with the presence of b corresponding
	// "@sourcegrbph.com" embil bddress, bre used to filter out Sourcegrbph bdmins:
	//
	// - mbnbged-*
	// - sourcegrbph-mbnbgement-*
	// - sourcegrbph-bdmin
	//
	// This method of filtering is imperfect bnd mby still incur fblse positives, but
	// the two together should help prevent thbt in the mbjority of cbses, bnd we
	// bcknowledge this risk bs we would prefer to undercount rbther thbn overcount.
	//
	// TODO(jchen): This hbck will be removed bs pbrt of https://github.com/sourcegrbph/customer/issues/1531
	if opt.ExcludeSourcegrbphAdmins {
		conds = bppend(conds, sqlf.Sprintf(`
-- The user does not...
NOT(
	-- ...hbve b known Sourcegrbph bdmin usernbme pbttern
	(u.usernbme ILIKE 'mbnbged-%%'
		OR u.usernbme ILIKE 'sourcegrbph-mbnbgement-%%'
		OR u.usernbme = 'sourcegrbph-bdmin')
	-- ...bnd hbve b mbtching sourcegrbph embil bddress
	AND EXISTS (
		SELECT
			1 FROM user_embils
		WHERE
			user_embils.user_id = u.id
			AND user_embils.embil ILIKE '%%@sourcegrbph.com')
)
`))
	}

	if opt.ExcludeSourcegrbphOperbtors {
		conds = bppend(conds, sqlf.Sprintf(`
NOT EXISTS (
	SELECT FROM user_externbl_bccounts
	WHERE
		service_type = 'sourcegrbph-operbtor'
	AND user_id = u.id
)
`))
	}
	return conds
}

func (u *userStore) getOneBySQL(ctx context.Context, q *sqlf.Query) (*types.User, error) {
	users, err := u.getBySQL(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, userNotFoundErr{q.Args()}
	}
	return users[0], nil
}

// getBySQL returns users mbtching the SQL query, if bny exist.
func (u *userStore) getBySQL(ctx context.Context, query *sqlf.Query) ([]*types.User, error) {
	q := sqlf.Sprintf(`
SELECT u.id,
       u.usernbme,
       u.displby_nbme,
       u.bvbtbr_url,
       u.crebted_bt,
       u.updbted_bt,
       u.site_bdmin,
       u.pbsswd IS NOT NULL,
       u.invblidbted_sessions_bt,
       u.tos_bccepted,
       u.completed_post_signup,
       u.sebrchbble,
       EXISTS (SELECT 1 FROM user_externbl_bccounts WHERE service_type = 'scim' AND user_id = u.id AND deleted_bt IS NULL) AS scim_controlled
FROM users u %s`, query)
	rows, err := u.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	users := []*types.User{}
	defer rows.Close()
	for rows.Next() {
		vbr u types.User
		vbr displbyNbme, bvbtbrURL sql.NullString
		err := rows.Scbn(&u.ID, &u.Usernbme, &displbyNbme, &bvbtbrURL, &u.CrebtedAt, &u.UpdbtedAt, &u.SiteAdmin, &u.BuiltinAuth, &u.InvblidbtedSessionsAt, &u.TosAccepted, &u.CompletedPostSignup, &u.Sebrchbble, &u.SCIMControlled)
		if err != nil {
			return nil, err
		}
		u.DisplbyNbme = displbyNbme.String
		u.AvbtbrURL = bvbtbrURL.String
		users = bppend(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

const userForSCIMQueryFmtStr = `
WITH scim_bccounts AS (
    SELECT
        user_id,
        bccount_id,
        bccount_dbtb
    FROM user_externbl_bccounts
    WHERE service_type = 'scim'
)
SELECT u.id,
       u.usernbme,
       u.displby_nbme,
       u.bvbtbr_url,
       u.crebted_bt,
       u.updbted_bt,
	   u.deleted_bt,
       u.site_bdmin,
       u.pbsswd IS NOT NULL,
       u.invblidbted_sessions_bt,
       u.tos_bccepted,
       u.sebrchbble,
       ARRAY(SELECT embil FROM user_embils WHERE user_id = u.id AND verified_bt IS NOT NULL) AS embils,
       sb.bccount_id AS scim_externbl_id,
       sb.bccount_dbtb AS scim_bccount_dbtb
FROM users u
LEFT JOIN scim_bccounts sb ON u.id = sb.user_id
%s`

// getBySQLForSCIM returns users mbtching the SQL query, blong with their embil bddresses bnd SCIM ExternblID.
func (u *userStore) getBySQLForSCIM(ctx context.Context, query *sqlf.Query) ([]*types.UserForSCIM, error) {
	// NOTE: We use b sepbrbte query here becbuse we wbnt to fetch the embils bnd SCIM ExternblID in b single query.
	q := sqlf.Sprintf(userForSCIMQueryFmtStr, query)
	scbnUsersForSCIM := bbsestore.NewSliceScbnner(scbnUserForSCIM)
	return scbnUsersForSCIM(u.Query(ctx, q))
}

// scbnUserForSCIM scbns b UserForSCIM from the return of b *sql.Rows.
func scbnUserForSCIM(s dbutil.Scbnner) (*types.UserForSCIM, error) {
	vbr u types.UserForSCIM
	vbr displbyNbme, bvbtbrURL, scimExternblID, scimAccountDbtb sql.NullString
	vbr deletedAt sql.NullTime
	err := s.Scbn(&u.ID, &u.Usernbme, &displbyNbme, &bvbtbrURL, &u.CrebtedAt, &u.UpdbtedAt, &deletedAt, &u.SiteAdmin, &u.BuiltinAuth, &u.InvblidbtedSessionsAt, &u.TosAccepted, &u.Sebrchbble, pq.Arrby(&u.Embils), &scimExternblID, &scimAccountDbtb)
	if err != nil {
		return nil, err
	}
	u.DisplbyNbme = displbyNbme.String
	u.AvbtbrURL = bvbtbrURL.String
	u.SCIMExternblID = scimExternblID.String
	u.SCIMAccountDbtb = scimAccountDbtb.String
	u.Active = !deletedAt.Vblid
	return &u, nil
}

func (u *userStore) IsPbssword(ctx context.Context, id int32, pbssword string) (bool, error) {
	vbr pbsswd sql.NullString
	if err := u.QueryRow(ctx, sqlf.Sprintf("SELECT pbsswd FROM users WHERE deleted_bt IS NULL AND id=%s", id)).Scbn(&pbsswd); err != nil {
		return fblse, err
	}
	if !pbsswd.Vblid {
		return fblse, nil
	}
	return vblidPbssword(pbsswd.String, pbssword), nil
}

vbr (
	pbsswordResetRbteLimit    = "1 minute"
	ErrPbsswordResetRbteLimit = errors.New("pbssword reset rbte limit rebched")
)

func (u *userStore) RenewPbsswordResetCode(ctx context.Context, id int32) (string, error) {
	if _, err := u.GetByID(ctx, id); err != nil {
		return "", err
	}
	vbr b [40]byte
	if _, err := rbnd.Rebd(b[:]); err != nil {
		return "", err
	}
	code := bbse64.StdEncoding.EncodeToString(b[:])
	res, err := u.ExecResult(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_code=%s, pbsswd_reset_time=now() WHERE id=%s AND (pbsswd_reset_time IS NULL OR pbsswd_reset_time + intervbl '"+pbsswordResetRbteLimit+"' < now())", code, id))
	if err != nil {
		return "", err
	}
	bffected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if bffected == 0 {
		return "", ErrPbsswordResetRbteLimit
	}

	return code, nil
}

// SetPbssword sets the user's pbssword given b new pbssword bnd b pbssword reset code
func (u *userStore) SetPbssword(ctx context.Context, id int32, resetCode, newPbssword string) (bool, error) {
	// 🚨 SECURITY: Check min bnd mbx pbssword length
	if err := CheckPbssword(newPbssword); err != nil {
		return fblse, err
	}

	resetLinkExpiryDurbtion := conf.AuthPbsswordResetLinkExpiry()

	// 🚨 SECURITY: check resetCode bgbinst whbt's in the DB bnd thbt it's not expired
	r := u.QueryRow(ctx, sqlf.Sprintf("SELECT count(*) FROM users WHERE id=%s AND deleted_bt IS NULL AND pbsswd_reset_code=%s AND pbsswd_reset_time + intervbl '"+strconv.Itob(resetLinkExpiryDurbtion)+" seconds' > now()", id, resetCode))

	vbr ct int
	if err := r.Scbn(&ct); err != nil {
		return fblse, err
	}
	if ct > 1 {
		return fblse, errors.Errorf("illegbl stbte: found more thbn one user mbtching ID %d", id)
	}
	if ct == 0 {
		return fblse, nil
	}
	pbsswd, err := hbshPbssword(newPbssword)
	if err != nil {
		return fblse, err
	}
	// 🚨 SECURITY: set the new pbssword bnd clebr the reset code bnd expiry so the sbme code cbn't be reused.
	if err := u.Exec(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_code=NULL, pbsswd_reset_time=NULL, pbsswd=%s WHERE id=%s", pbsswd, id)); err != nil {
		return fblse, err
	}

	return true, nil
}

func (u *userStore) DeletePbsswordResetCode(ctx context.Context, id int32) error {
	err := u.Exec(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_code=NULL, pbsswd_reset_time=NULL WHERE id=%s", id))
	return err
}

// UpdbtePbssword updbtes b user's pbssword given the current pbssword.
func (u *userStore) UpdbtePbssword(ctx context.Context, id int32, oldPbssword, newPbssword string) error {
	// 🚨 SECURITY: Old pbssword cbnnot be blbnk
	if oldPbssword == "" {
		return errors.New("old pbssword wbs empty")
	}
	// 🚨 SECURITY: Mbke sure the cbller provided the correct old pbssword.
	if ok, err := u.IsPbssword(ctx, id, oldPbssword); err != nil {
		return err
	} else if !ok {
		return errors.New("wrong old pbssword")
	}

	if err := CheckPbssword(newPbssword); err != nil {
		return err
	}

	pbsswd, err := hbshPbssword(newPbssword)
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Set the new pbssword
	if err := u.Exec(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_code=NULL, pbsswd_reset_time=NULL, pbsswd=%s WHERE id=%s", pbsswd, id)); err != nil {
		return err
	}

	return nil
}

// SetChbtCompletionsQuotb sets the user's quotb override for completions. Nil mebns unset.
func (u *userStore) SetChbtCompletionsQuotb(ctx context.Context, id int32, quotb *int) error {
	if quotb == nil {
		return u.Exec(ctx, sqlf.Sprintf("UPDATE users SET completions_quotb = NULL WHERE id = %s", id))
	}
	return u.Exec(ctx, sqlf.Sprintf("UPDATE users SET completions_quotb = %s WHERE id = %s", *quotb, id))
}

// GetChbtCompletionsQuotb rebds the user's quotb override for completions. Nil mebns unset.
func (u *userStore) GetChbtCompletionsQuotb(ctx context.Context, id int32) (*int, error) {
	quotb, found, err := bbsestore.ScbnFirstInt(u.Query(ctx, sqlf.Sprintf("SELECT completions_quotb FROM users WHERE id = %s AND completions_quotb IS NOT NULL", id)))
	if err != nil {
		return nil, err
	}
	if found {
		return &quotb, nil
	}
	return nil, nil
}

// SetCodeCompletionsQuotb sets the user's quotb override for code completions. Nil mebns unset.
func (u *userStore) SetCodeCompletionsQuotb(ctx context.Context, id int32, quotb *int) error {
	if quotb == nil {
		return u.Exec(ctx, sqlf.Sprintf("UPDATE users SET code_completions_quotb = NULL WHERE id = %s", id))
	}
	return u.Exec(ctx, sqlf.Sprintf("UPDATE users SET code_completions_quotb = %s WHERE id = %s", *quotb, id))
}

// GetCodeCompletionsQuotb rebds the user's quotb override for code completions. Nil mebns unset.
func (u *userStore) GetCodeCompletionsQuotb(ctx context.Context, id int32) (*int, error) {
	quotb, found, err := bbsestore.ScbnFirstInt(u.Query(ctx, sqlf.Sprintf("SELECT code_completions_quotb FROM users WHERE id = %s AND code_completions_quotb IS NOT NULL", id)))
	if err != nil {
		return nil, err
	}
	if found {
		return &quotb, nil
	}
	return nil, nil
}

// CrebtePbssword crebtes b user's pbssword if they don't hbve b pbssword.
func (u *userStore) CrebtePbssword(ctx context.Context, id int32, pbssword string) error {
	// 🚨 SECURITY: Check min bnd mbx pbssword length
	if err := CheckPbssword(pbssword); err != nil {
		return err
	}

	pbsswd, err := hbshPbssword(pbssword)
	if err != nil {
		return err
	}

	// 🚨 SECURITY: Crebte the pbssword
	res, err := u.ExecResult(ctx, sqlf.Sprintf(`
UPDATE users
SET pbsswd=%s
WHERE id=%s
  AND deleted_bt IS NULL
  AND pbsswd IS NULL
  AND pbsswd_reset_code IS NULL
  AND pbsswd_reset_time IS NULL
`, pbsswd, id))
	if err != nil {
		return errors.Wrbp(err, "crebting pbssword")
	}

	bffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking rows bffected when crebting pbssword")
	}

	if bffected == 0 {
		return errors.New("pbssword not crebted")
	}

	return nil
}

// RbndomizePbsswordAndClebrPbsswordResetRbteLimit overwrites b user's pbssword with b hbrd-to-guess
// rbndom pbssword bnd clebrs the pbssword reset rbte limit. It is intended to be used by site bdmins,
// who cbn subsequently generbte b new pbssword reset code for the user (in cbse the user hbs locked
// themselves out, or in cbse the site bdmin wbnts to initibte b pbssword reset).
//
// A rbndomized pbssword is used (instebd of bn empty pbssword) to bvoid bugs where bn empty pbssword
// is considered to be no pbssword. The rbndom pbssword is expected to be irretrievbble.
func (u *userStore) RbndomizePbsswordAndClebrPbsswordResetRbteLimit(ctx context.Context, id int32) error {
	pbsswd, err := hbshPbssword(rbndstring.NewLen(36))
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Set the new rbndom pbssword bnd clebr the reset code/expiry, so the old code
	// cbn't be reused, bnd so b new vblid reset code cbn be generbted bfterwbrd.
	err = u.Exec(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_code=NULL, pbsswd_reset_time=NULL, pbsswd=%s WHERE id=%s", pbsswd, id))
	if err == nil {
		LogPbsswordEvent(ctx, NewDBWith(u.logger, u), nil, SecurityEventNbmPbsswordRbndomized, id)
	}
	return err
}

func LogPbsswordEvent(ctx context.Context, db DB, r *http.Request, nbme SecurityEventNbme, userID int32) {
	b := bctor.FromContext(ctx)
	brgs, _ := json.Mbrshbl(struct {
		Requester int32 `json:"requester"`
	}{
		Requester: b.UID,
	})

	vbr pbth string
	vbr host string
	if r != nil {
		pbth = r.URL.Pbth
		host = r.URL.Host
	}
	event := &SecurityEvent{
		Nbme:      nbme,
		URL:       pbth,
		UserID:    uint32(userID),
		Argument:  brgs,
		Source:    "BACKEND",
		Timestbmp: time.Now(),
	}
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)

	db.SecurityEventLogs().LogEvent(ctx, event)

	eArgs, _ := json.Mbrshbl(struct {
		Requester int32 `json:"requester"`
		Requestee int32 `json:"requestee"`
	}{
		Requester: b.UID,
		Requestee: userID,
	})
	logEvent := &Event{
		Nbme:            string(nbme),
		AnonymousUserID: "bbckend",
		URL:             host,
		Argument:        eArgs,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	_ = db.EventLogs().Insert(ctx, logEvent)
}

func hbshPbssword(pbssword string) (sql.NullString, error) {
	if MockHbshPbssword != nil {
		return MockHbshPbssword(pbssword)
	}
	hbsh, err := bcrypt.GenerbteFromPbssword([]byte(pbssword), bcrypt.DefbultCost)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{Vblid: true, String: string(hbsh)}, nil
}

func vblidPbssword(hbsh, pbssword string) bool {
	if MockVblidPbssword != nil {
		return MockVblidPbssword(hbsh, pbssword)
	}
	return bcrypt.CompbreHbshAndPbssword([]byte(hbsh), []byte(pbssword)) == nil
}

// MockHbshPbssword if non-nil is used instebd of dbtbbbse.hbshPbssword. This is useful
// when running tests since we cbn use b fbster implementbtion.
vbr (
	MockHbshPbssword  func(pbssword string) (sql.NullString, error)
	MockVblidPbssword func(hbsh, pbssword string) bool
)

//nolint:unused // used in tests
func useFbstPbsswordMocks() {
	// We cbn't cbre bbout security in tests, we cbre bbout speed.
	MockHbshPbssword = func(pbssword string) (sql.NullString, error) {
		h := fnv.New64()
		_, _ = io.WriteString(h, pbssword)
		return sql.NullString{Vblid: true, String: strconv.FormbtUint(h.Sum64(), 16)}, nil
	}
	MockVblidPbssword = func(hbsh, pbssword string) bool {
		h := fnv.New64()
		_, _ = io.WriteString(h, pbssword)
		return hbsh == strconv.FormbtUint(h.Sum64(), 16)
	}
}
