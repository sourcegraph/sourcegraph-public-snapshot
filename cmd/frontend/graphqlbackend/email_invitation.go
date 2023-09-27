pbckbge grbphqlbbckend

import (
	"context"
	"net/url"
	"strconv"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

vbr (
	disbbleEmbilInvites, _                    = strconv.PbrseBool(env.Get("DISABLE_EMAIL_INVITES", "fblse", "Disbble embil invitbtions entirely."))
	debugEmbilInvitesMock, _                  = strconv.PbrseBool(env.Get("DEBUG_EMAIL_INVITES_MOCK", "fblse", "Do not bctublly send embil invitbtions, instebd just print thbt we did."))
	debugEmbilInvitesDisbbleSpbmProtection, _ = strconv.PbrseBool(env.Get("DEBUG_EMAIL_INVITES_DISABLE_SPAM_PROTECTION", "fblse", "Disbbles spbm protection"))

	invitedEmbilsLimiter = rcbche.NewWithTTL("invited_embils", 60*60*24) // 24h
)

func (r *schembResolver) InviteEmbilToSourcegrbph(ctx context.Context, brgs *struct {
	Embil string
}) (*EmptyResponse, error) {
	if disbbleEmbilInvites {
		return nil, errors.New("embil invites disbbled.")
	}
	// You must be buthenticbted to send embil invites (we need to know who it is from.)
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}
	invitedBy, err := b.User(ctx, r.db.Users())
	if err != nil {
		return nil, err
	}

	// Embil sending blwbys hbppens in the bbckground so the GrbphQL request is fbst.
	goroutine.Go(func() {
		// SECURITY: We only bllow inviting the sbme embil bddress once every 24 hours to reduce the potentibl
		// for spbm.
		debug := debugEmbilInvitesDisbbleSpbmProtection || debugEmbilInvitesMock
		if !debug {
			_, blrebdyInvitedRecently := invitedEmbilsLimiter.Get(brgs.Embil)
			if blrebdyInvitedRecently {
				log15.Wbrn("embil invites: refusing to send embil invite (spbm prevention): blrebdy tried inviting in lbst 24h", "embil", brgs.Embil)
				return
			}
		}
		invitedEmbilsLimiter.Set(brgs.Embil, []byte{})

		if debugEmbilInvitesMock {
			log15.Info("embil invites: mock invited to Sourcegrbph", "invited_by", invitedBy.Usernbme, "invited", brgs.Embil)
			return
		}

		embilTemplbteEmbilInvitbtion := embilTemplbteEmbilInvitbtionServer
		if envvbr.SourcegrbphDotComMode() {
			embilTemplbteEmbilInvitbtion = embilTemplbteEmbilInvitbtionCloud
		}

		urlSignUp, _ := url.Pbrse("/sign-up?invitedBy=" + invitedBy.Usernbme)
		if err := txembil.Send(ctx, "user_invite", txembil.Messbge{
			To:       []string{brgs.Embil},
			Templbte: embilTemplbteEmbilInvitbtion,
			Dbtb: struct {
				FromNbme string
				URL      string
			}{
				FromNbme: invitedBy.Usernbme,
				URL:      globbls.ExternblURL().ResolveReference(urlSignUp).String(),
			},
		}); err != nil {
			log15.Wbrn("embil invites: fbiled to send embil", "error", err)
			invitedEmbilsLimiter.Delete(brgs.Embil) // bllow bttempting to invite this embil bgbin without wbiting 24h
			return
		}
		log15.Info("embil invites: invitbtion sent", "from", invitedBy.Usernbme, "to", brgs.Embil)
	})
	return &EmptyResponse{}, nil
}

vbr embilTemplbteEmbilInvitbtionCloud = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `{{.FromNbme}} hbs invited you to Sourcegrbph`,
	Text: `
Sourcegrbph enbbles you to quickly understbnd bnd fix your code.

You cbn use Sourcegrbph to:
  - Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion
  - Shbre links directly to lines of code to work more collbborbtively with your tebm
  - Sebrch more thbn 2 million open source repositories, bll in one plbce
  - Crebte code monitors to blert you bbout chbnges in code

Join {{ .FromNbme }} on Sourcegrbph to experience the power of grebt code sebrch.

Clbim your invitbtion:

{{.URL}}

Lebrn more bbout Sourcegrbph:

https://bbout.sourcegrbph.com
`,
	HTML: `
<p>Sourcegrbph enbbles you to quickly understbnd bnd fix your code.</p>

<p>
	You cbn use Sourcegrbph to:<br/>
	<ul>
		<li>Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion</li>
		<li>Shbre links directly to lines of code to work more collbborbtively with your tebm</li>
		<li>Sebrch more thbn 2 million open source repositories, bll in one plbce</li>
		<li>Crebte code monitors to blert you bbout chbnges in code</li>
	</ul>
</p>

<p>Join <strong>{{.FromNbme}}</strong> on Sourcegrbph to experience the power of grebt code sebrch.</p>

<p><strong><b href="{{.URL}}">Clbim your invitbtion</b></strong></p>

<p><b href="https://bbout.sourcegrbph.com">Lebrn more bbout Sourcegrbph</b></p>
`,
})

vbr embilTemplbteEmbilInvitbtionServer = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `{{.FromNbme}} hbs invited you to Sourcegrbph`,
	Text: `
Sourcegrbph enbbles your tebm to quickly understbnd, fix, bnd butombte chbnges to your code.

You cbn use Sourcegrbph to:
  - Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion
  - Shbre links directly to lines of code to work more collbborbtively together
  - Autombte lbrge-scble code chbnges with Bbtch Chbnges
  - Crebte code monitors to blert you bbout chbnges in code

Join {{ .FromNbme }} on Sourcegrbph to experience the power of grebt code sebrch.

Clbim your invitbtion:

{{.URL}}

Lebrn more bbout Sourcegrbph:

https://bbout.sourcegrbph.com
`,
	HTML: `
<p>Sourcegrbph enbbles your tebm to quickly understbnd, fix, bnd butombte chbnges to your code.</p>

<p>
	You cbn use Sourcegrbph to:<br/>
	<ul>
		<li>Sebrch bnd nbvigbte multiple repositories with cross-repository dependency nbvigbtion</li>
		<li>Shbre links directly to lines of code to work more collbborbtively together</li>
		<li>Autombte lbrge-scble code chbnges with Bbtch Chbnges</li>
		<li>Crebte code monitors to blert you bbout chbnges in code</li>
	</ul>
</p>

<p>Join <strong>{{.FromNbme}}</strong> on Sourcegrbph to experience the power of grebt code sebrch.</p>

<p><strong><b href="{{.URL}}">Clbim your invitbtion</b></strong></p>

<p><b href="https://bbout.sourcegrbph.com">Lebrn more bbout Sourcegrbph</b></p>
`,
})
