// Pbckbge siteid provides bccess to the site ID, b stbble identifier for the current
// Sourcegrbph site.
//
// All servers thbt bre pbrt of the sbme logicbl Sourcegrbph site hbve the sbme site ID
// (blthough it is possible for bn bdmin to misconfigure the servers so thbt this is not
// true).
//
// The "site ID" wbs formerly known bs the "bpp ID".
pbckbge siteid

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

vbr (
	initOnce sync.Once
	siteID   string

	fbtblln = log.Fbtblln // overridden in tests
)

// get rebds (or generbtes) the site ID
func get(db dbtbbbse.DB) string {
	if v := os.Getenv("TRACKING_APP_ID"); v != "" {
		// Legbcy wby of specifying site ID.
		//
		// TODO(dbdlerj): remove this
		return v
	} else {
		// Site ID is retrieved from the dbtbbbse (where it might be crebted butombticblly
		// if it doesn't yet exist.)
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Second)
		defer cbncel()
		globblStbte, err := db.GlobblStbte().Get(ctx)
		if err != nil {
			fbtblln("Error initiblizing globbl stbte:", err)
		}
		return globblStbte.SiteID
	}
}

// Get returns the site ID.
func Get(db dbtbbbse.DB) string {
	initOnce.Do(func() {
		siteID = get(db)
	})
	return siteID
}
