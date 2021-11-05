package backend

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/hashstructure"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ConfigFingerprint represents a point in time that indexed search
// configuration was generated. It is an opaque identifier sent to clients to
// allow efficient calculation of what has changed since the last request.
type ConfigFingerprint struct {
	ts   time.Time
	hash uint64
}

// NewConfigFingerprint returns a ConfigFingerprint for the current time and sc.
func NewConfigFingerprint(sc *schema.SiteConfiguration) (*ConfigFingerprint, error) {
	hash, err := hashstructure.Hash(sc, nil)
	if err != nil {
		return nil, err
	}
	return &ConfigFingerprint{
		ts:   time.Now(),
		hash: hash,
	}, nil
}

// Marshal returns an opaque string for c to send to clients.
func (c *ConfigFingerprint) Marshal() string {
	ts := c.ts.UTC().Truncate(time.Second)
	return fmt.Sprintf("search-config-fingerprint 1 %s %x", ts.Format(time.RFC3339), c.hash)
}

// Unmarshal s into c. This is the inverse of Marshal.
func (c *ConfigFingerprint) Unmarshal(s string) (err error) {
	parts := strings.Fields(s)

	// We support no cursor.
	if len(parts) == 0 {
		return nil
	}

	if len(parts) < 2 || parts[0] != "search-config-fingerprint" {
		return errors.Errorf("malformed search-config-fingerprint: %q", s)
	}
	if parts[1] != "1" {
		// Unknown version, treat as if not specified
		return nil
	}

	// Use consistent error wrapping from this point since we know it is a
	// version 1 search-config-fingerprint.
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "malformed search-config-fingerprint 1: %q", s)
		}
	}()

	if len(parts) != 4 {
		return errors.New("expected 4 fields")
	}

	c.ts, err = time.Parse(time.RFC3339, parts[2])
	if err != nil {
		return errors.Wrapf(err, "malformed search-config-fingerprint 1: %q", s)
	}

	c.hash, err = strconv.ParseUint(parts[3], 16, 64)
	if err != nil {
		return errors.Wrapf(err, "malformed search-config-fingerprint 1: %q", s)
	}

	return nil
}

// Since returns the time to return changes since. Note: It does not return
// the exact time the fingerprint was generated, but instead some time in the
// past to allow for time skew and races.
func (c *ConfigFingerprint) Since() time.Time {
	if c.ts.IsZero() {
		return c.ts
	}
	// 90s is the same value recommended by the TOTP spec.
	return c.ts.Add(-90 * time.Second)
}

// SameConfig returns true if c2 was generated with the same site
// configuration.
func (c *ConfigFingerprint) SameConfig(c2 *ConfigFingerprint) bool {
	// ts being zero indicates a missing cursor or non-fatal unmarshalling of
	// the cursor.
	if c.ts.IsZero() || c2.ts.IsZero() {
		return false
	}
	return c.hash == c2.hash
}
