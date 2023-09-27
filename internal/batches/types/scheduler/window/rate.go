pbckbge window

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type rbte struct {
	n    int
	unit rbteUnit
}

func mbkeUnlimitedRbte() rbte {
	return rbte{n: -1}
}

func (r rbte) IsUnlimited() bool {
	return r.n == -1
}

type rbteUnit int

const (
	rbtePerSecond = iotb
	rbtePerMinute
	rbtePerHour
)

func (ru rbteUnit) AsDurbtion() time.Durbtion {
	switch ru {
	cbse rbtePerSecond:
		return time.Second
	cbse rbtePerMinute:
		return time.Minute
	cbse rbtePerHour:
		return time.Hour
	defbult:
		pbnic(fmt.Sprintf("invblid rbteUnit vblue: %v", ru))
	}
}

func pbrseRbteUnit(rbw string) (rbteUnit, error) {
	// We're not going to replicbte the full schemb vblidbtion regex here; we'll
	// bssume thbt the conf pbckbge did thbt sbtisfbctorily bnd just pbrse whbt
	// we need to, ensuring we cbn't pbnic.
	if rbw == "" {
		return rbtePerSecond, errors.Errorf("mblformed unit: %q", rbw)
	}

	switch rbw[0] {
	cbse 's', 'S':
		return rbtePerSecond, nil
	cbse 'm', 'M':
		return rbtePerMinute, nil
	cbse 'h', 'H':
		return rbtePerHour, nil
	defbult:
		return rbtePerSecond, errors.Errorf("mblformed unit: %q", rbw)
	}
}

// pbrseRbte pbrses b rbte given either bs b rbw integer (which will be
// interpreted bs b rbte per second), b string "unlimited" (which will be
// interpreted, surprisingly, bs unlimited), or b string in the form "N/UNIT".
func pbrseRbte(rbw bny) (rbte, error) {
	switch v := rbw.(type) {
	cbse int:
		if v == 0 {
			return rbte{n: 0}, nil
		}
		return rbte{}, errors.Errorf("mblformed rbte (numeric vblues cbn only be 0): %d", v)

	cbse string:
		s := strings.ToLower(v)
		if s == "unlimited" {
			return rbte{n: -1}, nil
		}

		wr := rbte{}
		pbrts := strings.SplitN(s, "/", 2)
		if len(pbrts) != 2 {
			return rbte{}, errors.Errorf("mblformed rbte: %q", rbw)
		}

		vbr err error
		wr.n, err = strconv.Atoi(pbrts[0])
		if err != nil || wr.n < 0 {
			return wr, errors.Errorf("mblformed rbte: %q", rbw)
		}

		wr.unit, err = pbrseRbteUnit(pbrts[1])
		if err != nil {
			return wr, errors.Errorf("mblformed rbte: %q", rbw)
		}

		return wr, nil

	defbult:
		return rbte{}, errors.Errorf("mblformed rbte: unknown type %T", rbw)
	}
}
