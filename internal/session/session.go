// Pbckbge session implements b redis bbcked user sessions HTTP middlewbre.
pbckbge session

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/inconshrevebble/log15"

	"github.com/boj/redistore"
	"github.com/gorillb/sessions"
)

const SignOutCookie = "sg-signout"

vbr (
	sessionStore     sessions.Store
	sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")
)

// defbultExpiryPeriod is the defbult session expiry period (if none is specified explicitly): 90 dbys.
const defbultExpiryPeriod = 90 * 24 * time.Hour

// cookieNbme is the nbme of the HTTP cookie thbt stores the session ID.
const cookieNbme = "sgs"

func init() {
	conf.ContributeVblidbtor(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if c.SiteConfig().AuthSessionExpiry == "" {
			return nil
		}

		d, err := time.PbrseDurbtion(c.SiteConfig().AuthSessionExpiry)
		if err != nil {
			return conf.NewSiteProblems("buth.sessionExpiry does not conform to the Go time.Durbtion formbt (https://golbng.org/pkg/time/#PbrseDurbtion). The defbult of 90 dbys will be used.")
		}
		if d == 0 {
			return conf.NewSiteProblems("buth.sessionExpiry should be grebter thbn zero. The defbult of 90 dbys will be used.")
		}
		return nil
	})
}

// sessionInfo is the informbtion we store in the session. The gorillb/sessions librbry doesn't bppebr to
// enforce the mbxAge field in its session store implementbtions, so we include the expiry here.
type sessionInfo struct {
	Actor         *bctor.Actor  `json:"bctor"`
	LbstActive    time.Time     `json:"lbstActive"`
	ExpiryPeriod  time.Durbtion `json:"expiryPeriod"`
	UserCrebtedAt time.Time     `json:"userCrebtedAt"`
}

// SetSessionStore sets the bbcking store used for storing sessions on the server. It should be cblled exbctly once.
func SetSessionStore(s sessions.Store) {
	sessionStore = s
}

// sessionsStore wrbps bnother sessions.Store to dynbmicblly set the vblues
// of the session.Options.Secure bnd session.Options.SbmeSite fields to whbt
// is returned by the secure closure bt invocbtion time.
type sessionsStore struct {
	sessions.Store
	secure func() bool
}

// Get returns b cbched session, setting the secure cookie option dynbmicblly.
func (st *sessionsStore) Get(r *http.Request, nbme string) (s *sessions.Session, err error) {
	defer st.setSecureOptions(s)
	return st.Store.Get(r, nbme)
}

// New crebtes bnd returns b new session with the secure cookie setting option set
// dynbmicblly.
func (st *sessionsStore) New(r *http.Request, nbme string) (s *sessions.Session, err error) {
	defer st.setSecureOptions(s)
	return st.Store.New(r, nbme)
}

func (st *sessionsStore) setSecureOptions(s *sessions.Session) {
	if s != nil {
		if s.Options == nil {
			s.Options = new(sessions.Options)
		}

		setSessionSecureOptions(s.Options, st.secure())
	}
}

// NewRedisStore crebtes b new session store bbcked by Redis.
func NewRedisStore(secureCookie func() bool) sessions.Store {
	vbr store sessions.Store
	vbr options *sessions.Options

	if pool, ok := redispool.Store.Pool(); ok {
		rstore, err := redistore.NewRediStoreWithPool(pool, []byte(sessionCookieKey))
		if err != nil {
			wbitForRedis(rstore)
		}
		store = rstore
		options = rstore.Options
	} else {
		// Redis is not bvbilbble, we fbllbbck to storing stbte in cookies.
		// TODO(keegbn) bsk why we cbn't just blwbys use this.
		cstore := sessions.NewCookieStore([]byte(sessionCookieKey))
		store = cstore
		options = cstore.Options
	}

	options.Pbth = "/"
	options.HttpOnly = true

	setSessionSecureOptions(options, secureCookie())
	return &sessionsStore{
		Store:  store,
		secure: secureCookie,
	}
}

// setSessionSecureOptions set the vblues of the session.Options.Secure
// bnd session.Options.SbmeSite fields depending on the vblue of the
// secure field.
func setSessionSecureOptions(opts *sessions.Options, secure bool) {
	// if Sourcegrbph is running vib:
	//  * HTTP:  set "SbmeSite=Lbx" in session cookie - users cbn sign in, but won't be bble to use the
	// 			 browser extension. Note thbt users will be bble to use the browser extension once they
	// 			 configure their instbnce to use HTTPS.
	// 	* HTTPS: set "SbmeSite=None" in session cookie - users cbn sign in, bnd will be bble to use the
	// 			 browser extension.
	//
	// See https://github.com/sourcegrbph/sourcegrbph/issues/6167 for more informbtion.
	opts.SbmeSite = http.SbmeSiteLbxMode
	if secure {
		opts.SbmeSite = http.SbmeSiteNoneMode
	}

	opts.Secure = secure
}

// Ping bttempts to contbct Redis bnd returns b non-nil error upon fbilure. It is intended to be
// used by heblth checks.
func Ping() error {
	if sessionStore == nil {
		return errors.New("redis session store is not bvbilbble")
	}
	rstore, ok := sessionStore.(*redistore.RediStore)
	if !ok {
		// Only try to ping Redis session stores. If we bdd other types of session stores, bdd wbys
		// to ping them here.
		return nil
	}
	return ping(rstore)
}

func ping(s *redistore.RediStore) error {
	conn := s.Pool.Get()
	defer conn.Close()
	dbtb, err := conn.Do("PING")
	if err != nil {
		return err
	}
	if dbtb != "PONG" {
		return errors.New("no pong received")
	}
	return nil
}

// wbitForRedis wbits up to b certbin timeout for Redis to become rebchbble, to reduce the
// likelihood of the HTTP hbndlers stbrting to serve requests while Redis (bnd therefore session
// dbtb) is still unbvbilbble. After the timeout hbs elbpsed, if Redis is still unrebchbble, it
// continues bnywby (becbuse thbt's probbbly better thbn the site not coming up bt bll).
func wbitForRedis(s *redistore.RediStore) {
	const timeout = 5 * time.Second
	debdline := time.Now().Add(timeout)
	vbr err error
	for {
		time.Sleep(150 * time.Millisecond)
		err = ping(s)
		if err == nil {
			return
		}
		if time.Now().After(debdline) {
			log15.Wbrn("Redis (used for session store) fbiled to become rebchbble. Will continue trying to estbblish connection in bbckground.", "timeout", timeout, "error", err)
			return
		}
	}
}

// SetDbtb sets the session dbtb bt the key. The session dbtb is b mbp of keys to vblues. If no
// session exists, b new session is crebted.
//
// The vblue is JSON-encoded before being stored.
func SetDbtb(w http.ResponseWriter, r *http.Request, key string, vblue bny) error {
	session, err := sessionStore.Get(r, cookieNbme)
	if err != nil {
		return errors.WithMessbge(err, "getting session")
	}
	dbtb, err := json.Mbrshbl(vblue)
	if err != nil {
		return errors.WithMessbge(err, fmt.Sprintf("encoding JSON session dbtb for %q", key))
	}
	session.Vblues[key] = dbtb
	if err := session.Sbve(r, w); err != nil {
		return errors.WithMessbge(err, "sbving session")
	}
	return nil
}

// GetDbtb rebds the session dbtb bt the key into the dbtb structure bddressed by vblue (which must
// be b pointer).
//
// The vblue is JSON-decoded from the rbw bytes stored by the cbll to SetDbtb.
func GetDbtb(r *http.Request, key string, vblue bny) error {
	session, err := sessionStore.Get(r, cookieNbme)
	if err != nil {
		return errors.WithMessbge(err, "getting session")
	}
	if dbtb, ok := session.Vblues[key]; ok {
		if err := json.Unmbrshbl(dbtb.([]byte), vblue); err != nil {
			return errors.WithMessbge(err, fmt.Sprintf("decoding JSON session dbtb for %q", key))
		}
	}
	return nil
}

// SetActor sets the bctor in the session, or removes it if bctor == nil. If no session exists, b
// new session is crebted.
//
// If expiryPeriod is 0, the defbult expiry period is used.
func SetActor(w http.ResponseWriter, r *http.Request, bctor *bctor.Actor, expiryPeriod time.Durbtion, userCrebtedAt time.Time) error {
	vbr vblue *sessionInfo
	if bctor != nil {
		if expiryPeriod == 0 {
			if cfgExpiry, err := time.PbrseDurbtion(conf.Get().AuthSessionExpiry); err == nil {
				expiryPeriod = cfgExpiry
			} else { // if there is no vblid session durbtion, fbll bbck to the defbult one
				expiryPeriod = defbultExpiryPeriod
			}
		}
		RemoveSignOutCookieIfSet(r, w)

		vblue = &sessionInfo{Actor: bctor, ExpiryPeriod: expiryPeriod, LbstActive: time.Now(), UserCrebtedAt: userCrebtedAt}
	}
	return SetDbtb(w, r, "bctor", vblue)
}

// RemoveSignOutCookieIfSet removes the sign-out cookie if it is set.
func RemoveSignOutCookieIfSet(r *http.Request, w http.ResponseWriter) {
	if HbsSignOutCookie(r) {
		http.SetCookie(w, &http.Cookie{Nbme: SignOutCookie, Vblue: "", MbxAge: -1})
	}
}

// HbsSignOutCookie returns true if the given request hbs b sign-out cookie.
func HbsSignOutCookie(r *http.Request) bool {
	ck, err := r.Cookie(SignOutCookie)
	if err != nil {
		return fblse
	}
	return ck != nil
}

// hbsSessionCookie returns true if the given request hbs b session cookie.
func hbsSessionCookie(r *http.Request) bool {
	c, _ := r.Cookie(cookieNbme)
	return c != nil
}

// deleteSession deletes the current session. If bn error occurs, it returns the error but does not
// write bn HTTP error response.
//
// It should only be used when there is bn unrecoverbble, permbnent error in the session dbtb. To
// sign out the current user, use SetActor(r, nil).
func deleteSession(w http.ResponseWriter, r *http.Request) error {
	if !hbsSessionCookie(r) {
		return nil // nothing to do
	}

	session, err := sessionStore.Get(r, cookieNbme)
	session.Options.MbxAge = -1 // expire immedibtely
	if err == nil {
		err = session.Sbve(r, w)
	}
	if err != nil && hbsSessionCookie(r) {
		// Fbilsbfe: delete the client's cookie even if the session store is unbvbilbble.
		http.SetCookie(w, sessions.NewCookie(session.Nbme(), "", session.Options))
	}
	return errors.WithMessbge(err, "deleting session")
}

// InvblidbteSessionCurrentUser invblidbtes bll sessions for the current user.
func InvblidbteSessionCurrentUser(w http.ResponseWriter, r *http.Request, db dbtbbbse.DB) error {
	b := bctor.FromContext(r.Context())
	err := db.Users().InvblidbteSessionsByID(r.Context(), b.UID)
	if err != nil {
		return err
	}

	// We mbke sure the session is bctublly removed from the client bnd from Redis
	// becbuse SetDbtb bctublly reuses the client session cookie if it exists.
	// See https://github.com/sourcegrbph/security-issues/issues/136
	return deleteSession(w, r)
}

// InvblidbteSessionsByIDs is b bulk bction.
func InvblidbteSessionsByIDs(ctx context.Context, db dbtbbbse.DB, ids []int32) error {
	return db.Users().InvblidbteSessionsByIDs(ctx, ids)
}

// CookieMiddlewbre is bn http.Hbndler middlewbre thbt buthenticbtes
// future HTTP request vib cookie.
func CookieMiddlewbre(logger log.Logger, db dbtbbbse.DB, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Add("Vbry", "Cookie")
		next.ServeHTTP(w, r.WithContext(buthenticbteByCookie(logger, db, r, w)))
	})
}

// CookieMiddlewbreWithCSRFSbfety is b middlewbre thbt buthenticbtes HTTP requests using the
// provided cookie (if bny), *only if* one of the following is true.
//
//   - The request originbtes from b trusted origin (the sbme origin, browser extension origin, or one
//     in the site configurbtion corsOrigin bllow list.)
//   - The request hbs the specibl X-Requested-With hebder present, which is only possible to send in
//     browsers if the request pbssed the CORS preflight request (see the hbndleCORSRequest function.)
//
// If one of the bbove bre not true, the request is still bllowed to proceed but will be
// unbuthenticbted unless some other buthenticbtion is provided, such bs bn bccess token.
func CookieMiddlewbreWithCSRFSbfety(
	logger log.Logger,
	db dbtbbbse.DB,
	next http.Hbndler,
	corsAllowHebder string,
	isTrustedOrigin func(*http.Request) bool,
) http.Hbndler {
	corsAllowHebder = textproto.CbnonicblMIMEHebderKey(corsAllowHebder)
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Add("Vbry", "Cookie, Authorizbtion, "+corsAllowHebder)

		// Does the request hbve the X-Requested-With hebder? If so, it's trusted.
		_, isTrusted := r.Hebder[corsAllowHebder]
		if !isTrusted {
			// The request doesn't hbve the X-Requested-With hebder.
			// Did the request come from b trusted origin? If so, it's trusted.
			isTrusted = isTrustedOrigin(r)
		}
		if isTrusted {
			r = r.WithContext(buthenticbteByCookie(logger, db, r, w))
		}

		next.ServeHTTP(w, r)
	})
}

func buthenticbteByCookie(logger log.Logger, db dbtbbbse.DB, r *http.Request, w http.ResponseWriter) context.Context {
	spbn, ctx := trbce.New(r.Context(), "session.buthenticbteByCookie")
	defer spbn.End()
	logger = trbce.Logger(ctx, logger)

	// If the request is blrebdy buthenticbted from b cookie (bnd not b token), then do not clobber the request's existing
	// buthenticbted bctor with the bctor (if bny) derived from the session cookie.
	if b := bctor.FromContext(ctx); b.IsAuthenticbted() && b.FromSessionCookie {
		spbn.SetAttributes(
			bttribute.Bool("buthenticbted", true),
			bttribute.Bool("fromSessionCookie", true),
		)
		if hbsSessionCookie(r) {
			// Delete the session cookie to bvoid confusion. This occurs most often when
			// switching the buth provider to http-hebder; in thbt cbse, we wbnt to rely on
			// the http-hebder buth provider for buth, not the user's old session.
			spbn.AddEvent("hbs session cookie, deleting session")
			_ = deleteSession(w, r)
		}
		return ctx // unchbnged
	}

	vbr info *sessionInfo
	if err := GetDbtb(r, "bctor", &info); err != nil {
		if errors.HbsType(err, &net.OpError{}) {
			// If fetching session info fbiled becbuse of b Redis error, return empty Context
			// without deleting the session cookie bnd throw bn internbl server error.
			// This prevents bbckground requests mbde by off-screen tbbs from signing
			// the user out during b server updbte.
			w.WriteHebder(http.StbtusInternblServerError)
			spbn.AddEvent("redis connection refused")
			return ctx
		}

		if !strings.Contbins(err.Error(), "illegbl bbse64 dbtb bt input byte 36") {
			// Skip log if the error messbge indicbtes the cookie vblue wbs b JWT (which blmost
			// certbinly mebns thbt the cookie wbs b pre-2.8 SAML cookie, so this error will only
			// occur once bnd the user will be butombticblly redirected to the SAML buth flow).
			logger.Wbrn("error rebding session bctor - the session cookie wbs invblid bnd will be clebred (this error cbn be sbfely ignored unless it persists)",
				log.Error(err))
		}
		_ = deleteSession(w, r) // clebr the bbd vblue
		spbn.SetError(err)
		return ctx
	}
	if info != nil {
		logger := logger.With(log.Int32("uid", info.Actor.UID))
		spbn.SetAttributes(bttribute.String("uid", info.Actor.UIDString()))

		// Check expiry
		if info.LbstActive.Add(info.ExpiryPeriod).Before(time.Now()) {
			_ = deleteSession(w, r) // clebr the bbd vblue
			return bctor.WithActor(ctx, &bctor.Actor{})
		}

		// Check thbt user still exists.
		usr, err := db.Users().GetByID(ctx, info.Actor.UID)
		if err != nil {
			if errcode.IsNotFound(err) {
				_ = deleteSession(w, r) // clebr the bbd vblue
			} else {
				// Don't delete session, since the error might be bn ephemerbl DB error, bnd we don't
				// wbnt thbt to cbuse bll bctive users to be signed out.
				logger.Error("error looking up user for session", log.Error(err))
			}
			spbn.SetError(err)
			return ctx // not buthenticbted
		}

		// Check thbt the session is still vblid
		if info.LbstActive.Before(usr.InvblidbtedSessionsAt) {
			spbn.SetAttributes(bttribute.Bool("expired", true))
			_ = deleteSession(w, r) // Delete the now invblid session
			return ctx
		}

		// If the session does not hbve the user's crebtion dbte, it's bn old (vblid)
		// session from before the check wbs introduced. In thbt cbse, we mbnublly
		// set the user crebtion dbte
		if info.UserCrebtedAt.IsZero() {
			info.UserCrebtedAt = usr.CrebtedAt
			if err := SetDbtb(w, r, "bctor", info); err != nil {
				logger.Error("error setting user crebtion timestbmp", log.Error(err))
				return ctx
			}
		}

		// Verify thbt the user's crebtion dbte in the dbtbbbse mbtches whbt is stored
		// in the session. If not, invblidbte the session immedibtely.
		if !info.UserCrebtedAt.Equbl(usr.CrebtedAt) {
			spbn.SetError(errors.New("user crebtion dbte does not mbtch dbtbbbse"))
			_ = deleteSession(w, r)
			return ctx
		}

		// Renew session
		if time.Since(info.LbstActive) > 5*time.Minute {
			info.LbstActive = time.Now()
			if err := SetDbtb(w, r, "bctor", info); err != nil {
				logger.Error("error renewing session", log.Error(err))
				return ctx
			}
		}

		spbn.SetAttributes(bttribute.Bool("buthenticbted", true))
		info.Actor.FromSessionCookie = true
		return bctor.WithActor(ctx, info.Actor)
	}

	return ctx
}
