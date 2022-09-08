package userpasswd

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	stderrors "github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func mockSiteConfigSigningKey() string {
	signingKey := "Zm9v"

	siteConfig := schema.SiteConfiguration{
		AuthUnlockAccountLinkExpiry:     5,
		AuthUnlockAccountLinkSigningKey: signingKey,
	}

	conf.Mock(&conf.Unified{
		SiteConfiguration: siteConfig,
	})

	return signingKey
}

func mockDefaultSiteConfig() {
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{}})
}

func TestLockoutStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("explicit reset", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Minute)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should be locked out after one failed attempt
		s.IncreaseFailedAttempt(1)
		_, locked = s.IsLockedOut(1)
		assert.True(t, locked)

		// Should be unlocked after reset
		s.Reset(1)
		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})

	t.Run("automatically released", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(1, 2*time.Second, time.Minute)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should be locked out after one failed attempt
		s.IncreaseFailedAttempt(1)
		_, locked = s.IsLockedOut(1)
		assert.True(t, locked)

		// Should be unlocked after three seconds, wait for an extra second to eliminate flakiness
		time.Sleep(3 * time.Second)
		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})

	t.Run("failed attempts far apart", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should not be locked out after the consecutive period
		s.IncreaseFailedAttempt(1)
		time.Sleep(2 * time.Second) // Wait for an extra second to eliminate flakiness
		s.IncreaseFailedAttempt(1)

		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})

	t.Run("missing unlock account token signing key", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		path, _, err := s.GenerateUnlockAccountURL(1)

		assert.EqualError(t, err, `signing key not provided, cannot validate JWT on unlock account URL. Please add "auth.unlockAccountLinkSigningKey" to site configuration.`)
		assert.Empty(t, path)

	})

	t.Run("generates an account unlock url", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		mockSiteConfigSigningKey()
		defer mockDefaultSiteConfig()

		path, _, err := s.GenerateUnlockAccountURL(1)

		assert.Empty(t, err)

		assert.Contains(t, path, "http://example.com/unlock-account")

	})

	t.Run("generates an expected jwt token", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		signingKey := mockSiteConfigSigningKey()
		defer mockDefaultSiteConfig()

		_, token, err := s.GenerateUnlockAccountURL(1)

		assert.Empty(t, err)

		parsed, err := jwt.ParseWithClaims(token, &unlockAccountClaims{}, func(token *jwt.Token) (any, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, stderrors.Newf("Not using HMAC for signing, found %v", token.Method)
			}

			return base64.StdEncoding.DecodeString(signingKey)
		})

		if err != nil {
			t.Fatal(err)
		}
		if !parsed.Valid {
			t.Fatalf("parsed JWT not valid")
		}

		claims, ok := parsed.Claims.(*unlockAccountClaims)
		if !ok {
			t.Fatalf("parsed JWT claims not ok")
		}

		if claims.Subject != "1" || claims.ExpiresAt == nil {
			t.Fatalf("claims from JWT do not match expectations %v", claims)
		}

		// if GenerateUnlockAccountURL runs within a different second
		// (jwt.TimePrecision) to the next line, our want will be different
		// than the claims ExpiresAt. Additionally CI can be busy, so lets add
		// a decent amount of fudge to this (10s).
		want := time.Now().Add(5 * time.Minute).Truncate(jwt.TimePrecision)
		got := *&claims.ExpiresAt.Time
		if durationAbs(want.Sub(got)) > 10*time.Second {
			t.Fatalf("unexpected ExpiresAt time:\ngot:  %s\nwant: %s", got, want)
		}
	})

	t.Run("correctly verifies unlock account token", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		mockSiteConfigSigningKey()
		defer mockDefaultSiteConfig()

		_, token, err := s.GenerateUnlockAccountURL(1)

		assert.Empty(t, err)

		valid, error := s.VerifyUnlockAccountTokenAndReset(token)

		assert.Empty(t, error)

		if !valid {
			t.Fatalf("provided token is invalid")
		}

	})

	t.Run("fails verification on unlock account token", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		mockSiteConfigSigningKey()
		defer mockDefaultSiteConfig()

		_, token, err := s.GenerateUnlockAccountURL(1)

		assert.Empty(t, err)

		s.Reset(1)

		valid, error := s.VerifyUnlockAccountTokenAndReset(token)

		assert.EqualError(t, error, "No previously generated token exists for the specified user")
		assert.False(t, valid)
	})
}

// durationAbs returns the absolute value of d.
//
// Copy-pasta from Duration.Abs in the stdlib, but only available in go1.19.
// Can remove this helper once we require go1.19 as a minimum.
func durationAbs(d time.Duration) time.Duration {
	switch {
	case d >= 0:
		return d
	case d == -1<<63:
		return 1<<63 - 1
	default:
		return -d
	}
}
