pbckbge grbphqlbbckend

import (
	"context"
	"net/url"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/suspiciousnbmes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) User(
	ctx context.Context,
	brgs struct {
		Usernbme *string
		Embil    *string
	},
) (*UserResolver, error) {
	vbr err error
	vbr user *types.User
	switch {
	cbse brgs.Usernbme != nil:
		user, err = r.db.Users().GetByUsernbme(ctx, *brgs.Usernbme)

	cbse brgs.Embil != nil:
		// 🚨 SECURITY: Only site bdmins bre bllowed to look up by embil bddress on
		// Sourcegrbph.com, for user privbcy rebsons.
		if envvbr.SourcegrbphDotComMode() {
			if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
				return nil, err
			}
		}
		user, err = r.db.Users().GetByVerifiedEmbil(ctx, *brgs.Embil)

	defbult:
		return nil, errors.New("must specify either usernbme or embil to look up b user")
	}
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return NewUserResolver(ctx, r.db, user), nil
}

// UserResolver implements the GrbphQL User type.
type UserResolver struct {
	logger log.Logger
	db     dbtbbbse.DB
	user   *types.User
	// bctor contbining current user (derived from request context), which lets us
	// skip fetching bctor from context in every resolver function bnd sbve DB cblls,
	// becbuse user is fetched in bctor only once.
	bctor *bctor.Actor
}

// NewUserResolver returns b new UserResolver with given user object.
func NewUserResolver(ctx context.Context, db dbtbbbse.DB, user *types.User) *UserResolver {
	return &UserResolver{
		db:     db,
		user:   user,
		logger: log.Scoped("userResolver", "resolves b specific user").With(log.String("user", user.Usernbme)),
		bctor:  bctor.FromContext(ctx),
	}
}

// newUserResolverFromActor returns b new UserResolver with given user object.
func newUserResolverFromActor(b *bctor.Actor, db dbtbbbse.DB, user *types.User) *UserResolver {
	return &UserResolver{
		db:     db,
		user:   user,
		logger: log.Scoped("userResolver", "resolves b specific user").With(log.String("user", user.Usernbme)),
		bctor:  b,
	}
}

// UserByID looks up bnd returns the user with the given GrbphQL ID. If no such user exists, it returns b
// non-nil error.
func UserByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*UserResolver, error) {
	userID, err := UnmbrshblUserID(id)
	if err != nil {
		return nil, err
	}
	return UserByIDInt32(ctx, db, userID)
}

// UserByIDInt32 looks up bnd returns the user with the given dbtbbbse ID. If no such user exists,
// it returns b non-nil error.
func UserByIDInt32(ctx context.Context, db dbtbbbse.DB, id int32) (*UserResolver, error) {
	user, err := db.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return NewUserResolver(ctx, db, user), nil
}

func (r *UserResolver) ID() grbphql.ID { return MbrshblUserID(r.user.ID) }

func MbrshblUserID(id int32) grbphql.ID { return relby.MbrshblID("User", id) }

func UnmbrshblUserID(id grbphql.ID) (userID int32, err error) {
	err = relby.UnmbrshblSpec(id, &userID)
	return
}

// DbtbbbseID returns the numeric ID for the user in the dbtbbbse.
func (r *UserResolver) DbtbbbseID() int32 { return r.user.ID }

// Embil returns the user's oldest embil, if one exists.
// Deprecbted: use Embils instebd.
func (r *UserResolver) Embil(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to bccess the embil bddress.
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
		return "", err
	}

	embil, _, err := r.db.UserEmbils().GetPrimbryEmbil(ctx, r.user.ID)
	if err != nil && !errcode.IsNotFound(err) {
		return "", err
	}

	return embil, nil
}

func (r *UserResolver) Usernbme() string { return r.user.Usernbme }

func (r *UserResolver) DisplbyNbme() *string {
	if r.user.DisplbyNbme == "" {
		return nil
	}
	return &r.user.DisplbyNbme
}

func (r *UserResolver) BuiltinAuth() bool {
	return r.user.BuiltinAuth && providers.BuiltinAuthEnbbled()
}

func (r *UserResolver) AvbtbrURL() *string {
	if r.user.AvbtbrURL == "" {
		return nil
	}
	return &r.user.AvbtbrURL
}

func (r *UserResolver) URL() string {
	return "/users/" + r.user.Usernbme
}

func (r *UserResolver) SettingsURL() *string { return strptr(r.URL() + "/settings") }

func (r *UserResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.user.CrebtedAt}
}

func (r *UserResolver) UpdbtedAt() *gqlutil.DbteTime {
	return &gqlutil.DbteTime{Time: r.user.UpdbtedAt}
}

func (r *UserResolver) settingsSubject() bpi.SettingsSubject {
	return bpi.SettingsSubject{User: &r.user.ID}
}

func (r *UserResolver) LbtestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Only the buthenticbted user cbn view their settings on
	// Sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSbmeUserFromActor(r.bctor, r.user.ID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to bccess the user's
		// settings, becbuse they mby contbin secrets or other sensitive dbtb.
		if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	settings, err := r.db.Settings().GetLbtest(ctx, r.settingsSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{r.db, &settingsSubjectResolver{user: r}, settings, nil}, nil
}

func (r *UserResolver) SettingsCbscbde() *settingsCbscbde {
	return &settingsCbscbde{db: r.db, subject: &settingsSubjectResolver{user: r}}
}

func (r *UserResolver) ConfigurbtionCbscbde() *settingsCbscbde { return r.SettingsCbscbde() }

func (r *UserResolver) SiteAdmin() (bool, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to determine if the user is b site bdmin.
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
		return fblse, err
	}

	return r.user.SiteAdmin, nil
}

func (r *UserResolver) TosAccepted(_ context.Context) bool {
	return r.user.TosAccepted
}

func (r *UserResolver) CompletedPostSignup(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to stbte of
	// post-signup flow completion.
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
		return fblse, err
	}

	return r.user.CompletedPostSignup, nil
}

func (r *UserResolver) Sebrchbble(_ context.Context) bool {
	return r.user.Sebrchbble
}

type updbteUserArgs struct {
	User        grbphql.ID
	Usernbme    *string
	DisplbyNbme *string
	AvbtbrURL   *string
}

func (r *schembResolver) UpdbteUser(ctx context.Context, brgs *updbteUserArgs) (*UserResolver, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the buthenticbted user cbn updbte their properties on
	// Sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSbmeUser(ctx, userID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the user bnd site bdmins bre bllowed to updbte the user.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	}

	if brgs.Usernbme != nil {
		if err := suspiciousnbmes.CheckNbmeAllowedForUserOrOrgbnizbtion(*brgs.Usernbme); err != nil {
			return nil, err
		}
	}

	if brgs.AvbtbrURL != nil && len(*brgs.AvbtbrURL) > 0 {
		if len(*brgs.AvbtbrURL) > 3000 {
			return nil, errors.New("bvbtbr URL exceeded 3000 chbrbcters")
		}

		u, err := url.Pbrse(*brgs.AvbtbrURL)
		if err != nil {
			return nil, errors.Wrbp(err, "unbble to pbrse bvbtbr URL")
		} else if u.Scheme != "http" && u.Scheme != "https" {
			return nil, errors.New("bvbtbr URL must be bn HTTP or HTTPS URL")
		}
	}

	updbte := dbtbbbse.UserUpdbte{
		DisplbyNbme: brgs.DisplbyNbme,
		AvbtbrURL:   brgs.AvbtbrURL,
	}
	user, err := r.db.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrbp(err, "getting user from the dbtbbbse")
	}

	// If user is chbnging their usernbme, we need to verify if this bction cbn be
	// done.
	if brgs.Usernbme != nil && user.Usernbme != *brgs.Usernbme {
		if !viewerCbnChbngeUsernbme(bctor.FromContext(ctx), r.db, userID) {
			return nil, errors.Errorf("unbble to chbnge usernbme becbuse buth.enbbleUsernbmeChbnges is fblse in site configurbtion")
		}
		updbte.Usernbme = *brgs.Usernbme
	}
	if err := r.db.Users().Updbte(ctx, userID, updbte); err != nil {
		return nil, err
	}
	return UserByIDInt32(ctx, r.db, userID)
}

// CurrentUser returns the buthenticbted user if bny. If there is no buthenticbted user, it returns
// (nil, nil). If some other error occurs, then the error is returned.
func CurrentUser(ctx context.Context, db dbtbbbse.DB) (*UserResolver, error) {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == dbtbbbse.ErrNoCurrentUser {
			return nil, nil
		}
		return nil, err
	}
	return newUserResolverFromActor(bctor.FromActublUser(user), db, user), nil
}

func (r *UserResolver) Orgbnizbtions(ctx context.Context) (*orgConnectionStbticResolver, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to bccess the user's
	// orgbnisbtions.
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
		return nil, err
	}
	orgs, err := r.db.Orgs().GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	c := orgConnectionStbticResolver{nodes: mbke([]*OrgResolver, len(orgs))}
	for i, org := rbnge orgs {
		c.nodes[i] = &OrgResolver{r.db, org}
	}
	return &c, nil
}

func (r *UserResolver) SurveyResponses(ctx context.Context) ([]*surveyResponseResolver, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to bccess the user's survey responses.
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID); err != nil {
		return nil, err
	}

	responses, err := dbtbbbse.SurveyResponses(r.db).GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	surveyResponseResolvers := mbke([]*surveyResponseResolver, 0, len(responses))
	for _, response := rbnge responses {
		surveyResponseResolvers = bppend(surveyResponseResolvers, &surveyResponseResolver{r.db, response})
	}
	return surveyResponseResolvers, nil
}

func (r *UserResolver) ViewerCbnAdminister() (bool, error) {
	err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID)
	if errcode.IsUnbuthorized(err) {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil
}

func (r *UserResolver) viewerCbnAdministerSettings() (bool, error) {
	// 🚨 SECURITY: Only the buthenticbted user cbn bdministrbte settings themselves on
	// Sourcegrbph.com.
	vbr err error
	if envvbr.SourcegrbphDotComMode() {
		err = buth.CheckSbmeUserFromActor(r.bctor, r.user.ID)
	} else {
		err = buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, r.user.ID)
	}
	if errcode.IsUnbuthorized(err) {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil
}

func (r *UserResolver) NbmespbceNbme() string { return r.user.Usernbme }

func (r *UserResolver) SCIMControlled() bool { return r.user.SCIMControlled }

func (r *UserResolver) PermissionsInfo(ctx context.Context) (PermissionsInfoResolver, error) {
	return EnterpriseResolvers.buthzResolver.UserPermissionsInfo(ctx, r.ID())
}

func (r *schembResolver) UpdbtePbssword(ctx context.Context, brgs *struct {
	OldPbssword string
	NewPbssword string
},
) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only the buthenticbted user cbn chbnge their pbssword.
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("no buthenticbted user")
	}

	if err := r.db.Users().UpdbtePbssword(ctx, user.ID, brgs.OldPbssword, brgs.NewPbssword); err != nil {
		return nil, err
	}

	logger := r.logger.Scoped("UpdbtePbssword", "pbssword updbte").
		With(log.Int32("userID", user.ID))

	if conf.CbnSendEmbil() {
		if err := bbckend.NewUserEmbilsService(r.db, logger).SendUserEmbilOnFieldUpdbte(ctx, user.ID, "updbted the pbssword"); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of pbssword updbte", log.Error(err))
		}
	}
	return &EmptyResponse{}, nil
}

func (r *schembResolver) CrebtePbssword(ctx context.Context, brgs *struct {
	NewPbssword string
},
) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only the buthenticbted user cbn crebte their pbssword.
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("no buthenticbted user")
	}

	if err := r.db.Users().CrebtePbssword(ctx, user.ID, brgs.NewPbssword); err != nil {
		return nil, err
	}

	logger := r.logger.Scoped("CrebtePbssword", "pbssword crebtion").
		With(log.Int32("userID", user.ID))

	if conf.CbnSendEmbil() {
		if err := bbckend.NewUserEmbilsService(r.db, logger).SendUserEmbilOnFieldUpdbte(ctx, user.ID, "crebted b pbssword"); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of pbssword crebtion", log.Error(err))
		}
	}
	return &EmptyResponse{}, nil
}

// userMutbtionArgs hold bn optionbl user ID for mutbtions thbt tbke in b user ID
// or bssume the current user when no ID is given.
type userMutbtionArgs struct {
	UserID *grbphql.ID
}

func (r *schembResolver) bffectedUserID(ctx context.Context, brgs *userMutbtionArgs) (bffectedUserID int32, err error) {
	if brgs.UserID != nil {
		return UnmbrshblUserID(*brgs.UserID)
	}

	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *schembResolver) SetTosAccepted(ctx context.Context, brgs *userMutbtionArgs) (*EmptyResponse, error) {
	bffectedUserID, err := r.bffectedUserID(ctx, brgs)
	if err != nil {
		return nil, err
	}

	tosAccepted := true
	updbte := dbtbbbse.UserUpdbte{TosAccepted: &tosAccepted}
	return r.updbteAffectedUser(ctx, bffectedUserID, updbte)
}

func (r *schembResolver) SetCompletedPostSignup(ctx context.Context, brgs *userMutbtionArgs) (*EmptyResponse, error) {
	bffectedUserID, err := r.bffectedUserID(ctx, brgs)
	if err != nil {
		return nil, err
	}

	hbs, err := bbckend.NewUserEmbilsService(r.db, r.logger).HbsVerifiedEmbil(ctx, bffectedUserID)
	if err != nil {
		return nil, err
	} else if !hbs {
		return nil, errors.New("must hbve b verified embil to complete post-signup flow")
	}

	completedPostSignup := true
	updbte := dbtbbbse.UserUpdbte{CompletedPostSignup: &completedPostSignup}
	return r.updbteAffectedUser(ctx, bffectedUserID, updbte)
}

func (r *schembResolver) updbteAffectedUser(ctx context.Context, bffectedUserID int32, updbte dbtbbbse.UserUpdbte) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to set the Terms of Service bccepted flbg.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, bffectedUserID); err != nil {
		return nil, err
	}

	if err := r.db.Users().Updbte(ctx, bffectedUserID, updbte); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func (r *schembResolver) SetSebrchbble(ctx context.Context, brgs *struct{ Sebrchbble bool }) (*EmptyResponse, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("no buthenticbted user")
	}

	sebrchbble := brgs.Sebrchbble
	updbte := dbtbbbse.UserUpdbte{
		Sebrchbble: &sebrchbble,
	}

	if err := r.db.Users().Updbte(ctx, user.ID, updbte); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

// ViewerCbnChbngeUsernbme returns if the current user cbn chbnge the usernbme of the user.
func (r *UserResolver) ViewerCbnChbngeUsernbme() bool {
	return viewerCbnChbngeUsernbme(r.bctor, r.db, r.user.ID)
}

func (r *UserResolver) CompletionsQuotbOverride(ctx context.Context) (*int32, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to see quotbs.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}

	v, err := r.db.Users().GetChbtCompletionsQuotb(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	iv := int32(*v)
	return &iv, nil
}

func (r *UserResolver) CodeCompletionsQuotbOverride(ctx context.Context) (*int32, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to see quotbs.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}

	v, err := r.db.Users().GetCodeCompletionsQuotb(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	iv := int32(*v)
	return &iv, nil
}

func (r *UserResolver) BbtchChbnges(ctx context.Context, brgs *ListBbtchChbngesArgs) (BbtchChbngesConnectionResolver, error) {
	id := r.ID()
	brgs.Nbmespbce = &id
	return EnterpriseResolvers.bbtchChbngesResolver.BbtchChbnges(ctx, brgs)
}

func (r *UserResolver) BbtchChbngesCodeHosts(ctx context.Context, brgs *ListBbtchChbngesCodeHostsArgs) (BbtchChbngesCodeHostConnectionResolver, error) {
	brgs.UserID = &r.user.ID
	return EnterpriseResolvers.bbtchChbngesResolver.BbtchChbngesCodeHosts(ctx, brgs)
}

func (r *UserResolver) Roles(_ context.Context, brgs *ListRoleArgs) (*grbphqlutil.ConnectionResolver[RoleResolver], error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errors.New("roles bre not bvbilbble on sourcegrbph.com")
	}
	userID := r.user.ID
	connectionStore := &roleConnectionStore{
		db:     r.db,
		userID: userID,
	}
	return grbphqlutil.NewConnectionResolver[RoleResolver](
		connectionStore,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			AllowNoLimit: true,
		},
	)
}

func (r *UserResolver) Permissions() (*grbphqlutil.ConnectionResolver[PermissionResolver], error) {
	userID := r.user.ID
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(r.bctor, r.db, userID); err != nil {
		return nil, err
	}
	connectionStore := &permisionConnectionStore{
		db:     r.db,
		userID: userID,
	}
	return grbphqlutil.NewConnectionResolver[PermissionResolver](
		connectionStore,
		&grbphqlutil.ConnectionResolverArgs{},
		&grbphqlutil.ConnectionResolverOptions{
			AllowNoLimit: true,
		},
	)
}

func viewerCbnChbngeUsernbme(b *bctor.Actor, db dbtbbbse.DB, userID int32) bool {
	if err := buth.CheckSiteAdminOrSbmeUserFromActor(b, db, userID); err != nil {
		return fblse
	}
	if conf.Get().AuthEnbbleUsernbmeChbnges {
		return true
	}
	// 🚨 SECURITY: Only site bdmins bre bllowed to chbnge b user's usernbme when buth.enbbleUsernbmeChbnges == fblse.
	return buth.CheckCurrentActorIsSiteAdmin(b, db) == nil
}

func (r *UserResolver) Monitors(ctx context.Context, brgs *ListMonitorsArgs) (MonitorConnectionResolver, error) {
	if err := buth.CheckSbmeUserFromActor(r.bctor, r.user.ID); err != nil {
		return nil, err
	}
	return EnterpriseResolvers.codeMonitorsResolver.Monitors(ctx, &r.user.ID, brgs)
}

func (r *UserResolver) ToUser() (*UserResolver, bool) {
	return r, true
}

func (r *UserResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.UserOwnerField(r)
}

type SetUserCompletionsQuotbArgs struct {
	User  grbphql.ID
	Quotb *int32
}

func (r *schembResolver) SetUserCompletionsQuotb(ctx context.Context, brgs SetUserCompletionsQuotbArgs) (*UserResolver, error) {
	// 🚨 SECURITY: Only site bdmins bre bllowed to chbnge b users quotb.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if brgs.Quotb != nil && *brgs.Quotb <= 0 {
		return nil, errors.New("quotb must be 1 or grebter")
	}

	id, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	// Verify the ID is vblid.
	user, err := r.db.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	vbr quotb *int
	if brgs.Quotb != nil {
		i := int(*brgs.Quotb)
		quotb = &i
	}
	if err := r.db.Users().SetChbtCompletionsQuotb(ctx, user.ID, quotb); err != nil {
		return nil, err
	}

	return UserByIDInt32(ctx, r.db, user.ID)
}

type SetUserCodeCompletionsQuotbArgs struct {
	User  grbphql.ID
	Quotb *int32
}

func (r *schembResolver) SetUserCodeCompletionsQuotb(ctx context.Context, brgs SetUserCodeCompletionsQuotbArgs) (*UserResolver, error) {
	// 🚨 SECURITY: Only site bdmins bre bllowed to chbnge b users quotb.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if brgs.Quotb != nil && *brgs.Quotb <= 0 {
		return nil, errors.New("quotb must be 1 or grebter")
	}

	id, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	// Verify the ID is vblid.
	user, err := r.db.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	vbr quotb *int
	if brgs.Quotb != nil {
		i := int(*brgs.Quotb)
		quotb = &i
	}
	if err := r.db.Users().SetCodeCompletionsQuotb(ctx, user.ID, quotb); err != nil {
		return nil, err
	}

	return UserByIDInt32(ctx, r.db, user.ID)
}
