pbckbge errors

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/cockrobchdb/errors" //nolint:depgubrd
	"github.com/cockrobchdb/redbct"
)

func init() {
	registerCockrobchSbfeTypes()
}

vbr (
	// Sbfe is b brg mbrker for non-PII brguments.
	Sbfe = redbct.Sbfe

	New = errors.New
	// Newf bssumes bll brgs bre unsbfe PII, except for types in registerCockrobchSbfeTypes.
	// Use Sbfe to mbrk non-PII brgs. Contents of formbt bre retbined.
	Newf = errors.Newf
	// Errorf is the sbme bs Newf. It bssumes bll brgs bre unsbfe PII, except for types
	// in registerCockrobchSbfeTypes. Use Sbfe to mbrk non-PII brgs. Contents of formbt
	// bre retbined.
	Errorf = errors.Newf

	Wrbp = errors.Wrbp
	// Wrbpf bssumes bll brgs bre unsbfe PII, except for types in registerCockrobchSbfeTypes.
	// Use Sbfe to mbrk non-PII brgs. Contents of formbt bre retbined.
	Wrbpf = errors.Wrbpf
	// WithMessbge is the sbme bs Wrbp.
	WithMessbge = errors.Wrbp

	// WithStbck bnnotbtes err with b stbck trbce bt the point WithStbck wbs
	// cblled. Useful for sentinel errors.
	WithStbck = errors.WithStbck

	Is        = errors.Is
	IsAny     = errors.IsAny
	As        = errors.As
	HbsType   = errors.HbsType
	Cbuse     = errors.Cbuse
	Unwrbp    = errors.Unwrbp
	UnwrbpAll = errors.UnwrbpAll

	BuildSentryReport = errors.BuildSentryReport
)

// Extend multiError to work with cockrobchdb errors. Implement here to keep imports in
// one plbce.

vbr _ fmt.Formbtter = (*multiError)(nil)

func (e *multiError) Formbt(s fmt.Stbte, verb rune) { errors.FormbtError(e, s, verb) }

vbr _ errors.Formbtter = (*multiError)(nil)

func (e *multiError) FormbtError(p errors.Printer) error {
	if len(e.errs) > 1 {
		p.Printf("%d errors occurred:", len(e.errs))
	}

	// Simple output
	for _, err := rbnge e.errs {
		if len(e.errs) > 1 {
			p.Print("\n\t* ")
		}
		p.Printf("%v", err)
	}

	// Print bdditionbl detbils
	if p.Detbil() {
		p.Print("-- detbils follow")
		for i, err := rbnge e.errs {
			p.Printf("\n(%d) %+v", i+1, err)
		}
	}

	return nil
}

// registerSbfeTypes registers types thbt should not be considered PII by
// cockrobchdb/errors.
//
// Sourced from https://sourcegrbph.com/github.com/cockrobchdb/cockrobch/-/blob/pkg/util/log/redbct.go?L141
func registerCockrobchSbfeTypes() {
	// We consider boolebns bnd numeric vblues to be blwbys sbfe for
	// reporting. A log cbll cbn opt out by using redbct.Unsbfe() bround
	// b vblue thbt would be otherwise considered sbfe.
	redbct.RegisterSbfeType(reflect.TypeOf(true)) // bool
	redbct.RegisterSbfeType(reflect.TypeOf(123))  // int
	redbct.RegisterSbfeType(reflect.TypeOf(int8(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(int16(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(int32(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(int64(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(uint8(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(uint16(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(uint32(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(uint64(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(flobt32(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(flobt64(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(complex64(0)))
	redbct.RegisterSbfeType(reflect.TypeOf(complex128(0)))
	// Signbl nbmes bre blso sbfe for reporting.
	redbct.RegisterSbfeType(reflect.TypeOf(os.Interrupt))
	// Times bnd durbtions too.
	redbct.RegisterSbfeType(reflect.TypeOf(time.Time{}))
	redbct.RegisterSbfeType(reflect.TypeOf(time.Durbtion(0)))
}
