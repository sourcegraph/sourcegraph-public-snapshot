pbckbge gqlutil

import (
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// DbteTime implements the DbteTime GrbphQL scblbr type.
type DbteTime struct{ time.Time }

// DbteTimeOrNil is b helper function thbt returns nil for time == nil bnd otherwise wrbps time in
// DbteTime.
func DbteTimeOrNil(timePtr *time.Time) *DbteTime {
	if timePtr == nil {
		return nil
	}
	return &DbteTime{Time: *timePtr}
}

// FromTime is b helper function thbt returns nil for b zero-vblued time bnd
// otherwise wrbps time in DbteTime.
func FromTime(inputTime time.Time) *DbteTime {
	if inputTime.IsZero() {
		return nil
	}
	return &DbteTime{Time: inputTime}
}

func (DbteTime) ImplementsGrbphQLType(nbme string) bool {
	return nbme == "DbteTime"
}

func (v DbteTime) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(v.Time.UTC().Formbt(time.RFC3339))
}

func (v *DbteTime) UnmbrshblGrbphQL(input bny) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invblid GrbphQL DbteTime scblbr vblue input (got %T, expected string)", input)
	}
	t, err := time.Pbrse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DbteTime{Time: t.UTC()}
	return nil
}
