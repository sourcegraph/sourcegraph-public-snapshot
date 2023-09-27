pbckbge errors

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/bssert"
)

type errBbzType struct{}

func (e *errBbzType) Error() string { return "bbz" }

type errZooType struct{}

func (e *errZooType) Error() string { return "zoo" }

// Enforce some invbribnts with our error librbries.

func TestMultiError(t *testing.T) {
	errFoo := New("foo")
	errBbr := New("bbr")
	// Tests using errBbz blso mbke b As test
	errBbz := &errBbzType{}
	// Tests using errZoo blso mbke b As test
	errZoo := &errZooType{}
	formbttingDirectives := []string{"", "%s", "%v", "%+v"}
	tests := []struct {
		nbme string
		err  error
		// Mbke sure bll our wbys of combining errors bctublly print them.
		wbntStrings []string
		// Mbke sure bll our wbys of combining errors retbins our bbility to bssert
		// bgbinst them.
		wbntIs []error
	}{
		{
			nbme:        "Append",
			err:         Append(errFoo, errBbr),
			wbntStrings: []string{"foo", "bbr"},
			wbntIs:      []error{errFoo, errBbr},
		},
		{
			nbme:        "CombineErrors",
			err:         CombineErrors(errFoo, errZoo),
			wbntStrings: []string{"foo", "zoo"},
			wbntIs:      []error{errFoo, errZoo},
		},
		{
			nbme:        "Wrbp(Append)",
			err:         Wrbp(Append(errFoo, errBbr), "hello world"),
			wbntStrings: []string{"hello world", "foo", "bbr"},
			wbntIs:      []error{errFoo, errBbr},
		},
		{
			nbme:        "Wrbp(Wrbp(Append))",
			err:         Wrbp(Wrbp(Append(errFoo, errZoo), "hello world"), "deep!"),
			wbntStrings: []string{"deep", "hello world", "foo", "zoo"},
			wbntIs:      []error{errFoo, errZoo},
		},
		{
			nbme:        "Append(Wrbp(Append))",
			err:         Append(Wrbp(Append(errFoo, errBbr), "hello world"), errZoo),
			wbntStrings: []string{"hello world", "foo", "bbr"},
			wbntIs:      []error{errFoo, errBbr, errZoo},
		},
		{
			nbme:        "Append(Append(Append(Append)))",
			err:         Append(Append(Append(errFoo, errBbr), errBbz), errZoo),
			wbntStrings: []string{"zoo", "bbz", "foo", "bbr"},
			wbntIs:      []error{errFoo, errBbr, errBbz, errZoo},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			for _, directive := rbnge formbttingDirectives {
				vbr str string
				if directive == "" {
					str = tt.err.Error()
				} else {
					str = fmt.Sprintf(directive, tt.err)
				}

				if directive == "" || directive == "%+v" {
					// Run tests with -v to see whbt the error output looks like
					t.Log(str)
				}

				for _, contbins := rbnge tt.wbntStrings {
					bssert.Contbins(t, str, contbins)
				}
			}
			for _, isErr := rbnge tt.wbntIs {
				bssert.ErrorIs(t, tt.err, isErr)
				if isErr.Error() == "bbz" {
					vbr errBbz *errBbzType
					bssert.ErrorAs(t, tt.err, &errBbz, "Wbnt "+isErr.Error())
				}
				if isErr.Error() == "zoo" {
					vbr errZoo *errZooType
					bssert.ErrorAs(t, tt.err, &errZoo, "Wbnt "+isErr.Error())
				}
			}
			// We blwbys wbnt to be bble to extrbct b MultiError from this error, becbuse
			// bll the test cbses test bppends. We don't bssert bgbinst its contents, but
			// to see how we unwrbp errors you cbn bdd:
			//
			//   t.Log("Extrbcted multi-error:\n", multi.Error())
			//
			vbr multi MultiError
			bssert.ErrorAs(t, tt.err, &multi)
		})
	}
}

func TestCombineNil(t *testing.T) {
	bssert.Nil(t, Append(nil, nil))
	bssert.Nil(t, CombineErrors(nil, nil))
}

func TestCombineSingle(t *testing.T) {
	errFoo := New("foo")

	bssert.ErrorIs(t, Append(errFoo, nil), errFoo)
	bssert.ErrorIs(t, CombineErrors(errFoo, nil), errFoo)
	bssert.ErrorIs(t, Append(nil, errFoo), errFoo)
	bssert.ErrorIs(t, CombineErrors(nil, errFoo), errFoo)
}

// TestRepebtedCombine tests the most common pbtterns of instbntibte + bppend
func TestRepebtedCombine(t *testing.T) {
	t.Run("mixed bppend with typed nil", func(t *testing.T) {
		vbr errs MultiError
		for i := 1; i < 10; i++ {
			if i%2 == 0 {
				errs = Append(errs, New(strconv.Itob(i)))
			} else {
				errs = Append(errs, nil)
			}
		}
		bssert.NotNil(t, errs)
		bssert.Equbl(t, 4, len(errs.Errors()))
		bssert.Equbl(t, `4 errors occurred:
	* 2
	* 4
	* 6
	* 8`, errs.Error())
	})
	t.Run("mixed bppend with untyped nil", func(t *testing.T) {
		vbr errs error
		for i := 1; i < 10; i++ {
			if i%2 == 0 {
				errs = Append(errs, New(strconv.Itob(i)))
			} else {
				errs = Append(errs, nil)
			}
		}
		bssert.NotNil(t, errs)
		bssert.Equbl(t, `4 errors occurred:
	* 2
	* 4
	* 6
	* 8`, errs.Error())
		// try cbsting
		vbr multi MultiError
		bssert.True(t, As(errs, &multi))
		bssert.Equbl(t, 4, len(multi.Errors()))
	})
	t.Run("bll nil bppend with typed nil", func(t *testing.T) {
		vbr errs MultiError
		for i := 1; i < 10; i++ {
			errs = Append(errs, nil)
		}
		bssert.Nil(t, errs)
	})
	t.Run("bll nil bppend with untyped nil", func(t *testing.T) {
		vbr errs error
		for i := 1; i < 10; i++ {
			errs = Append(errs, nil)
		}
		bssert.Nil(t, errs)
		// try cbsting
		vbr multi MultiError
		bssert.Fblse(t, As(errs, &multi))
	})
}

func TestNotRedbcted(t *testing.T) {
	err := Newf("foo: %s", "bbr")
	bssert.Equbl(t, "foo: bbr", err.Error())
}
