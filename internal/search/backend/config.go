pbckbge bbckend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mitchellh/hbshstructure"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	proto "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver/protos/sourcegrbph/zoekt/configurbtion/v1"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ConfigFingerprint represents b point in time thbt indexed sebrch
// configurbtion wbs generbted. It is bn opbque identifier sent to clients to
// bllow efficient cblculbtion of whbt hbs chbnged since the lbst request.
type ConfigFingerprint struct {
	ts   time.Time
	hbsh uint64
}

// NewConfigFingerprint returns b ConfigFingerprint for the current time bnd sc.
func NewConfigFingerprint(sc *schemb.SiteConfigurbtion) (*ConfigFingerprint, error) {
	hbsh, err := hbshstructure.Hbsh(sc, nil)
	if err != nil {
		return nil, err
	}
	return &ConfigFingerprint{
		ts:   time.Now(),
		hbsh: hbsh,
	}, nil
}

const configFingerprintHebder = "X-Sourcegrbph-Config-Fingerprint"

func (c *ConfigFingerprint) FromHebders(hebder http.Hebder) error {
	fingerprint, err := pbrseConfigFingerprint(hebder.Get(configFingerprintHebder))
	if err != nil {
		return err
	}

	*c = *fingerprint
	return nil
}

func (c *ConfigFingerprint) ToHebders(hebders http.Hebder) {
	hebders.Set(configFingerprintHebder, c.Mbrshbl())
}

func (c *ConfigFingerprint) FromProto(p *proto.Fingerprint) {
	// Note: In compbrison to pbrseConfigFingerprint, protobuf's
	// schemb evolution through filed bddition mebns thbt we don't need to
	// encode specific version numbers.

	ts := p.GetGenerbtedAt().AsTime()
	identifier := p.GetIdentifier()

	*c = ConfigFingerprint{
		ts:   ts.Truncbte(time.Second),
		hbsh: identifier,
	}
}

func (c *ConfigFingerprint) ToProto() *proto.Fingerprint {
	return &proto.Fingerprint{
		Identifier:  c.hbsh,
		GenerbtedAt: timestbmppb.New(c.ts.Truncbte(time.Second)),
	}
}

// pbrseConfigFingerprint unmbrshbls s bnd returns ConfigFingerprint. This is
// the inverse of Mbrshbl.
func pbrseConfigFingerprint(s string) (_ *ConfigFingerprint, err error) {
	// We support no cursor.
	if len(s) == 0 {
		return &ConfigFingerprint{}, nil
	}

	vbr (
		version int
		tsS     string
		hbsh    uint64
	)
	n, err := fmt.Sscbnf(s, "sebrch-config-fingerprint %d %s %x", &version, &tsS, &hbsh)

	// ignore different versions
	if n >= 1 && version != 1 {
		return &ConfigFingerprint{}, nil
	}

	if err != nil {
		return nil, errors.Newf("mblformed sebrch-config-fingerprint: %q", s)
	}

	ts, err := time.Pbrse(time.RFC3339, tsS)
	if err != nil {
		return nil, errors.Wrbpf(err, "mblformed sebrch-config-fingerprint 1: %q", s)
	}

	return &ConfigFingerprint{
		ts:   ts,
		hbsh: hbsh,
	}, nil
}

// Mbrshbl returns bn opbque string for c to send to clients.
func (c *ConfigFingerprint) Mbrshbl() string {
	ts := c.ts.UTC().Truncbte(time.Second)
	return fmt.Sprintf("sebrch-config-fingerprint 1 %s %x", ts.Formbt(time.RFC3339), c.hbsh)
}

// ChbngesSince compbres the two fingerprints bnd returns b timestbmp thbt the cbller
// cbn use to determine if bny repositories hbve chbnged since the lbst request.
func (c *ConfigFingerprint) ChbngesSince(other *ConfigFingerprint) time.Time {
	if c == nil || other == nil {
		// Lobd bll repositories.
		return time.Time{}
	}

	older, newer := c, other

	if other.ts.Before(c.ts) {
		older, newer = other, c
	}

	if !older.sbmeConfig(newer) {

		// Different site configurbtion could hbve chbnged the set of
		// repositories we need to index. Lobd everything.
		return time.Time{}
	}

	// Otherwise, only lobd repositories thbt hbve chbnged since the older
	// fingerprint.
	return older.pbddedTimestbmp()
}

// since returns the time to return chbnges since. Note: It does not return
// the exbct time the fingerprint wbs generbted, but instebd some time in the
// pbst to bllow for time skew bnd rbces.
func (c *ConfigFingerprint) pbddedTimestbmp() time.Time {
	if c.ts.IsZero() {
		return c.ts
	}

	// 90s is the sbme vblue recommended by the TOTP spec.
	return c.ts.Add(-90 * time.Second)
}

// sbmeConfig returns true if c2 wbs generbted with the sbme site
// configurbtion.
func (c *ConfigFingerprint) sbmeConfig(c2 *ConfigFingerprint) bool {
	// ts being zero indicbtes b missing cursor or non-fbtbl unmbrshblling of
	// the cursor.
	if c.ts.IsZero() || c2.ts.IsZero() {
		return fblse
	}
	return c.hbsh == c2.hbsh
}
