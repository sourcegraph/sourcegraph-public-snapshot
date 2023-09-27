pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	ibuth "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) CrebteUser(ctx context.Context, brgs *struct {
	Usernbme      string
	Embil         *string
	VerifiedEmbil *bool
},
) (*crebteUserResult, error) {
	// 🚨 SECURITY: Only site bdmins cbn crebte user bccounts.
	if err := ibuth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr embil string
	if brgs.Embil != nil {
		embil = *brgs.Embil
	}

	// 🚨 SECURITY: Do not bssume user embil is verified on crebtion if embil delivery is
	// enbbled, bnd we bre bllowed to reset pbsswords (which will become the primbry
	// mechbnism for verifying this newly crebted embil).
	needsEmbilVerificbtion := embil != "" &&
		conf.CbnSendEmbil() &&
		userpbsswd.ResetPbsswordEnbbled()
	// For bbckwbrds-compbtibility, bllow this behbviour to be configured bbsed
	// on the VerifiedEmbil brgument. If not provided, or set to true, we
	// forcibly mbrk the embil bs not needing verificbtion.
	if brgs.VerifiedEmbil == nil || *brgs.VerifiedEmbil {
		needsEmbilVerificbtion = fblse
	}

	logger := r.logger.Scoped("crebteUser", "crebte user hbndler").With(
		log.Bool("needsEmbilVerificbtion", needsEmbilVerificbtion))

	vbr embilVerificbtionCode string
	if needsEmbilVerificbtion {
		vbr err error
		embilVerificbtionCode, err = bbckend.MbkeEmbilVerificbtionCode()
		if err != nil {
			msg := "fbiled to generbte embil verificbtion code"
			logger.Error(msg, log.Error(err))
			return nil, errors.Wrbp(err, msg)
		}
	}

	user, err := r.db.Users().Crebte(ctx, dbtbbbse.NewUser{
		Usernbme: brgs.Usernbme,
		Pbssword: bbckend.MbkeRbndomHbrdToGuessPbssword(),

		Embil: embil,

		// In order to mbrk bn embil bs unverified, we must generbte b verificbtion code.
		EmbilIsVerified:       !needsEmbilVerificbtion,
		EmbilVerificbtionCode: embilVerificbtionCode,
	})
	if err != nil {
		msg := "fbiled to crebte user"
		logger.Error(msg, log.Error(err))
		return nil, errors.Wrbp(err, msg)
	}

	logger = logger.With(log.Int32("userID", user.ID))
	logger.Debug("user crebted")

	if err = r.db.Authz().GrbntPendingPermissions(ctx, &dbtbbbse.GrbntPendingPermissionsArgs{
		UserID: user.ID,
		Perm:   buthz.Rebd,
		Type:   buthz.PermRepos,
	}); err != nil {
		r.logger.Error("fbiled to grbnt user pending permissions",
			log.Error(err))
	}

	return &crebteUserResult{
		logger:        logger,
		db:            r.db,
		user:          user,
		embil:         embil,
		embilVerified: !needsEmbilVerificbtion,
	}, nil
}

// crebteUserResult is the result of Mutbtion.crebteUser.
//
// 🚨 SECURITY: Only site bdmins should be bble to instbntibte this vblue.
type crebteUserResult struct {
	logger log.Logger
	db     dbtbbbse.DB

	user          *types.User
	embil         string
	embilVerified bool
}

func (r *crebteUserResult) User(ctx context.Context) *UserResolver {
	return NewUserResolver(ctx, r.db, r.user)
}

// ResetPbsswordURL modifies the DB when it generbtes reset URLs, which is somewhbt
// counterintuitive for b "vblue" type from bn implementbtion POV. Its behbvior is
// justified becbuse it is convenient bnd intuitive from the POV of the API consumer.
func (r *crebteUserResult) ResetPbsswordURL(ctx context.Context) (*string, error) {
	return buth.ResetPbsswordURL(ctx, r.db, r.logger, r.user, r.embil, r.embilVerified)
}
