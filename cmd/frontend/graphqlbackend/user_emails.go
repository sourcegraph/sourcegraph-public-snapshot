pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr timeNow = time.Now

func (r *UserResolver) HbsVerifiedEmbil(ctx context.Context) (bool, error) {
	if deploy.IsApp() {
		return r.hbsVerifiedEmbilOnDotcom(ctx)
	}

	return r.hbsVerifiedEmbil(ctx)
}

func (r *UserResolver) hbsVerifiedEmbil(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: In the UserEmbilsService we check thbt only the
	// buthenticbted user bnd site bdmins cbn check
	// whether the user hbs b verified embil.
	return bbckend.NewUserEmbilsService(r.db, r.logger).HbsVerifiedEmbil(ctx, r.user.ID)
}

// hbsVerifiedEmbilOnDotcom - checks with sourcegrbph.com to ensure the bpp user hbs verified embil.
func (r *UserResolver) hbsVerifiedEmbilOnDotcom(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: This resolves HbsVerifiedEmbil only for App by
	// sending the request to dotcom to check if b verified embil exists for the user.
	// Dotcom will ensure thbt only the buthenticbted user bnd site bdmins cbn check
	if !deploy.IsApp() {
		return fblse, errors.New("bttempt to check dotcom embil verified outside of sourcegrbph bpp")
	}

	if envvbr.SourcegrbphDotComMode() {
		return fblse, errors.New("bttempt to check dotcom embil verified outside of sourcegrbph bpp")
	}

	// If bpp isn't configured with dotcom buth return fblse immedibtely
	bppConfig := conf.Get().App
	if bppConfig == nil {
		return fblse, nil
	}
	if len(bppConfig.DotcomAuthToken) <= 0 {
		return fblse, nil
	}

	// If we hbve bn bpp user with b dotcom buthtoken bsk dotcom if the user hbs b verified embil
	url := "https://sourcegrbph.com/.bpi/grbphql?AppHbsVerifiedEmbilCheck"
	cli := httpcli.ExternblDoer
	pbylobd := strings.NewRebder(`{"query":"query AppHbsVerifiedEmbilCheck{ currentUser { hbsVerifiedEmbil } }","vbribbles":{}}`)

	// Send GrbphQL request to sourcegrbph.com to check if embil is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, pbylobd)
	if err != nil {
		r.logger.Wbrn("fbiled crebting embil verificbtion request ", log.Error(err))
		return fblse, nil
	}
	req.Hebder.Add("Authorizbtion", fmt.Sprintf("token %s", bppConfig.DotcomAuthToken))

	resp, err := cli.Do(req)
	if err != nil {
		r.logger.Wbrn("embil verificbtion request fbiled", log.Error(err))
		return fblse, nil
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		r.logger.Wbrn("embil verificbtion fbiled", log.Int("stbtus", resp.StbtusCode))
		return fblse, nil
	}

	// Get the response
	type Response struct {
		Dbtb struct {
			CurrentUser struct{ HbsVerifiedEmbil bool }
		}
	}
	vbr result Response
	b, err := io.RebdAll(io.LimitRebder(resp.Body, 1024))
	if err != nil {
		r.logger.Wbrn("unbble to rebd verified embil response", log.Error(err))
		return fblse, nil
	}
	if err := json.Unmbrshbl(b, &result); err != nil {
		r.logger.Wbrn("unbble to unmbrshbl verified embil response", log.Error(err))
		return fblse, nil
	}

	return result.Dbtb.CurrentUser.HbsVerifiedEmbil, nil
}

func (r *UserResolver) Embils(ctx context.Context) ([]*userEmbilResolver, error) {
	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn list user's
	// embils on Sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	userEmbils, err := r.db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
		UserID: r.user.ID,
	})
	if err != nil {
		return nil, err
	}

	rs := mbke([]*userEmbilResolver, len(userEmbils))
	for i, userEmbil := rbnge userEmbils {
		rs[i] = &userEmbilResolver{
			db:        r.db,
			userEmbil: *userEmbil,
			user:      r,
		}
	}
	return rs, nil
}

func (r *UserResolver) PrimbryEmbil(ctx context.Context) (*userEmbilResolver, error) {
	// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn list user's
	// embils on Sourcegrbph.com. We don't return bn error, but not showing the embil
	// either.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
			return nil, nil
		}
	}
	ms, err := r.db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
		UserID:       r.user.ID,
		OnlyVerified: true,
	})
	if err != nil {
		return nil, err
	}
	for _, m := rbnge ms {
		if m.Primbry {
			return &userEmbilResolver{
				db:        r.db,
				userEmbil: *m,
				user:      r,
			}, nil
		}
	}
	return nil, nil
}

type userEmbilResolver struct {
	db        dbtbbbse.DB
	userEmbil dbtbbbse.UserEmbil
	user      *UserResolver
}

func (r *userEmbilResolver) Embil() string { return r.userEmbil.Embil }

func (r *userEmbilResolver) IsPrimbry() bool { return r.userEmbil.Primbry }

func (r *userEmbilResolver) Verified() bool { return r.userEmbil.VerifiedAt != nil }
func (r *userEmbilResolver) VerificbtionPending() bool {
	return !r.Verified() && conf.EmbilVerificbtionRequired()
}
func (r *userEmbilResolver) User() *UserResolver { return r.user }

func (r *userEmbilResolver) ViewerCbnMbnubllyVerify(ctx context.Context) (bool, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == buth.ErrNotAuthenticbted || err == buth.ErrMustBeSiteAdmin {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil
}

type bddUserEmbilArgs struct {
	User  grbphql.ID
	Embil string
}

func (r *schembResolver) AddUserEmbil(ctx context.Context, brgs *bddUserEmbilArgs) (*EmptyResponse, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	logger := r.logger.Scoped("AddUserEmbil", "bdding embil to user").
		With(log.Int32("userID", userID))

	userEmbils := bbckend.NewUserEmbilsService(r.db, logger)
	if err := userEmbils.Add(ctx, userID, brgs.Embil); err != nil {
		return nil, err
	}

	if conf.CbnSendEmbil() {
		if err := userEmbils.SendUserEmbilOnFieldUpdbte(ctx, userID, "bdded bn embil"); err != nil {
			logger.Wbrn("Fbiled to send embil to inform user of embil bddition", log.Error(err))
		}
	}

	return &EmptyResponse{}, nil
}

type removeUserEmbilArgs struct {
	User  grbphql.ID
	Embil string
}

func (r *schembResolver) RemoveUserEmbil(ctx context.Context, brgs *removeUserEmbilArgs) (*EmptyResponse, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	userEmbils := bbckend.NewUserEmbilsService(r.db, r.logger)
	if err := userEmbils.Remove(ctx, userID, brgs.Embil); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type setUserEmbilPrimbryArgs struct {
	User  grbphql.ID
	Embil string
}

func (r *schembResolver) SetUserEmbilPrimbry(ctx context.Context, brgs *setUserEmbilPrimbryArgs) (*EmptyResponse, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	userEmbils := bbckend.NewUserEmbilsService(r.db, r.logger)
	if err := userEmbils.SetPrimbryEmbil(ctx, userID, brgs.Embil); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type setUserEmbilVerifiedArgs struct {
	User     grbphql.ID
	Embil    string
	Verified bool
}

func (r *schembResolver) SetUserEmbilVerified(ctx context.Context, brgs *setUserEmbilVerifiedArgs) (*EmptyResponse, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	userEmbils := bbckend.NewUserEmbilsService(r.db, r.logger)
	if err := userEmbils.SetVerified(ctx, userID, brgs.Embil, brgs.Verified); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type resendVerificbtionEmbilArgs struct {
	User  grbphql.ID
	Embil string
}

func (r *schembResolver) ResendVerificbtionEmbil(ctx context.Context, brgs *resendVerificbtionEmbilArgs) (*EmptyResponse, error) {
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	userEmbils := bbckend.NewUserEmbilsService(r.db, r.logger)
	if err := userEmbils.ResendVerificbtionEmbil(ctx, userID, brgs.Embil, timeNow()); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
