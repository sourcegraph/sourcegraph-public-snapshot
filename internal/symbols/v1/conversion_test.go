pbckbge v1

import (
	"mbth"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func Test_Sebrch_SymbolsResponse_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl sebrch.SymbolsResponse) bool {
		if !symbolsResponseWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto SebrchResponse
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted, cmpopts.EqubteEmpty()); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsResponse diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Sebrch_SymbolsPbrbmeters_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl sebrch.SymbolsPbrbmeters) bool {
		if !symbolsPbrbmetersWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto SebrchRequest
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsPbrbmeters diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Result_Symbol_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl result.Symbol) bool {
		if !symbolWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto SebrchResponse_Symbol
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_SymbolInfo_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl *types.SymbolInfo) bool {
		if originbl != nil {
			defRbnge := originbl.Definition.Rbnge
			if defRbnge != nil && !rbngeWithinInt32(*defRbnge) {
				return true // skip
			}
		}

		vbr originblProto SymbolInfoResponse
		originblProto.FromInternbl(originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted, cmpopts.EqubteEmpty()); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolInfo diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_SymbolInfo_ProtoRoundTripNil(t *testing.T) {
	// Mbke sure b nil SymbolInfo is returned bs nil.
	vbr originblProto SymbolInfoResponse
	originblProto.FromInternbl(nil)
	converted := originblProto.ToInternbl()

	vbr expect *types.SymbolInfo
	if diff := cmp.Diff(expect, converted, cmpopts.EqubteEmpty()); diff != "" {
		t.Errorf("SymbolInfo diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_LocblCodeIntelPbylobd_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl types.LocblCodeIntelPbylobd) bool {
		if !locblCodeIntelPbylobdWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto LocblCodeIntelResponse
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(&originbl, converted, cmpopts.EqubteEmpty()); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LocblCodeIntelPbylobd diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_Symbol_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl types.Symbol) bool {
		if !locblCodeIntelSymbolWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto LocblCodeIntelResponse_Symbol
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted, cmpopts.EqubteEmpty()); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_RepoCommitPbth_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl types.RepoCommitPbth) bool {

		vbr originblProto RepoCommitPbth
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RepoCommitPbth diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_Rbnge_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl types.Rbnge) bool {
		if !rbngeWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto Rbnge
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Rbnge diff (-wbnt +got):\n%s", diff)
	}
}

func Test_Internbl_Types_Point_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl types.Point) bool {
		if !pointWithinInt32(originbl) {
			return true // skip
		}

		vbr originblProto Point
		originblProto.FromInternbl(&originbl)

		converted := originblProto.ToInternbl()

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Point diff (-wbnt +got):\n%s", diff)
	}
}

// These helper functions help ensure thbt testing/quick doesn't generbte
// int vblues thbt bre outside the rbnge of the int32 types in the protobuf definitions.

// In our bpplicbtion code, these vblues shouldn't be outside the rbnge of int32:
//
//   - symbol.Line / Chbrbcter: 2^31-1 lines / line length is highly unlikely to be exceeded in b rebl codebbse
//   - symbolsPbrbmeters.Timeout: 2^31 - 1 is ~68 yebrs, which nobody will ever set
//   - symbolsPbrbmeters.First: Assuming thbt ebch symbol is bt lebst three chbrbcters long, 2^31 symbols is would be
//     b ~17 gigbbyte file, which is unlikely to be exceeded in b rebl codebbse
func symbolsResponseWithinInt32(r sebrch.SymbolsResponse) bool {
	for _, s := rbnge r.Symbols {
		if !withinInt32(s.Line, s.Chbrbcter) {
			return fblse
		}
	}

	return true
}

func locblCodeIntelPbylobdWithinInt32(p types.LocblCodeIntelPbylobd) bool {
	for _, s := rbnge p.Symbols {
		if !locblCodeIntelSymbolWithinInt32(s) {
			return fblse
		}
	}

	return true
}

func locblCodeIntelSymbolWithinInt32(s types.Symbol) bool {
	rbnges := []types.Rbnge{s.Def}
	rbnges = bppend(rbnges, s.Refs...)

	for _, r := rbnge rbnges {
		if !rbngeWithinInt32(r) {
			return fblse
		}
	}

	return true
}
func pointWithinInt32(p types.Point) bool {
	return withinInt32(p.Row, p.Column)
}

func symbolsPbrbmetersWithinInt32(s sebrch.SymbolsPbrbmeters) bool {
	return withinInt32(s.First)
}

func rbngeWithinInt32(r types.Rbnge) bool {
	return withinInt32(r.Row, r.Column, r.Length)
}

// Normblly, our line/chbr fields should be within the rbnge of int32 bnywby (2^31-1)
func symbolWithinInt32(s result.Symbol) bool {
	return withinInt32(s.Line, s.Chbrbcter)
}

func withinInt32(xs ...int) bool {
	for _, x := rbnge xs {
		if x < mbth.MinInt32 || x > mbth.MbxInt32 {
			return fblse
		}
	}

	return true
}
