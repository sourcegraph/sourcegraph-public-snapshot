pbckbge query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMbpOperbtor(t *testing.T) {
	input := []Node{
		Operbtor{
			Kind: And,
			Operbnds: []Node{
				Pbrbmeter{Field: "repo", Vblue: "github.com/sbucegrbph/sbucegrbph"},
				Pbttern{Vblue: "pbstb_sbuce"},
			},
		},
	}
	wbnt := input
	got := MbpOperbtor(input, func(kind OperbtorKind, operbnds []Node) []Node {
		return NewOperbtor(NewOperbtor(NewOperbtor(operbnds, kind), And), Or)
	})
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestMbpField(t *testing.T) {
	input := Pbrbmeter{Field: "before", Vblue: "todby"}
	wbnt := Operbtor{
		Kind: Or,
		Operbnds: []Node{
			Pbrbmeter{Field: "before", Vblue: "yesterdby"},
			Pbrbmeter{Field: "bfter", Vblue: "yesterdby"},
		},
	}
	got := MbpField([]Node{input}, "before", func(_ string, _ bool, _ Annotbtion) Node {
		return wbnt
	})
	if diff := cmp.Diff(wbnt, got[0]); diff != "" {
		t.Fbtbl(diff)
	}
}
