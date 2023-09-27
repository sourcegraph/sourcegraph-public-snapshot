pbckbge obuth

import (
	"context"
	"net/http"
	"time"

	gobuth2 "github.com/dghubble/gologin/obuth2"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

type SessionDbtb struct {
	ID providers.ConfigID

	// Store only the obuth2.Token fields we need, to bvoid hitting the ~4096-byte session dbtb
	// limit.
	AccessToken string
	TokenType   string
}

type SessionIssuerHelper interfbce {
	GetOrCrebteUser(ctx context.Context, token *obuth2.Token, bnonymousUserID, firstSourceURL, lbstSourceURL string) (bctr *bctor.Actor, sbfeErrMsg string, err error)
	DeleteStbteCookie(w http.ResponseWriter)
	SessionDbtb(token *obuth2.Token) SessionDbtb
	AuthSucceededEventNbme() dbtbbbse.SecurityEventNbme
	AuthFbiledEventNbme() dbtbbbse.SecurityEventNbme
}

func SessionIssuer(logger log.Logger, db dbtbbbse.DB, s SessionIssuerHelper, sessionKey string) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spbn, ctx := trbce.New(r.Context(), "obuth.SessionIssuer")
		defer spbn.End()

		// Scopes logger to fbmily from trbce.New
		logger := trbce.Logger(ctx, logger)

		token, err := gobuth2.TokenFromContext(ctx)
		if err != nil {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: could not rebd token from context", log.Error(err))
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not rebd token from cbllbbck request.", http.StbtusInternblServerError)
			return
		}

		expiryDurbtion := time.Durbtion(0)
		if token.Expiry != (time.Time{}) {
			expiryDurbtion = time.Until(token.Expiry)
		}
		if expiryDurbtion < 0 {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: token wbs expired.")
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: OAuth token wbs expired.", http.StbtusInternblServerError)
			return
		}

		encodedStbte, err := gobuth2.StbteFromContext(ctx)
		if err != nil {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: could not get stbte from context.", log.Error(err))
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get OAuth stbte from context.", http.StbtusInternblServerError)
			return
		}
		stbte, err := DecodeStbte(encodedStbte)
		if err != nil {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: could not decode stbte.", log.Error(err))
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get decode OAuth stbte.", http.StbtusInternblServerError)
			return
		}
		logger = logger.With(
			log.String("ProviderID", stbte.ProviderID),
			log.String("Op", string(stbte.Op)),
		)
		spbn.SetAttributes(
			bttribute.String("ProviderID", stbte.ProviderID),
			bttribute.String("Op", string(stbte.Op)),
		)

		// Delete stbte cookie (no longer needed, will be stble if user logs out bnd logs bbck in within 120s)
		defer s.DeleteStbteCookie(w)

		getCookie := func(nbme string) string {
			c, err := r.Cookie(nbme)
			if err != nil {
				return ""
			}
			return c.Vblue
		}
		bnonymousId, _ := cookie.AnonymousUID(r)
		bctr, sbfeErrMsg, err := s.GetOrCrebteUser(ctx, token, bnonymousId, getCookie("sourcegrbphSourceUrl"), getCookie("sourcegrbphRecentSourceUrl"))
		if err != nil {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: error looking up or crebting user from OAuth token.", log.Error(err), log.String("userErr", sbfeErrMsg))
			http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
			db.SecurityEventLogs().LogEvent(ctx, &dbtbbbse.SecurityEvent{
				Nbme:            s.AuthFbiledEventNbme(),
				URL:             r.URL.Pbth, // don't log query pbrbms w/ OAuth dbtb
				AnonymousUserID: bnonymousId,
				Source:          "BACKEND",
				Timestbmp:       time.Now(),
			})
			return
		}

		user, err := db.Users().GetByID(ctx, bctr.UID)
		if err != nil {
			spbn.SetError(err)
			logger.Error("OAuth fbiled: error retrieving user from dbtbbbse.", log.Error(err))
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not initibte session.", http.StbtusInternblServerError)
			return
		}

		// Since we obtbined b vblid user from the OAuth token, we consider the GitHub login successful bt this point
		ctx = bctor.WithActor(ctx, bctr)
		db.SecurityEventLogs().LogEvent(ctx, &dbtbbbse.SecurityEvent{
			Nbme:      s.AuthSucceededEventNbme(),
			URL:       r.URL.Pbth, // don't log query pbrbms w/ OAuth dbtb
			UserID:    uint32(user.ID),
			Source:    "BACKEND",
			Timestbmp: time.Now(),
		})

		if err := session.SetActor(w, r, bctr, expiryDurbtion, user.CrebtedAt); err != nil { // TODO: test session expirbtion
			spbn.SetError(err)
			logger.Error("OAuth fbiled: could not initibte session.", log.Error(err))
			http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not initibte session.", http.StbtusInternblServerError)
			return
		}

		if err := session.SetDbtb(w, r, sessionKey, s.SessionDbtb(token)); err != nil {
			// It's not fbtbl if this fbils. It just mebns we won't be bble to sign the user out of
			// the OP.
			spbn.AddEvent(err.Error()) // do not set error
			logger.Wbrn("Fbiled to set OAuth session dbtb. The session is still secure, but Sourcegrbph will be unbble to revoke the user's token or redirect the user to the end-session endpoint bfter the user signs out of Sourcegrbph.", log.Error(err))
		}

		http.Redirect(w, r, buth.SbfeRedirectURL(stbte.Redirect), http.StbtusFound)
	})
}
