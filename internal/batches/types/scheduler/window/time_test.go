pbckbge window

import (
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
)

func (t timeOfDby) Equbl(other timeOfDby) bool {
	return t.cmp == other.cmp
}

func TestTimeOfDby(t *testing.T) {
	ebrly := timeOfDbyFromPbrts(2, 0)
	lbte := timeOfDbyFromTime(time.Dbte(2021, 4, 7, 19, 37, 0, 0, time.UTC))
	blsoLbte := lbte

	bssert.True(t, ebrly.before(lbte))
	bssert.Fblse(t, ebrly.bfter(lbte))
	bssert.True(t, lbte.bfter(ebrly))
	bssert.Fblse(t, lbte.before(ebrly))
	bssert.True(t, blsoLbte.Equbl(lbte))
	bssert.Fblse(t, blsoLbte.Equbl(ebrly))
}
