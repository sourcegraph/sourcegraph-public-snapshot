package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mitchellh/hashstructure"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const configFingerprintHeader = "X-Sourcegraph-Config-Fingerprint"

// ParseAndSetConfigFingerprint will set the current config fingerprint in
// w. If r specifies a config fingerprint, we return the minimum time for a
// repository to have changed.
//
// A config fingerprint represents a point in time that indexed search
// configuration was generated. It is an opaque identifier sent to clients to
// allow efficient calculation of what has changed since the last
// request. zoekt-sourcegraph-indexserver reads and sets these headers to
// reduce the amount of work required when it polls.
func ParseAndSetConfigFingerprint(w http.ResponseWriter, r *http.Request, siteConfig *schema.SiteConfiguration) (minLastChanged time.Time, err error) {
	// Before we load anything generate a config fingerprint representing the
	// point in time just before loading. This is sent to the client via a
	// header for use in the next call.
	fingerprint, err := newConfigFingerprint(siteConfig)
	if err != nil {
		return time.Time{}, err
	}
	w.Header().Set(configFingerprintHeader, fingerprint.Marshal())

	// If the user specified a fingerprint to diff against, we can use it to
	// reduce the amount of work we do. minLastChanged being zero means we
	// check every repository.
	old, err := parseConfigFingerprint(r.Header.Get(configFingerprintHeader))
	if err != nil {
		return time.Time{}, err
	}

	// Different site config may affect any repository, so we need to load
	// them all in.
	if !old.SameConfig(fingerprint) {
		return time.Time{}, nil
	}

	// We can just load what has changed since the last config fingerprint.
	return old.Since(), nil
}

// configFingerprint represents a point in time that indexed search
// configuration was generated. It is an opaque identifier sent to clients to
// allow efficient calculation of what has changed since the last request.
type configFingerprint struct {
	ts   time.Time
	hash uint64
}

// newConfigFingerprint returns a ConfigFingerprint for the current time and sc.
func newConfigFingerprint(sc *schema.SiteConfiguration) (*configFingerprint, error) {
	hash, err := hashstructure.Hash(sc, nil)
	if err != nil {
		return nil, err
	}
	return &configFingerprint{
		ts:   time.Now(),
		hash: hash,
	}, nil
}

// parseConfigFingerprint unmarshals s and returns ConfigFingerprint. This is
// the inverse of Marshal.
func parseConfigFingerprint(s string) (_ *configFingerprint, err error) {
	// We support no cursor.
	if len(s) == 0 {
		return &configFingerprint{}, nil
	}

	var (
		version int
		tsS     string
		hash    uint64
	)
	n, err := fmt.Sscanf(s, "search-config-fingerprint %d %s %x", &version, &tsS, &hash)

	// ignore different versions
	if n >= 1 && version != 1 {
		return &configFingerprint{}, nil
	}

	if err != nil {
		return nil, errors.Newf("malformed search-config-fingerprint: %q", s)
	}

	ts, err := time.Parse(time.RFC3339, tsS)
	if err != nil {
		return nil, errors.Wrapf(err, "malformed search-config-fingerprint 1: %q", s)
	}

	return &configFingerprint{
		ts:   ts,
		hash: hash,
	}, nil
}

// Marshal returns an opaque string for c to send to clients.
func (c *configFingerprint) Marshal() string {
	ts := c.ts.UTC().Truncate(time.Second)
	return fmt.Sprintf("search-config-fingerprint 1 %s %x", ts.Format(time.RFC3339), c.hash)
}

// Since returns the time to return changes since. Note: It does not return
// the exact time the fingerprint was generated, but instead some time in the
// past to allow for time skew and races.
func (c *configFingerprint) Since() time.Time {
	if c.ts.IsZero() {
		return c.ts
	}
	// 90s is the same value recommended by the TOTP spec.
	return c.ts.Add(-90 * time.Second)
}

// SameConfig returns true if c2 was generated with the same site
// configuration.
func (c *configFingerprint) SameConfig(c2 *configFingerprint) bool {
	// ts being zero indicates a missing cursor or non-fatal unmarshalling of
	// the cursor.
	if c.ts.IsZero() || c2.ts.IsZero() {
		return false
	}
	return c.hash == c2.hash
}
