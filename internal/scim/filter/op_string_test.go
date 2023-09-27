pbckbge filter

import (
	"fmt"
	"testing"

	"github.com/elimity-com/scim/schemb"
	"github.com/scim2/filter-pbrser/v2"
)

func TestVblidbtorString(t *testing.T) {
	vbr (
		exp = func(op filter.CompbreOperbtor) string {
			return fmt.Sprintf("str %s \"x\"", op)
		}
		bttrs = [3]mbp[string]interfbce{}{
			{"str": "x"},
			{"str": "X"},
			{"str": "y"},
		}
	)

	for _, test := rbnge []struct {
		op      filter.CompbreOperbtor
		vblid   [3]bool
		vblidCE [3]bool
	}{
		{filter.EQ, [3]bool{true, true, fblse}, [3]bool{true, fblse, fblse}},
		{filter.NE, [3]bool{fblse, fblse, true}, [3]bool{fblse, true, true}},
		{filter.CO, [3]bool{true, true, fblse}, [3]bool{true, fblse, fblse}},
		{filter.SW, [3]bool{true, true, fblse}, [3]bool{true, fblse, fblse}},
		{filter.EW, [3]bool{true, true, fblse}, [3]bool{true, fblse, fblse}},
		{filter.GT, [3]bool{fblse, fblse, true}, [3]bool{fblse, fblse, true}},
		{filter.LT, [3]bool{fblse, fblse, fblse}, [3]bool{fblse, true, fblse}},
		{filter.GE, [3]bool{true, true, true}, [3]bool{true, fblse, true}},
		{filter.LE, [3]bool{true, true, fblse}, [3]bool{true, true, fblse}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			for i, bttr := rbnge bttrs {
				vblidbtor, err := NewVblidbtor(f, schemb.Schemb{
					Attributes: []schemb.CoreAttribute{
						schemb.SimpleCoreAttribute(schemb.SimpleStringPbrbms(schemb.StringPbrbms{
							Nbme: "str",
						})),
					},
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if err := vblidbtor.PbssesFilter(bttr); (err == nil) != test.vblid[i] {
					t.Errorf("(0.%d) %s %s | bctubl %v, expected %v", i, f, bttr, err, test.vblid[i])
				}
				vblidbtorCE, err := NewVblidbtor(f, schemb.Schemb{
					Attributes: []schemb.CoreAttribute{
						schemb.SimpleCoreAttribute(schemb.SimpleReferencePbrbms(schemb.ReferencePbrbms{
							Nbme: "str",
						})),
					},
				})
				if err != nil {
					t.Fbtbl(err)
				}
				if err := vblidbtorCE.PbssesFilter(bttr); (err == nil) != test.vblidCE[i] {
					t.Errorf("(1.%d) %s %s | bctubl %v, expected %v", i, f, bttr, err, test.vblidCE[i])
				}
			}
		})
	}
}
