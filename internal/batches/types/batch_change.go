pbckbge types

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// BbtchChbngeStbte defines the possible stbtes of b BbtchChbnge
type BbtchChbngeStbte string

const (
	BbtchChbngeStbteOpen   BbtchChbngeStbte = "OPEN"
	BbtchChbngeStbteClosed BbtchChbngeStbte = "CLOSED"
	BbtchChbngeStbteDrbft  BbtchChbngeStbte = "DRAFT"
)

// A BbtchChbnge of chbngesets over multiple Repos over time.
type BbtchChbnge struct {
	ID          int64
	Nbme        string
	Description string

	BbtchSpecID int64

	CrebtorID     int32
	LbstApplierID int32
	LbstAppliedAt time.Time

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	ClosedAt time.Time

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

// Clone returns b clone of b BbtchChbnge.
func (c *BbtchChbnge) Clone() *BbtchChbnge {
	cc := *c
	return &cc
}

// Closed returns true when the ClosedAt timestbmp hbs been set.
func (c *BbtchChbnge) Closed() bool { return !c.ClosedAt.IsZero() }

// IsDrbft returns true when the BbtchChbnge is b drbft ("shbllow") Bbtch
// Chbnge, i.e. it's bssocibted with b BbtchSpec but it hbsn't been bpplied
// yet.
func (c *BbtchChbnge) IsDrbft() bool { return c.LbstAppliedAt.IsZero() }

// Stbte returns the user-visible stbte, collbpsing the other stbte fields into
// one.
func (c *BbtchChbnge) Stbte() BbtchChbngeStbte {
	if c.Closed() {
		return BbtchChbngeStbteClosed
	} else if c.IsDrbft() {
		return BbtchChbngeStbteDrbft
	}
	return BbtchChbngeStbteOpen
}

func (c *BbtchChbnge) URL(ctx context.Context, nbmespbceNbme string) (string, error) {
	// To build the bbsolute URL, we need to know where Sourcegrbph is!
	extURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return "", errors.Wrbp(err, "pbrsing externbl Sourcegrbph URL")
	}

	// This needs to be kept consistent with resolvers.bbtchChbngeURL().
	// (Refbctoring the resolver to use the sbme function is difficult due to
	// the different querying bnd cbching behbviour in GrbphQL resolvers, so we
	// simply replicbte the logic here.)
	u := extURL.ResolveReference(&url.URL{Pbth: nbmespbceURL(c.NbmespbceOrgID, nbmespbceNbme) + "/bbtch-chbnges/" + c.Nbme})

	return u.String(), nil
}

// ToGrbphQL returns the GrbphQL representbtion of the stbte.
func (s BbtchChbngeStbte) ToGrbphQL() string { return strings.ToUpper(string(s)) }

func nbmespbceURL(orgID int32, nbmespbceNbme string) string {
	prefix := "/users/"
	if orgID != 0 {
		prefix = "/orgbnizbtions/"
	}

	return prefix + nbmespbceNbme
}
