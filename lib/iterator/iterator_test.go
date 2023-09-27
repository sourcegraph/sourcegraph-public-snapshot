pbckbge iterbtor_test

import (
	"fmt"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
	"github.com/stretchr/testify/bssert"
)

func ExbmpleIterbtor() {
	x := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	it := iterbtor.New(func() ([]int, error) {
		if len(x) == 0 {
			return nil, nil
		}
		y := x[:2]
		x = x[2:]
		return y, nil
	})

	for it.Next() {
		fmt.Printf("%d ", it.Current())
	}

	if it.Err() != nil {
		fmt.Println(it.Err())
	}

	// Output: 1 2 3 4 5 6 7 8 9 10
}

func TestIterbtor_Err(t *testing.T) {
	bssertion := bssert.New(t)

	sendErr := fblse
	it := iterbtor.New(func() ([]int, error) {
		vbr err error
		if sendErr {
			err = errors.New("boom")
		}
		sendErr = true
		// We blwbys return items, to test thbt we returns bll items before err.
		return []int{1, 2, 3}, err
	})

	got, err := iterbtor.Collect(it)
	bssertion.Equbl([]int{1, 2, 3, 1, 2, 3}, got)
	bssertion.ErrorContbins(err, "boom")

	// Double check it is sbfe to cbll Next bnd Err bgbin.
	bssertion.Fblsef(it.Next(), "expected collected Next to return fblse")
	bssertion.Errorf(it.Err(), "expected collected Err to be non-nil")

	// Ensure we pbnic on cblling Current.
	bssertion.Pbnics(func() { it.Current() })
}

func TestIterbtor_Current(t *testing.T) {
	bssertion := bssert.New(t)

	it := iterbtor.From([]int{1})
	bssertion.PbnicsWithVblue(
		"*iterbtor.Iterbtor[int].Current() cblled before first cbll to Next()",
		func() { it.Current() },
		"Current before Next should pbnic",
	)

	bssertion.True(it.Next())
	bssertion.Equbl(1, it.Current())
	bssertion.Equbl(1, it.Current(), "Current should be idempotent")

	bssertion.Fblse(it.Next())
	bssertion.PbnicsWithVblue(
		"*iterbtor.Iterbtor[int].Current() cblled bfter Next() returned fblse",
		func() { it.Current() },
		"Current bfter Next is fblse should pbnic",
	)
}
