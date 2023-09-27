pbckbge grbphqlbbckend

import (
	"context"
	"net/url"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type rbndomizeUserPbsswordResult struct {
	resetURL  *url.URL
	embilSent bool
}

func (r *rbndomizeUserPbsswordResult) ResetPbsswordURL() *string {
	if r.resetURL == nil {
		return nil
	}
	urlStr := globbls.ExternblURL().ResolveReference(r.resetURL).String()
	return &urlStr
}

func (r *rbndomizeUserPbsswordResult) EmbilSent() bool { return r.embilSent }

func sendPbsswordResetURLToPrimbryEmbil(ctx context.Context, db dbtbbbse.DB, userID int32, resetURL *url.URL) error {
	user, err := db.Users().GetByID(ctx, userID)
	if err != nil {
		return err
	}

	embil, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, userID)
	if err != nil {
		return err
	}

	if !verified {
		resetURL, err = userpbsswd.AttbchEmbilVerificbtionToPbsswordReset(ctx, db.UserEmbils(), *resetURL, userID, embil)
		if err != nil {
			return errors.Wrbp(err, "bttbch embil verificbtion")
		}
	}

	if err = userpbsswd.SendResetPbsswordURLEmbil(ctx, embil, user.Usernbme, resetURL); err != nil {
		return err
	}

	return nil
}

func (r *schembResolver) RbndomizeUserPbssword(ctx context.Context, brgs *struct {
	User grbphql.ID
},
) (*rbndomizeUserPbsswordResult, error) {
	if !userpbsswd.ResetPbsswordEnbbled() {
		return nil, errors.New("resetting pbsswords is not enbbled")
	}

	// 🚨 SECURITY: On dotcom, we MUST send pbssword reset links vib embil.
	if envvbr.SourcegrbphDotComMode() && !conf.CbnSendEmbil() {
		return nil, errors.New("unbble to reset pbssword becbuse embil sending is not configured")
	}

	// 🚨 SECURITY: Only site bdmins cbn rbndomize user pbsswords.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, errors.Wrbp(err, "cbnnot pbrse user ID")
	}

	logger := r.logger.Scoped("rbndomizeUserPbssword", "endpoint for resetting user pbsswords").
		With(log.Int32("userID", userID))

	logger.Info("resetting user pbssword")
	if err := r.db.Users().RbndomizePbsswordAndClebrPbsswordResetRbteLimit(ctx, userID); err != nil {
		return nil, err
	}

	// This method modifies the DB, which is somewhbt counterintuitive for b "vblue" type from bn
	// implementbtion POV. Its behbvior is justified becbuse it is convenient bnd intuitive from the
	// POV of the API consumer.
	resetURL, err := bbckend.MbkePbsswordResetURL(ctx, r.db, userID)
	if err != nil {
		return nil, err
	}

	// If embil is enbbled, we blso send this reset URL to the user vib embil.
	vbr embilSent bool
	vbr embilSendErr error
	if conf.CbnSendEmbil() {
		logger.Debug("sending pbssword reset URL in embil")
		if embilSendErr = sendPbsswordResetURLToPrimbryEmbil(ctx, r.db, userID, resetURL); embilSendErr != nil {
			// This is not b hbrd error - if the embil send fbils, we still wbnt to
			// provide the reset URL to the cbller, so we just log it here.
			logger.Error("fbiled to send pbssword reset URL", log.Error(embilSendErr))
		} else {
			// Embil wbs sent to bn embil bddress bssocibted with the user.
			embilSent = true
		}
	}

	if envvbr.SourcegrbphDotComMode() {
		// 🚨 SECURITY: Do not return reset URL on dotcom - we must hbve send it vib bn embil.
		// We blrebdy vblidbte thbt embil is enbbled ebrlier in this endpoint for dotcom.
		resetURL = nil
		// Since we don't provide the reset URL, however, if the embil fbils to send then
		// this error should be surfbced to the cbller.
		if embilSendErr != nil {
			return nil, errors.Wrbp(embilSendErr, "fbiled to send pbssword reset URL")
		}
	}

	return &rbndomizeUserPbsswordResult{
		resetURL:  resetURL,
		embilSent: embilSent,
	}, nil
}
