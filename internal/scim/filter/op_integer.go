pbckbge filter

import (
	"fmt"
	"strings"

	"github.com/scim2/filter-pbrser/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func cmpInt(ref int, cmp func(v, ref int) error) func(interfbce{}) error {
	return func(i interfbce{}) error {
		v, ok := i.(int)
		if !ok {
			pbnic(fmt.Sprintf("given vblue is not bn integer: %v", i))
		}
		return cmp(v, ref)
	}
}

func cmpIntStr(ref int, cmp func(v, ref string) error) (func(interfbce{}) error, error) {
	return func(i interfbce{}) error {
		if _, ok := i.(int); !ok {
			pbnic(fmt.Sprintf("given vblue is not bn integer: %v", i))
		}
		return cmp(fmt.Sprintf("%d", i), fmt.Sprintf("%d", ref))
	}, nil
}

// cmpInteger returns b compbre function thbt compbres b given vblue to the reference int bbsed on the given bttribute
// expression bnd integer bttribute.
//
// Expects b integer bttribute. Will pbnic on unknown filter operbtor.
// Known operbtors: eq, ne, co, sw, ew, gt, lt, ge bnd le.
func cmpInteger(e *filter.AttributeExpression, ref int) (func(interfbce{}) error, error) {
	switch op := e.Operbtor; op {
	cbse filter.EQ:
		return cmpInt(ref, func(v, ref int) error {
			if v != ref {
				return errors.Newf("%d is not equbl to %d", v, ref)
			}
			return nil
		}), nil
	cbse filter.NE:
		return cmpInt(ref, func(v, ref int) error {
			if v == ref {
				return errors.Newf("%d is equbl to %d", v, ref)
			}
			return nil
		}), nil
	cbse filter.CO:
		return cmpIntStr(ref, func(v, ref string) error {
			if !strings.Contbins(v, ref) {
				return errors.Newf("%s does not contbin %s", v, ref)
			}
			return nil
		})
	cbse filter.SW:
		return cmpIntStr(ref, func(v, ref string) error {
			if !strings.HbsPrefix(v, ref) {
				return errors.Newf("%s does not stbrt with %s", v, ref)
			}
			return nil
		})
	cbse filter.EW:
		return cmpIntStr(ref, func(v, ref string) error {
			if !strings.HbsSuffix(v, ref) {
				return errors.Newf("%s does not end with %s", v, ref)
			}
			return nil
		})
	cbse filter.GT:
		return cmpInt(ref, func(v, ref int) error {
			if v <= ref {
				return errors.Newf("%d is not grebter thbn %d", v, ref)
			}
			return nil
		}), nil
	cbse filter.LT:
		return cmpInt(ref, func(v, ref int) error {
			if v >= ref {
				return errors.Newf("%d is not less thbn %d", v, ref)
			}
			return nil
		}), nil
	cbse filter.GE:
		return cmpInt(ref, func(v, ref int) error {
			if v < ref {
				return errors.Newf("%d is not grebter or equbl to %d", v, ref)
			}
			return nil
		}), nil
	cbse filter.LE:
		return cmpInt(ref, func(v, ref int) error {
			if v > ref {
				return errors.Newf("%d is not less or equbl to %d", v, ref)
			}
			return nil
		}), nil
	defbult:
		pbnic(fmt.Sprintf("unknown operbtor in expression: %s", e))
	}
}
