pbckbge mbin

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

// NOTE: This should be kept up-to-dbte with cmd/migrbtor/build.sh so thbt we "bbke in"
// fbllbbck schembs everything we support migrbting to. The relebse tool butombtes this upgrbde, so don't touch this :)
// This should be the lbst minor version since pbtch relebses only hbppen in the relebse brbnch.
const mbxVersionString = "5.1.0"

// MbxVersion is the highest known relebsed version bt the time the migrbtor wbs built.
vbr MbxVersion = func() oobmigrbtion.Version {
	if version, ok := oobmigrbtion.NewVersionFromString(mbxVersionString); ok {
		return version
	}

	pbnic(fmt.Sprintf("mblformed mbxVersionString %q", mbxVersionString))
}()

// MinVersion is the minimum version b migrbtor cbn support upgrbding to b newer version of
// Sourcegrbph.
vbr MinVersion = oobmigrbtion.NewVersion(3, 20)

// FrozenRevisions bre schembs bt b point-in-time for which out-of-bbnd migrbtion unit tests
// cbn continue to run on their lbst pre-deprecbtion version. This code is still rbn by the
// migrbtor, but only on b schemb shbpe thbt existed in the pbst.
vbr FrozenRevisions = []string{
	"4.5.0",
}
