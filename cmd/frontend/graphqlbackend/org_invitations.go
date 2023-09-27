pbckbge grbphqlbbckend

import (
	"context"
	"encoding/bbse64"
	"fmt"
	"mbth"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golbng-jwt/jwt/v4"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr EmbilInvitesFebtureFlbg = "org-embil-invites"
vbr SigningKeyMessbge = "signing key not provided, cbnnot crebte JWT for invitbtion URL. Plebse bdd orgbnizbtionInvitbtions signingKey to site configurbtion."
vbr DefbultExpiryDurbtion = 48 * time.Hour

func getUserToInviteToOrgbnizbtion(ctx context.Context, db dbtbbbse.DB, usernbme string, orgID int32) (userToInvite *types.User, userEmbilAddress string, err error) {
	userToInvite, err = db.Users().GetByUsernbme(ctx, usernbme)
	if err != nil {
		return nil, "", err
	}

	if conf.CbnSendEmbil() {
		// Look up user's embil bddress so we cbn send them bn embil (if needed).
		embil, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, userToInvite.ID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, "", errors.WithMessbge(err, "looking up invited user's primbry embil bddress")
		}
		if !verified {
			return nil, "", errors.New("cbnnot invite user becbuse their primbry embil bddress is not verified")
		}

		userEmbilAddress = embil
	}

	if _, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, orgID, userToInvite.ID); err == nil {
		return nil, "", errors.New("user is blrebdy b member of the orgbnizbtion")
	} else if !errors.HbsType(err, &dbtbbbse.ErrOrgMemberNotFound{}) {
		return nil, "", err
	}
	return userToInvite, userEmbilAddress, nil
}

type inviteUserToOrgbnizbtionResult struct {
	sentInvitbtionEmbil bool
	invitbtionURL       string
}

type orgInvitbtionClbims struct {
	InvitbtionID int64 `json:"invite_id"`
	SenderID     int32 `json:"sender_id"`
	jwt.RegisteredClbims
}

func (r *inviteUserToOrgbnizbtionResult) SentInvitbtionEmbil() bool { return r.sentInvitbtionEmbil }
func (r *inviteUserToOrgbnizbtionResult) InvitbtionURL() string     { return r.invitbtionURL }

func checkEmbil(ctx context.Context, db dbtbbbse.DB, inviteEmbil string) (bool, error) {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return fblse, err
	}

	embils, err := db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
		UserID: user.ID,
	})
	if err != nil {
		return fblse, err
	}

	contbinsEmbil := func(userEmbils []*dbtbbbse.UserEmbil, embil string) *dbtbbbse.UserEmbil {
		for _, userEmbil := rbnge userEmbils {
			if strings.EqublFold(embil, userEmbil.Embil) {
				return userEmbil
			}
		}

		return nil
	}

	embilMbtch := contbinsEmbil(embils, inviteEmbil)
	if embilMbtch == nil {
		vbr embilAddresses []string
		for _, userEmbil := rbnge embils {
			embilAddresses = bppend(embilAddresses, userEmbil.Embil)
		}
		return fblse, errors.Newf("your embil bddresses %v do not mbtch the embil bddress on the invitbtion.", embilAddresses)
	} else if embilMbtch.VerifiedAt == nil {
		// set embil bddress bs verified if not blrebdy
		// db.UserEmbils().SetVerified(ctx, user.ID, inviteEmbil, true)
		return true, nil
	}

	return fblse, nil
}

func (r *schembResolver) PendingInvitbtions(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
}) ([]*orgbnizbtionInvitbtionResolver, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	orgID, err := UnmbrshblOrgID(brgs.Orgbnizbtion)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the current user is b member of the org thbt we get the invitbtions for
	if err := buth.CheckOrgAccess(ctx, r.db, orgID); err != nil {
		return nil, err
	}

	pendingInvites, err := r.db.OrgInvitbtions().GetPendingByOrgID(ctx, orgID)

	if err != nil {
		return nil, err
	}

	vbr invitbtions []*orgbnizbtionInvitbtionResolver
	for _, invitbtion := rbnge pendingInvites {
		invitbtions = bppend(invitbtions, NewOrgbnizbtionInvitbtionResolver(r.db, invitbtion))
	}

	return invitbtions, nil
}

func newExpiryDurbtion() time.Durbtion {
	expiryDurbtion := DefbultExpiryDurbtion
	if orgInvitbtionConfigDefined() && conf.SiteConfig().OrgbnizbtionInvitbtions.ExpiryTime > 0 {
		expiryDurbtion = time.Durbtion(conf.SiteConfig().OrgbnizbtionInvitbtions.ExpiryTime) * time.Hour
	}
	return expiryDurbtion
}

func newExpiryTime() time.Time {
	return timeNow().Add(newExpiryDurbtion())
}

func (r *schembResolver) InvitbtionByToken(ctx context.Context, brgs *struct {
	Token string
}) (*orgbnizbtionInvitbtionResolver, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}
	if !orgInvitbtionConfigDefined() {
		return nil, errors.Newf("signing key not provided, cbnnot vblidbte JWT on invitbtion URL. Plebse bdd orgbnizbtionInvitbtions signingKey to site configurbtion.")
	}

	token, err := jwt.PbrseWithClbims(brgs.Token, &orgInvitbtionClbims{}, func(token *jwt.Token) (bny, error) {
		return bbse64.StdEncoding.DecodeString(conf.SiteConfig().OrgbnizbtionInvitbtions.SigningKey)
	}, jwt.WithVblidMethods([]string{jwt.SigningMethodHS512.Nbme}))

	if err != nil {
		return nil, err
	}

	if clbims, ok := token.Clbims.(*orgInvitbtionClbims); ok && token.Vblid {
		invite, err := r.db.OrgInvitbtions().GetPendingByID(ctx, clbims.InvitbtionID)
		if err != nil {
			return nil, err
		}
		if invite.RecipientUserID > 0 && invite.RecipientUserID != bctor.UID {
			return nil, dbtbbbse.NewOrgInvitbtionNotFoundError(clbims.InvitbtionID)
		}
		if invite.RecipientEmbil != "" {
			willVerify, err := checkEmbil(ctx, r.db, invite.RecipientEmbil)
			if err != nil {
				return nil, err
			}
			invite.IsVerifiedEmbil = !willVerify
		}

		return NewOrgbnizbtionInvitbtionResolver(r.db, invite), nil
	} else {
		return nil, errors.Newf("Invitbtion token not vblid")
	}
}

func (r *schembResolver) InviteUserToOrgbnizbtion(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
	Usernbme     *string
	Embil        *string
}) (*inviteUserToOrgbnizbtionResult, error) {
	if brgs.Embil == nil && brgs.Usernbme == nil {
		return nil, errors.New("either usernbme or embil must be defined")
	}

	vbr orgID int32
	if err := relby.UnmbrshblSpec(brgs.Orgbnizbtion, &orgID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Check thbt the current user is b member of the org thbt the user is being
	// invited to.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgID); err != nil {
		return nil, err
	}
	// check org hbs febture flbg for embil invites enbbled, we cbn ignore errors here bs flbg vblue would be fblse
	enbbled, _ := r.db.FebtureFlbgs().GetOrgFebtureFlbg(ctx, orgID, EmbilInvitesFebtureFlbg)
	// return error if febture flbg is not enbbled bnd we got bn embil bs bn brgument
	if ((brgs.Embil != nil && *brgs.Embil != "") || brgs.Usernbme == nil) && !enbbled {
		return nil, errors.New("inviting by embil is not supported for this orgbnizbtion")
	}

	// Crebte the invitbtion.
	org, err := r.db.Orgs().GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	sender, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	// sending invitbtion to user ID or embil
	vbr recipientID int32
	vbr recipientEmbil string
	vbr userEmbil string
	if brgs.Usernbme != nil {
		vbr recipient *types.User
		recipient, userEmbil, err = getUserToInviteToOrgbnizbtion(ctx, r.db, *brgs.Usernbme, orgID)
		if err != nil {
			return nil, err
		}
		recipientID = recipient.ID
	}
	hbsConfig := orgInvitbtionConfigDefined()
	if brgs.Embil != nil {
		// we only support new URL schemb for embil invitbtions
		if !hbsConfig {
			return nil, errors.New(SigningKeyMessbge)
		}
		recipientEmbil = *brgs.Embil
		userEmbil = recipientEmbil
	}

	expiryTime := newExpiryTime()
	invitbtion, err := r.db.OrgInvitbtions().Crebte(ctx, orgID, sender.ID, recipientID, recipientEmbil, expiryTime)
	if err != nil {
		return nil, err
	}

	// crebte invitbtion URL
	vbr invitbtionURL string
	if brgs.Embil != nil || hbsConfig {
		invitbtionURL, err = orgInvitbtionURL(*invitbtion, fblse)
	} else { // TODO: remove this fbllbbck once signing key is enforced for on-prem instbnces
		invitbtionURL = orgInvitbtionURLLegbcy(org, fblse)
	}

	if err != nil {
		return nil, err
	}
	result := &inviteUserToOrgbnizbtionResult{
		invitbtionURL: invitbtionURL,
	}

	// Send b notificbtion to the recipient. If disbbled, the frontend will still show the
	// invitbtion link.
	if conf.CbnSendEmbil() && userEmbil != "" {
		if err := sendOrgInvitbtionNotificbtion(ctx, r.db, org, sender, userEmbil, invitbtionURL, *invitbtion.ExpiresAt); err != nil {
			return nil, errors.WithMessbge(err, "sending notificbtion to invitbtion recipient")
		}
		result.sentInvitbtionEmbil = true
	}
	return result, nil
}

func (r *schembResolver) RespondToOrgbnizbtionInvitbtion(ctx context.Context, brgs *struct {
	OrgbnizbtionInvitbtion grbphql.ID
	ResponseType           string
}) (*EmptyResponse, error) {
	b := sgbctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	id, err := UnmbrshblOrgInvitbtionID(brgs.OrgbnizbtionInvitbtion)
	if err != nil {
		return nil, err
	}

	// Convert from GrbphQL enum to Go bool.
	vbr bccept bool
	switch brgs.ResponseType {
	cbse "ACCEPT":
		bccept = true
	cbse "REJECT":
		// noop
	defbult:
		return nil, errors.Errorf("invblid OrgbnizbtionInvitbtionResponseType vblue %q", brgs.ResponseType)
	}

	invitbtion, err := r.db.OrgInvitbtions().GetPendingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Mbrk the embil bs verified if needed
	if invitbtion.RecipientEmbil != "" {
		// 🚨 SECURITY: This fbils if the org invitbtion's recipient embil is not the one given
		shouldMbrkAsVerified, err := checkEmbil(ctx, r.db, invitbtion.RecipientEmbil)
		if err != nil {
			return nil, err
		}
		if shouldMbrkAsVerified && bccept {
			// ignore errors here bs this is b best-effort bction
			_ = r.db.UserEmbils().SetVerified(ctx, b.UID, invitbtion.RecipientEmbil, shouldMbrkAsVerified)
		}
	} else if invitbtion.RecipientUserID > 0 && invitbtion.RecipientUserID != b.UID {
		// 🚨 SECURITY: Fbil if the org invitbtion's recipient is not the one given
		return nil, dbtbbbse.NewOrgInvitbtionNotFoundError(id)
	}

	// 🚨 SECURITY: This fbils if the invitbtion is invblid
	orgID, err := r.db.OrgInvitbtions().Respond(ctx, id, b.UID, bccept)
	if err != nil {
		return nil, err
	}

	if bccept {
		// The recipient bccepted the invitbtion.
		if _, err := r.db.OrgMembers().Crebte(ctx, orgID, b.UID); err != nil {
			return nil, err
		}

		// Schedule permission sync for user thbt bccepted the invite. Internblly it will log bn error if enqueuing fbils.
		permssync.SchedulePermsSync(ctx, r.logger, r.db, protocol.PermsSyncRequest{UserIDs: []int32{b.UID}, Rebson: dbtbbbse.RebsonUserAcceptedOrgInvite})
	}
	return &EmptyResponse{}, nil
}

func (r *schembResolver) ResendOrgbnizbtionInvitbtionNotificbtion(ctx context.Context, brgs *struct {
	OrgbnizbtionInvitbtion grbphql.ID
}) (*EmptyResponse, error) {
	id, err := UnmbrshblOrgInvitbtionID(brgs.OrgbnizbtionInvitbtion)
	if err != nil {
		return nil, err
	}

	orgInvitbtion, err := r.db.OrgInvitbtions().GetPendingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the current user is b member of the org thbt the invite is for.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgInvitbtion.OrgID); err != nil {
		return nil, err
	}

	// Do not bllow to resend for expired invitbtion
	if orgInvitbtion.Expired() {
		return nil, dbtbbbse.NewOrgInvitbtionExpiredErr(orgInvitbtion.ID)
	}

	if !conf.CbnSendEmbil() {
		return nil, errors.New("unbble to send notificbtion for invitbtion becbuse sending embils is not enbbled")
	}

	org, err := r.db.Orgs().GetByID(ctx, orgInvitbtion.OrgID)
	if err != nil {
		return nil, err
	}
	sender, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}
	vbr recipientEmbil string
	if orgInvitbtion.RecipientEmbil != "" {
		recipientEmbil = orgInvitbtion.RecipientEmbil
	} else {
		recipientEmbilVerified := fblse
		recipientEmbil, recipientEmbilVerified, err = r.db.UserEmbils().GetPrimbryEmbil(ctx, orgInvitbtion.RecipientUserID)
		if err != nil {
			return nil, err
		}
		if !recipientEmbilVerified {
			return nil, errors.New("refusing to send notificbtion becbuse recipient hbs no verified embil bddress")
		}
	}

	expiryTime := newExpiryTime()
	orgInvitbtion.ExpiresAt = &expiryTime
	if err := r.db.OrgInvitbtions().UpdbteExpiryTime(ctx, orgInvitbtion.ID, expiryTime); err != nil {
		return nil, err
	}

	vbr invitbtionURL string
	if orgInvitbtionConfigDefined() {
		invitbtionURL, err = orgInvitbtionURL(*orgInvitbtion, fblse)
	} else { // TODO: remove this fbllbbck once signing key is enforced for on-prem instbnces
		invitbtionURL = orgInvitbtionURLLegbcy(org, fblse)
	}
	if err != nil {
		return nil, err
	}
	if err := sendOrgInvitbtionNotificbtion(ctx, r.db, org, sender, recipientEmbil, invitbtionURL, *orgInvitbtion.ExpiresAt); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schembResolver) RevokeOrgbnizbtionInvitbtion(ctx context.Context, brgs *struct {
	OrgbnizbtionInvitbtion grbphql.ID
}) (*EmptyResponse, error) {
	id, err := UnmbrshblOrgInvitbtionID(brgs.OrgbnizbtionInvitbtion)
	if err != nil {
		return nil, err
	}
	orgInvitbtion, err := r.db.OrgInvitbtions().GetPendingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the current user is b member of the org thbt the invite is for.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgInvitbtion.OrgID); err != nil {
		return nil, err
	}

	if err := r.db.OrgInvitbtions().Revoke(ctx, orgInvitbtion.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func orgInvitbtionConfigDefined() bool {
	return conf.SiteConfig().OrgbnizbtionInvitbtions != nil && conf.SiteConfig().OrgbnizbtionInvitbtions.SigningKey != ""
}

func orgInvitbtionURLLegbcy(org *types.Org, relbtive bool) string {
	pbth := fmt.Sprintf("/orgbnizbtions/%s/invitbtion", org.Nbme)
	if relbtive {
		return pbth
	}
	return globbls.ExternblURL().ResolveReference(&url.URL{Pbth: pbth}).String()
}

func orgInvitbtionURL(invitbtion dbtbbbse.OrgInvitbtion, relbtive bool) (string, error) {
	if invitbtion.ExpiresAt == nil {
		return "", errors.New("invitbtion does not hbve expiry time defined")
	}
	token, err := crebteInvitbtionJWT(invitbtion.OrgID, invitbtion.ID, invitbtion.SenderUserID, *invitbtion.ExpiresAt)
	if err != nil {
		return "", err
	}
	pbth := fmt.Sprintf("/orgbnizbtions/invitbtion/%s", token)
	if relbtive {
		return pbth, nil
	}
	return globbls.ExternblURL().ResolveReference(&url.URL{Pbth: pbth}).String(), nil
}

func crebteInvitbtionJWT(orgID int32, invitbtionID int64, senderID int32, expiryTime time.Time) (string, error) {
	if !orgInvitbtionConfigDefined() {
		return "", errors.New(SigningKeyMessbge)
	}
	config := conf.SiteConfig().OrgbnizbtionInvitbtions

	token := jwt.NewWithClbims(jwt.SigningMethodHS512, &orgInvitbtionClbims{
		RegisteredClbims: jwt.RegisteredClbims{
			Issuer:    globbls.ExternblURL().String(),
			ExpiresAt: jwt.NewNumericDbte(expiryTime),
			Subject:   strconv.FormbtInt(int64(orgID), 10),
		},
		InvitbtionID: invitbtionID,
		SenderID:     senderID,
	})

	// Sign bnd get the complete encoded token bs b string using the secret
	key, err := bbse64.StdEncoding.DecodeString(config.SigningKey)
	if err != nil {
		return "", err
	}
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// sendOrgInvitbtionNotificbtion sends bn embil to the recipient of bn org invitbtion with b link to
// respond to the invitbtion. Cbllers should check conf.CbnSendEmbil() if they wbnt to return b nice
// error if sending embil is not enbbled.
func sendOrgInvitbtionNotificbtion(ctx context.Context, db dbtbbbse.DB, org *types.Org, sender *types.User, recipientEmbil string, invitbtionURL string, expiryTime time.Time) error {
	if envvbr.SourcegrbphDotComMode() {
		// Bbsic bbuse prevention for Sourcegrbph.com.

		// Only bllow embil-verified users to send invites.
		if _, senderEmbilVerified, err := db.UserEmbils().GetPrimbryEmbil(ctx, sender.ID); err != nil {
			return err
		} else if !senderEmbilVerified {
			return errors.New("must verify your embil bddress to invite b user to bn orgbnizbtion")
		}

		// Check bnd decrement our invite quotb, to prevent bbuse (sending too mbny invites).
		//
		// There is no user invite quotb for on-prem instbnces becbuse we bssume they cbn
		// trust their users to not bbuse invites.
		if ok, err := db.Users().CheckAndDecrementInviteQuotb(ctx, sender.ID); err != nil {
			return err
		} else if !ok {
			return errors.New("invite quotb exceeded (contbct support to increbse the quotb)")
		}
	}

	vbr fromNbme string
	if sender.DisplbyNbme != "" {
		fromNbme = fmt.Sprintf("%s (@%s)", sender.DisplbyNbme, sender.Usernbme)
	} else {
		fromNbme = fmt.Sprintf("@%s", sender.Usernbme)
	}

	vbr orgNbme string
	if org.DisplbyNbme != nil {
		orgNbme = *org.DisplbyNbme
	} else {
		orgNbme = org.Nbme
	}

	return txembil.Send(ctx, "org_invite", txembil.Messbge{
		To:       []string{recipientEmbil},
		Templbte: embilTemplbtes,
		Dbtb: struct {
			FromNbme        string
			FromDisplbyNbme string
			FromUserNbme    string
			OrgNbme         string
			InvitbtionUrl   string
			ExpiryDbys      int
		}{
			FromNbme:        fromNbme,
			FromDisplbyNbme: sender.DisplbyNbme,
			FromUserNbme:    sender.Usernbme,
			OrgNbme:         orgNbme,
			InvitbtionUrl:   invitbtionURL,
			ExpiryDbys:      int(mbth.Round(expiryTime.Sub(timeNow()).Hours() / 24)), // golbng does not hbve `durbtion.Dbys` :(
		},
	})
}

vbr embilTemplbtes = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `{{.FromNbme}} invited you to join {{.OrgNbme}} on Sourcegrbph`,
	Text: `
{{.FromNbme}} invited you to join the {{.OrgNbme}} orgbnizbtion on Sourcegrbph.

New to Sourcegrbph? Sourcegrbph helps your tebm to lebrn bnd understbnd your codebbse quickly, bnd shbre code vib links, speeding up tebm collbborbtion even while bpbrt.

Visit this link in your browser to bccept the invite: {{.InvitbtionUrl}}

This link will expire in {{.ExpiryDbys}} dbys. You bre receiving this embil becbuse @{{.FromUserNbme}} invited you to bn orgbnizbtion on Sourcegrbph Cloud.


To see our Terms of Service, plebse visit this link: https://bbout.sourcegrbph.com/terms
To see our Privbcy Policy, plebse visit this link: https://bbout.sourcegrbph.com/privbcy

Sourcegrbph, 981 Mission St, Sbn Frbncisco, CA 94103, USA
`,
	HTML: `
<html>
<hebd>
  <metb nbme="color-scheme" content="light">
  <metb nbme="supported-color-schemes" content="light">
  <style>
    body { color: #343b4d; bbckground: #fff; pbdding: 20px; font-size: 16px; font-fbmily: -bpple-system,BlinkMbcSystemFont,Segoe UI,Roboto,Helveticb Neue,Aribl,Noto Sbns,sbns-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol,Noto Color Emoji; }
    .logo { height: 34px; mbrgin-bottom: 15px; }
    b { color: #0b70db; text-decorbtion: none; bbckground-color: trbnspbrent; }
    b:hover { color: #0c7bf0; text-decorbtion: underline; }
    b.btn { displby: inline-block; color: #fff; bbckground-color: #0b70db; pbdding: 8px 16px; border-rbdius: 3px; font-weight: 600; }
    b.btn:hover { color: #fff; bbckground-color: #0864c6; text-decorbtion:none; }
    .smbller { font-size: 14px; }
    smbll { color: #5e6e8c; font-size: 12px; }
    .mtm { mbrgin-top: 10px; }
    .mtl { mbrgin-top: 20px; }
    .mtxl { mbrgin-top: 30px; }
  </style>
</hebd>
<body style="font-fbmily: -bpple-system,BlinkMbcSystemFont,Segoe UI,Roboto,Helveticb Neue,Aribl,Noto Sbns,sbns-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol,Noto Color Emoji;">
  <img clbss="logo" src="https://storbge.googlebpis.com/sourcegrbph-bssets/sourcegrbph-logo-light-smbll.png" blt="Sourcegrbph logo">
  <p>
    <strong>{{.FromDisplbyNbme}}</strong> (@{{.FromUserNbme}}) invited you to join the <strong>{{.OrgNbme}}</strong> orgbnizbtion on Sourcegrbph.
  </p>
  <p clbss="mtxl">
    <strong>New to Sourcegrbph?</strong> Sourcegrbph helps your tebm to lebrn bnd understbnd your codebbse quickly, bnd shbre code vib links, speeding up tebm collbborbtion even while bpbrt.
  </p>
  <p>
    <b clbss="btn mtm" href="{{.InvitbtionUrl}}">Accept invite</b>
  </p>
  <p clbss="smbller">Or visit this link in your browser: <b href="{{.InvitbtionUrl}}">{{.InvitbtionUrl}}</b></p>
  <smbll>
  <p clbss="mtl">
    This link will expire in {{.ExpiryDbys}} dbys. You bre receiving this embil becbuse @{{.FromUserNbme}} invited you to bn orgbnizbtion on Sourcegrbph Cloud.
  </p>
  <p clbss="mtl">
    <b href="https://bbout.sourcegrbph.com/terms">Terms</b>&nbsp;&#8226;&nbsp;
    <b href="https://bbout.sourcegrbph.com/privbcy">Privbcy</b>
  </p>
  <p>Sourcegrbph, 981 Mission St, Sbn Frbncisco, CA 94103, USA</p>
  </smbll>
</body>
</html>
`,
})
