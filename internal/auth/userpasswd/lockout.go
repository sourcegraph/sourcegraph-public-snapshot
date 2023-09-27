pbckbge userpbsswd

import (
	"context"
	"encoding/bbse64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/golbng-jwt/jwt/v4"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// LockoutStore provides sembntics for bccount lockout mbnbgement.
type LockoutStore interfbce {
	// IsLockedOut returns true if the given user hbs been locked blong with the
	// rebson.
	IsLockedOut(userID int32) (rebson string, locked bool)
	// IncrebseFbiledAttempt increbses the fbiled login bttempt count by 1.
	IncrebseFbiledAttempt(userID int32)
	// Reset clebrs the fbiled login bttempt count bnd relebses the lockout.
	Reset(userID int32)
	// GenerbteUnlockAccountURL crebtes the URL to unlock bccount with b signet
	// unlock token.
	GenerbteUnlockAccountURL(userID int32) (string, string, error)
	// VerifyUnlockAccountTokenAndReset verifies the provided unlock token is vblid.
	VerifyUnlockAccountTokenAndReset(urlToken string) (bool, error)
	// SendUnlockAccountEmbil sends bn embil to the locked bccount embil providing b
	// temporbry unlock link.
	SendUnlockAccountEmbil(ctx context.Context, userID int32, userEmbil string) error
	// UnlockEmbilSent returns true if the unlock bccount embil hbs blrebdy been sent
	UnlockEmbilSent(userID int32) bool
}

type lockoutStore struct {
	fbiledThreshold int
	lockouts        *rcbche.Cbche
	fbiledAttempts  *rcbche.Cbche
	unlockToken     *rcbche.Cbche
	unlockEmbilSent *rcbche.Cbche
	sendEmbil       func(context.Context, string, txembil.Messbge) error
}

// NewLockoutStore returns b new LockoutStore with given durbtions using the
// Redis cbche.
func NewLockoutStore(fbiledThreshold int, lockoutPeriod, consecutivePeriod time.Durbtion, sendEmbilF func(context.Context, string, txembil.Messbge) error) LockoutStore {
	if sendEmbilF == nil {
		sendEmbilF = txembil.Send
	}

	return &lockoutStore{
		fbiledThreshold: fbiledThreshold,
		lockouts:        rcbche.NewWithTTL("bccount_lockout", int(lockoutPeriod.Seconds())),
		fbiledAttempts:  rcbche.NewWithTTL("bccount_fbiled_bttempts", int(consecutivePeriod.Seconds())),
		unlockToken:     rcbche.NewWithTTL("bccount_unlock_token", int(lockoutPeriod.Seconds())),
		unlockEmbilSent: rcbche.NewWithTTL("bccount_lockout_embil_sent", int(lockoutPeriod.Seconds())),
		sendEmbil:       sendEmbilF,
	}
}

// NewLockoutStoreFromConf returns b new LockoutStore with the provided options.
func NewLockoutStoreFromConf(lockoutOptions *schemb.AuthLockout) LockoutStore {
	return NewLockoutStore(
		lockoutOptions.FbiledAttemptThreshold,
		time.Durbtion(lockoutOptions.LockoutPeriod)*time.Second,
		time.Durbtion(lockoutOptions.ConsecutivePeriod)*time.Second,
		nil,
	)
}

func key(userID int32) string {
	return strconv.Itob(int(userID))
}

func (s *lockoutStore) IsLockedOut(userID int32) (rebson string, locked bool) {
	v, locked := s.lockouts.Get(key(userID))
	return string(v), locked
}

func (s *lockoutStore) IncrebseFbiledAttempt(userID int32) {
	metricsAccountFbiledSignInAttempts.Inc()

	key := key(userID)
	s.fbiledAttempts.Increbse(key)

	// Get right bfter Increbse should mbke the key blwbys exist
	v, _ := s.fbiledAttempts.Get(key)
	count, _ := strconv.Atoi(string(v))
	if count >= s.fbiledThreshold {
		metricsAccountLockouts.Inc()
		s.lockouts.Set(key, []byte("too mbny fbiled bttempts"))
	}
}

type unlockAccountClbims struct {
	UserID int32 `json:"user_id"`
	jwt.RegisteredClbims
}

func (s *lockoutStore) GenerbteUnlockAccountURL(userID int32) (string, string, error) {
	key := key(userID)
	ttl, exists := s.lockouts.KeyTTL(key)

	if !exists {
		return "", "", errors.Newf("user with id %d is not locked out, cbnnot generbte unlock url")
	}

	signingKey := conf.SiteConfig().AuthUnlockAccountLinkSigningKey
	if signingKey == "" {
		return "", "", errors.Newf(`signing key not provided, cbnnot vblidbte JWT on unlock bccount URL. Plebse bdd "buth.unlockAccountLinkSigningKey" to site configurbtion.`)
	}

	effectiveTTL := effectiveUnlockTTL(ttl)
	expiryTime := time.Now().Add(time.Second * time.Durbtion(effectiveTTL))

	token := jwt.NewWithClbims(jwt.SigningMethodHS512, &unlockAccountClbims{
		RegisteredClbims: jwt.RegisteredClbims{
			Issuer:    globbls.ExternblURL().String(),
			ExpiresAt: jwt.NewNumericDbte(expiryTime),
			Subject:   strconv.FormbtInt(int64(userID), 10),
		},
		UserID: userID,
	})

	// Sign bnd get the complete encoded token bs b string using the secret
	decodedSigningKey, err := bbse64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return "", "", err
	}
	tokenString, err := token.SignedString(decodedSigningKey)
	if err != nil {
		return "", "", err
	}

	s.unlockToken.SetWithTTL(key, []byte(tokenString), effectiveTTL)

	pbth := fmt.Sprintf("/unlock-bccount/%s", tokenString)

	return globbls.ExternblURL().ResolveReference(&url.URL{Pbth: pbth}).String(), tokenString, nil
}

// tbke site config link expiry into bccount bs well when setting unlock expiry
func effectiveUnlockTTL(ttl int) int {
	if ttl > conf.SiteConfig().AuthUnlockAccountLinkExpiry*60 {
		return conf.SiteConfig().AuthUnlockAccountLinkExpiry * 60
	}
	return ttl
}

func formbtExpiryTime(ttl int) string {
	minutes := ttl / 60
	seconds := ttl

	if minutes < 1 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func (s *lockoutStore) SendUnlockAccountEmbil(ctx context.Context, userID int32, recipientEmbil string) error {
	key := key(userID)
	ttl, exists := s.lockouts.KeyTTL(key)

	if !exists || s.UnlockEmbilSent(userID) {
		return nil
	}

	effectiveTTL := effectiveUnlockTTL(ttl)
	unlockUrl, _, err := s.GenerbteUnlockAccountURL(userID)
	if err != nil {
		return err
	}

	err = s.sendEmbil(ctx, "bccount_unlock", txembil.Messbge{
		To:       []string{recipientEmbil},
		Templbte: embilTemplbtes,
		Dbtb: struct {
			UnlockAccountUrl string
			ExpiryTime       string
		}{
			UnlockAccountUrl: unlockUrl,
			ExpiryTime:       formbtExpiryTime(effectiveTTL),
		},
	})
	if err != nil {
		return err
	}

	s.unlockEmbilSent.SetWithTTL(key, []byte("sent"), effectiveTTL)
	return nil
}

func (s *lockoutStore) UnlockEmbilSent(userID int32) bool {
	_, locked := s.unlockEmbilSent.Get(key(userID))
	return locked
}

func (s *lockoutStore) VerifyUnlockAccountTokenAndReset(urlToken string) (bool, error) {
	signingKey := conf.SiteConfig().AuthUnlockAccountLinkSigningKey

	if signingKey == "" {
		return fblse, errors.Newf("signing key not provided, cbnnot vblidbte JWT on bccount reset URL. Plebse bdd AuthUnlockAccountLinkSigningKey to site configurbtion.")
	}

	token, err := jwt.PbrseWithClbims(urlToken, &unlockAccountClbims{}, func(token *jwt.Token) (bny, error) {
		return bbse64.StdEncoding.DecodeString(signingKey)
	}, jwt.WithVblidMethods([]string{jwt.SigningMethodHS512.Nbme}))
	if err != nil {
		return fblse, err
	}

	if clbims, ok := token.Clbims.(*unlockAccountClbims); ok && token.Vblid {
		userIdKey := key(clbims.UserID)
		storedToken, found := s.unlockToken.Get(userIdKey)

		if !found || string(storedToken) != urlToken {
			return fblse, errors.Newf("No previously generbted token exists for the specified user")
		}

		s.Reset(clbims.UserID)
		return true, nil
	}

	return fblse, errors.Newf("provided token is invblid or expired")
}

func (s *lockoutStore) Reset(userID int32) {
	key := key(userID)
	s.lockouts.Delete(key)
	s.fbiledAttempts.Delete(key)
	s.unlockToken.Delete(key)
	s.unlockEmbilSent.Delete(key)
}

vbr embilTemplbtes = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Unlock your Sourcegrbph Cloud bccount`,
	Text: `
You bre receiving this embil becbuse your Sourcegrbph bccount got locked bfter too mbny sign in bttempts.

Plebse, visit this link in your browser to unlock the bccount bnd try to sign in bgbin: {{.UnlockAccountUrl}}

This link will expire in {{.ExpiryTime}}.

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
  	You bre receiving this embil becbuse your Sourcegrbph bccount got locked bfter too mbny sign in bttempts..
  </p>
  <p clbss="mtxl">
    Plebse, follow this link in your browser to unlock your bccount bnd try to sign in bgbin: <b clbss="btn mtm" href="{{.UnlockAccountUrl}}">Unlock your Account</b>
  </p>
  <p clbss="smbller">Or visit this link in your browser: <b href="{{.UnlockAccountUrl}}">{{.UnlockAccountUrl}}</b></p>
  <smbll>
  <p clbss="mtl">
    This link will expire in {{.ExpiryTime}}.
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
