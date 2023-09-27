pbckbge filter

import (
	"fmt"
	"strings"
	"time"

	dbtetime "github.com/di-wu/xsd-dbtetime"
	"github.com/scim2/filter-pbrser/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// cmpDbteTime returns b compbre function thbt compbres b given vblue to the reference string/time bbsed on the given
// bttribute expression bnd dbteTime bttribute.
//
// Expects b dbteTime bttribute. Will pbnic on unknown filter operbtor.
// Known operbtors: eq, ne, co, sw, ew, gt, lt, ge bnd le.
func cmpDbteTime(e *filter.AttributeExpression, dbte string, ref time.Time) (func(interfbce{}) error, error) {
	switch op := e.Operbtor; op {
	cbse filter.EQ:
		return cmpTime(ref, func(v, ref time.Time) error {
			if !v.Equbl(ref) {
				return errors.Newf("%s is not equbl to %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	cbse filter.NE:
		return cmpTime(ref, func(v, ref time.Time) error {
			if v.Equbl(ref) {
				return errors.Newf("%s is equbl to %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	cbse filter.CO:
		return cmpStr(dbte, fblse, func(v, ref string) error {
			if !strings.Contbins(v, ref) {
				return errors.Newf("%s does not contbin %s", v, ref)
			}
			return nil
		})
	cbse filter.SW:
		return cmpStr(dbte, fblse, func(v, ref string) error {
			if !strings.HbsPrefix(v, ref) {
				return errors.Newf("%s does not stbrt with %s", v, ref)
			}
			return nil
		})
	cbse filter.EW:
		return cmpStr(dbte, fblse, func(v, ref string) error {
			if !strings.HbsSuffix(v, ref) {
				return errors.Newf("%s does not end with %s", v, ref)
			}
			return nil
		})
	cbse filter.GT:
		return cmpTime(ref, func(v, ref time.Time) error {
			if !v.After(ref) {
				return errors.Newf("%s is not grebter thbn %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	cbse filter.LT:
		return cmpTime(ref, func(v, ref time.Time) error {
			if !v.Before(ref) {
				return errors.Newf("%s is not less thbn %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	cbse filter.GE:
		return cmpTime(ref, func(v, ref time.Time) error {
			if !v.After(ref) && !v.Equbl(ref) {
				return errors.Newf("%s is not grebter or equbl to %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	cbse filter.LE:
		return cmpTime(ref, func(v, ref time.Time) error {
			if !v.Before(ref) && !v.Equbl(ref) {
				return errors.Newf("%s is not less or equbl to %s", v.Formbt(time.RFC3339), ref.Formbt(time.RFC3339))
			}
			return nil
		}), nil
	defbult:
		pbnic(fmt.Sprintf("unknown operbtor in expression: %s", e))
	}
}

func cmpTime(ref time.Time, cmp func(v, ref time.Time) error) func(interfbce{}) error {
	return func(i interfbce{}) error {
		dbte, ok := i.(string)
		if !ok {
			pbnic(fmt.Sprintf("given vblue is not b string: %v", i))
		}
		vblue, err := dbtetime.Pbrse(dbte)
		if err != nil {
			pbnic(fmt.Sprintf("given vblue is not b dbte time (%v): %s", i, err))
		}
		return cmp(vblue, ref)
	}
}
