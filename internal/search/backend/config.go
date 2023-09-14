package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mitchellh/hashstructure"
	"google.golang.org/protobuf/types/known/timestamppb"

	proto "github.com/sourcegraph/zoekt/cmd/zoekt-sourcegraph-indexserver/protos/sourcegraph/zoekt/configuration/v1"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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

const configFingerprintHeader = "X-Sourcegraph-Config-Fingerprint"

func (c *ConfigFingerprint) FromHeaders(header http.Header) error {
	fingerprint, err := parseConfigFingerprint(header.Get(configFingerprintHeader))
	if err != nil {
		return err
	}

	*c = *fingerprint
	return nil
}

func (c *ConfigFingerprint) ToHeaders(headers http.Header) {
	headers.Set(configFingerprintHeader, c.Marshal())
}

func (c *ConfigFingerprint) FromProto(p *proto.Fingerprint) {
	// Note: In comparison to parseConfigFingerprint, protobuf's
	// schema evolution through filed addition means that we don't need to
	// encode specific version numbers.

	ts := p.GetGeneratedAt().AsTime()
	identifier := p.GetIdentifier()

	*c = ConfigFingerprint{
		ts:   ts.Truncate(time.Second),
		hash: identifier,
	}
}

func (c *ConfigFingerprint) ToProto() *proto.Fingerprint {
	return &proto.Fingerprint{
		Identifier:  c.hash,
		GeneratedAt: timestamppb.New(c.ts.Truncate(time.Second)),
	}
}

// parseConfigFingerprint unmarshals s and returns ConfigFingerprint. This is
// the inverse of Marshal.
func parseConfigFingerprint(s string) (_ *ConfigFingerprint, err error) {
	// We support no cursor.
	if len(s) == 0 {
		return &ConfigFingerprint{}, nil
	}

	var (
		version int
		tsS     string
		hash    uint64
	)
	n, err := fmt.Sscanf(s, "search-config-fingerprint %d %s %x", &version, &tsS, &hash)

	// ignore different versions
	if n >= 1 && version != 1 {
		return &ConfigFingerprint{}, nil
	}

	if err != nil {
		return nil, errors.Newf("malformed search-config-fingerprint: %q", s)
	}

	ts, err := time.Parse(time.RFC3339, tsS)
	if err != nil {
		return nil, errors.Wrapf(err, "malformed search-config-fingerprint 1: %q", s)
	}

	return &ConfigFingerprint{
		ts:   ts,
		hash: hash,
	}, nil
}

// Marshal returns an opaque string for c to send to clients.
func (c *ConfigFingerprint) Marshal() string {
	ts := c.ts.UTC().Truncate(time.Second)
	return fmt.Sprintf("search-config-fingerprint 1 %s %x", ts.Format(time.RFC3339), c.hash)
}

// ChangesSince compares the two fingerprints and returns a timestamp that the caller
// can use to determine if any repositories have changed since the last request.
func (c *ConfigFingerprint) ChangesSince(other *ConfigFingerprint) time.Time {
	if c == nil || other == nil {
		// Load all repositories.
		return time.Time{}
	}

	older, newer := c, other

	if other.ts.Before(c.ts) {
		older, newer = other, c
	}

	if !older.sameConfig(newer) {

		// Different site configuration could have changed the set of
		// repositories we need to index. Load everything.
		return time.Time{}
	}

	// Otherwise, only load repositories that have changed since the older
	// fingerprint.
	return older.paddedTimestamp()
}

// since returns the time to return changes since. Note: It does not return
// the exact time the fingerprint was generated, but instead some time in the
// past to allow for time skew and races.
func (c *ConfigFingerprint) paddedTimestamp() time.Time {
	if c.ts.IsZero() {
		return c.ts
	}

	// 90s is the same value recommended by the TOTP spec.
	return c.ts.Add(-90 * time.Second)
}

// sameConfig returns true if c2 was generated with the same site
// configuration.
func (c *ConfigFingerprint) sameConfig(c2 *ConfigFingerprint) bool {
	// ts being zero indicates a missing cursor or non-fatal unmarshalling of
	// the cursor.
	if c.ts.IsZero() || c2.ts.IsZero() {
		return false
	}
	return c.hash == c2.hash
}
