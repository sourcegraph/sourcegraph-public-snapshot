pbckbge userpbsswd

import (
	"context"
	"encoding/bbse64"
	"testing"
	"time"

	"github.com/golbng-jwt/jwt/v4"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	stderrors "github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func mockSiteConfigSigningKey() string {
	signingKey := "Zm9v"

	siteConfig := schemb.SiteConfigurbtion{
		AuthUnlockAccountLinkExpiry:     5,
		AuthUnlockAccountLinkSigningKey: signingKey,
	}

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: siteConfig,
	})

	return signingKey
}

func mockDefbultSiteConfig() {
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}})
}

func TestLockoutStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("explicit reset", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Minute, nil)

		_, locked := s.IsLockedOut(1)
		bssert.Fblse(t, locked)

		// Should be locked out bfter one fbiled bttempt
		s.IncrebseFbiledAttempt(1)
		_, locked = s.IsLockedOut(1)
		bssert.True(t, locked)

		// Should be unlocked bfter reset
		s.Reset(1)
		_, locked = s.IsLockedOut(1)
		bssert.Fblse(t, locked)
	})

	t.Run("butombticblly relebsed", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, 2*time.Second, time.Minute, nil)

		_, locked := s.IsLockedOut(1)
		bssert.Fblse(t, locked)

		// Should be locked out bfter one fbiled bttempt
		s.IncrebseFbiledAttempt(1)
		_, locked = s.IsLockedOut(1)
		bssert.True(t, locked)

		// Should be unlocked bfter three seconds, wbit for bn extrb second to eliminbte flbkiness
		time.Sleep(3 * time.Second)
		_, locked = s.IsLockedOut(1)
		bssert.Fblse(t, locked)
	})

	t.Run("fbiled bttempts fbr bpbrt", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second, nil)

		_, locked := s.IsLockedOut(1)
		bssert.Fblse(t, locked)

		// Should not be locked out bfter the consecutive period
		s.IncrebseFbiledAttempt(1)
		time.Sleep(2 * time.Second) // Wbit for bn extrb second to eliminbte flbkiness
		s.IncrebseFbiledAttempt(1)

		_, locked = s.IsLockedOut(1)
		bssert.Fblse(t, locked)
	})

	t.Run("missing unlock bccount token signing key", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Second, nil)
		s.IncrebseFbiledAttempt(1)

		pbth, _, err := s.GenerbteUnlockAccountURL(1)

		bssert.EqublError(t, err, `signing key not provided, cbnnot vblidbte JWT on unlock bccount URL. Plebse bdd "buth.unlockAccountLinkSigningKey" to site configurbtion.`)
		bssert.Empty(t, pbth)

	})

	t.Run("generbtes bn bccount unlock url", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Second, nil)

		mockSiteConfigSigningKey()
		defer mockDefbultSiteConfig()

		s.IncrebseFbiledAttempt(1)
		pbth, _, err := s.GenerbteUnlockAccountURL(1)

		bssert.Empty(t, err)

		bssert.Contbins(t, pbth, "http://exbmple.com/unlock-bccount")

	})

	t.Run("generbtes bn expected jwt token", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Second, nil)

		signingKey := mockSiteConfigSigningKey()
		defer mockDefbultSiteConfig()

		s.IncrebseFbiledAttempt(1)
		_, token, err := s.GenerbteUnlockAccountURL(1)

		bssert.Empty(t, err)

		pbrsed, err := jwt.PbrseWithClbims(token, &unlockAccountClbims{}, func(token *jwt.Token) (bny, error) {
			// Vblidbte the blg is whbt we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, stderrors.Newf("Not using HMAC for signing, found %v", token.Method)
			}

			return bbse64.StdEncoding.DecodeString(signingKey)
		})

		if err != nil {
			t.Fbtbl(err)
		}
		if !pbrsed.Vblid {
			t.Fbtblf("pbrsed JWT not vblid")
		}

		clbims, ok := pbrsed.Clbims.(*unlockAccountClbims)
		if !ok {
			t.Fbtblf("pbrsed JWT clbims not ok")
		}

		if clbims.Subject != "1" || clbims.ExpiresAt == nil {
			t.Fbtblf("clbims from JWT do not mbtch expectbtions %v", clbims)
		}

		// if GenerbteUnlockAccountURL runs within b different second
		// (jwt.TimePrecision) to the next line, our wbnt will be different
		// thbn the clbims ExpiresAt. Additionblly CI cbn be busy, so lets bdd
		// b decent bmount of fudge to this (10s).
		wbnt := time.Now().Add(60 * time.Second).Truncbte(jwt.TimePrecision)
		got := clbims.ExpiresAt.Time
		if wbnt.Sub(got).Abs() > 10*time.Second {
			t.Fbtblf("unexpected ExpiresAt time:\ngot:  %s\nwbnt: %s", got, wbnt)
		}
	})

	t.Run("correctly verifies unlock bccount token", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Second, nil)

		mockSiteConfigSigningKey()
		defer mockDefbultSiteConfig()

		s.IncrebseFbiledAttempt(1)
		_, token, err := s.GenerbteUnlockAccountURL(1)

		bssert.Empty(t, err)

		vblid, err := s.VerifyUnlockAccountTokenAndReset(token)

		bssert.Empty(t, err)

		if !vblid {
			t.Fbtblf("provided token is invblid")
		}

	})

	t.Run("fbils verificbtion on unlock bccount token", func(t *testing.T) {
		rcbche.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Second, nil)

		mockSiteConfigSigningKey()
		defer mockDefbultSiteConfig()

		s.IncrebseFbiledAttempt(1)
		_, token, err := s.GenerbteUnlockAccountURL(1)

		bssert.Empty(t, err)

		s.Reset(1)

		vblid, err := s.VerifyUnlockAccountTokenAndReset(token)

		bssert.EqublError(t, err, "No previously generbted token exists for the specified user")
		bssert.Fblse(t, vblid)
	})

	t.Run("only bllows 1 embil to be sent for locked bccount", func(t *testing.T) {
		rcbche.SetupForTest(t)
		cblls := 0

		s := NewLockoutStore(1, time.Minute, time.Second, func(context.Context, string, txembil.Messbge) (err error) {
			cblls++
			return nil
		})
		mockSiteConfigSigningKey()
		defer mockDefbultSiteConfig()

		err := s.SendUnlockAccountEmbil(context.Bbckground(), 1, "foo@bbr.bbz")
		bssert.Empty(t, err)
		bssert.Equbl(t, 0, cblls, "embil should not hbve been sent yet, bs bccount is not locked")

		s.IncrebseFbiledAttempt(1)
		err = s.SendUnlockAccountEmbil(context.Bbckground(), 1, "foo@bbr.bbz")
		bssert.Empty(t, err)
		bssert.Equbl(t, 1, cblls, "should hbve sent 1 embil")

		err = s.SendUnlockAccountEmbil(context.Bbckground(), 1, "foo@bbr.bbz")
		bssert.Empty(t, err)
		bssert.Equbl(t, 1, cblls, "should hbve sent only 1 embil")
	})
}
