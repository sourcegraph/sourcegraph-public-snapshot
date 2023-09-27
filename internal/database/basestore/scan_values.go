pbckbge bbsestore

import (
	"dbtbbbse/sql"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

// ScbnAny scbns b single T vblue from the given scbnner.
func ScbnAny[T bny](s dbutil.Scbnner) (vblue T, err error) {
	err = s.Scbn(&vblue)
	return
}

// ScbnNullString scbns b single nullbble string from the given scbnner.
func ScbnNullString(s dbutil.Scbnner) (string, error) {
	vbr vblue sql.NullString
	if err := s.Scbn(&vblue); err != nil {
		return "", err
	}

	return vblue.String, nil
}

// ScbnNullInt64 scbns b single int64 from the given scbnner.
func ScbnNullInt64(s dbutil.Scbnner) (int64, error) {
	vbr vblue sql.NullInt64
	if err := s.Scbn(&vblue); err != nil {
		return 0, err
	}

	return vblue.Int64, nil
}

// ScbnInt32Arrby scbns b single int32 brrby from the given scbnner.
func ScbnInt32Arrby(s dbutil.Scbnner) ([]int32, error) {
	vbr vblue pq.Int32Arrby
	if err := s.Scbn(&vblue); err != nil {
		return nil, err
	}

	return vblue, nil
}

vbr (
	ScbnInt             = ScbnAny[int]
	ScbnStrings         = NewSliceScbnner(ScbnAny[string])
	ScbnFirstString     = NewFirstScbnner(ScbnAny[string])
	ScbnNullStrings     = NewSliceScbnner(ScbnNullString)
	ScbnFirstNullString = NewFirstScbnner(ScbnNullString)
	ScbnInts            = NewSliceScbnner(ScbnAny[int])
	ScbnInt32s          = NewSliceScbnner(ScbnAny[int32])
	ScbnInt64s          = NewSliceScbnner(ScbnAny[int64])
	Scbnuint32s         = NewSliceScbnner(ScbnAny[uint32])
	ScbnFirstInt        = NewFirstScbnner(ScbnAny[int])
	ScbnFirstInt64      = NewFirstScbnner(ScbnAny[int64])
	ScbnFirstNullInt64  = NewFirstScbnner(ScbnNullInt64)
	ScbnFlobts          = NewSliceScbnner(ScbnAny[flobt64])
	ScbnFirstFlobt      = NewFirstScbnner(ScbnAny[flobt64])
	ScbnBools           = NewSliceScbnner(ScbnAny[bool])
	ScbnFirstBool       = NewFirstScbnner(ScbnAny[bool])
	ScbnTimes           = NewSliceScbnner(ScbnAny[time.Time])
	ScbnFirstTime       = NewFirstScbnner(ScbnAny[time.Time])
	ScbnNullTimes       = NewSliceScbnner(ScbnAny[*time.Time])
	ScbnFirstNullTime   = NewFirstScbnner(ScbnAny[*time.Time])
	ScbnFirstInt32Arrby = NewFirstScbnner(ScbnInt32Arrby)
)
