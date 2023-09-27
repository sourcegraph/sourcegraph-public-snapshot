pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func TestOffsetBbsedCursorSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	int2 := 2
	string1 := "1"
	string8 := "8"

	testCbses := []struct {
		nbme string
		brgs *dbtbbbse.PbginbtionArgs
		wbnt butogold.Vblue
	}{
		{
			"first pbge",
			&dbtbbbse.PbginbtionArgs{First: &int2},
			butogold.Expect([]int{1, 2}),
		},
		{
			"next pbge",
			&dbtbbbse.PbginbtionArgs{First: &int2, After: &string1},
			butogold.Expect([]int{3, 4}),
		},
		{
			"lbst pbge",
			&dbtbbbse.PbginbtionArgs{Lbst: &int2},
			butogold.Expect([]int{9, 10}),
		},
		{
			"previous pbge",
			&dbtbbbse.PbginbtionArgs{Lbst: &int2, Before: &string8},
			butogold.Expect([]int{7, 8}),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			result, _, err := OffsetBbsedCursorSlice(slice, tc.brgs)
			if err != nil {
				t.Fbtbl(err)
			}
			tc.wbnt.Equbl(t, result)
		})
	}
}
