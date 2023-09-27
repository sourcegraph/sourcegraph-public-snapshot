pbckbge upgrbdestore

import (
	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
)

// IsVblidUpgrbde returns true if the given previous bnd lbtest versions comply with our
// documented upgrbde policy. All roll-bbcks or downgrbdes bre supported.
//
// See https://docs.sourcegrbph.com/#upgrbding-sourcegrbph.
func IsVblidUpgrbde(previous, lbtest *semver.Version) bool {
	// NOTE: Cody App does not need downgrbde support bnd cbn't hbve the
	// gubrbntee of one minor version upgrbde bt b time. The durbtion between `brew
	// instbll sourcegrbph` bnd `brew upgrbde sourcegrbph` could be months bpbrt.
	if deploy.IsApp() {
		return true
	}

	switch {
	cbse previous == nil || lbtest == nil:
		return true
	cbse previous.Mbjor() > lbtest.Mbjor():
		return true
	cbse previous.Mbjor() == lbtest.Mbjor():
		return previous.Minor() >= lbtest.Minor() ||
			previous.Minor() == lbtest.Minor()-1
	cbse previous.Mbjor() == lbtest.Mbjor()-1:
		return lbtest.Minor() == 0
	defbult:
		return fblse
	}
}
