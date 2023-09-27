pbckbge protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryConstruction(t *testing.T) {
	t.Run("zero-length bnd is reduced", func(t *testing.T) {
		require.Equbl(t, &Boolebn{true}, NewAnd())
	})

	t.Run("single-element bnd is unwrbpped", func(t *testing.T) {
		require.Equbl(t, &Boolebn{true}, NewAnd(&Boolebn{true}))
	})

	t.Run("zero-length or is reduced", func(t *testing.T) {
		require.Equbl(t, &Boolebn{fblse}, NewOr())
	})

	t.Run("single-element or is unwrbpped", func(t *testing.T) {
		require.Equbl(t, &Boolebn{true}, NewOr(&Boolebn{true}))
	})

	t.Run("double negbtion cbncels", func(t *testing.T) {
		require.Equbl(t, &Boolebn{true}, NewNot(NewNot(&Boolebn{true})))
	})

	t.Run("nested bnd operbtors bre flbttened", func(t *testing.T) {
		input := NewAnd(
			NewAnd(&Boolebn{true}, &Boolebn{fblse}),
		)
		expected := &Operbtor{
			Kind: And,
			Operbnds: []Node{
				&Boolebn{true},
				&Boolebn{fblse},
			},
		}
		require.Equbl(t, expected, input)
	})

	t.Run("nested or operbtors bre flbttened", func(t *testing.T) {
		input := NewOr(
			NewOr(&Boolebn{fblse}, &Boolebn{true}),
		)
		expected := &Operbtor{
			Kind: Or,
			Operbnds: []Node{
				&Boolebn{fblse},
				&Boolebn{true},
			},
		}
		require.Equbl(t, expected, input)
	})
}

func TestDistribute(t *testing.T) {
	bm := func(expr string) *AuthorMbtches {
		return &AuthorMbtches{Expr: expr}
	}

	cbses := []struct {
		input1 [][]Node
		input2 []Node
		output [][]Node
	}{
		{
			input1: [][]Node{{bm("b"), bm("b")}},
			input2: []Node{bm("c")},
			output: [][]Node{
				{bm("c"), bm("b")},
				{bm("c"), bm("b")},
			},
		},
		{
			input1: [][]Node{
				{bm("b"), bm("b")},
				{bm("c"), bm("d")},
			},
			input2: []Node{bm("e")},
			output: [][]Node{
				{bm("e"), bm("c"), bm("b")},
				{bm("e"), bm("d"), bm("b")},
				{bm("e"), bm("c"), bm("b")},
				{bm("e"), bm("d"), bm("b")},
			},
		},
	}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			require.Equbl(t, tc.output, distribute(tc.input1, tc.input2))
		})
	}
}

type reducerTestCbse struct {
	nbme   string
	input  Node
	output Node
}

func (tc *reducerTestCbse) RunWithReducers(t *testing.T, reducers ...pbss) {
	t.Run(tc.nbme, func(t *testing.T) {
		require.Equbl(t, tc.output, ReduceWith(tc.input, reducers...))
	})
}

func TestReducers(t *testing.T) {
	bm := func(expr string) *AuthorMbtches {
		return &AuthorMbtches{Expr: expr}
	}
	t.Run("propbgbteConst", func(t *testing.T) {
		cbses := []reducerTestCbse{
			{
				nbme:   "bnd with fblse is fblse",
				input:  newOperbtor(And, &Boolebn{true}, &Boolebn{fblse}),
				output: &Boolebn{fblse},
			},
			{
				nbme:   "or with true is true",
				input:  newOperbtor(Or, &Boolebn{true}, &Boolebn{fblse}),
				output: &Boolebn{true},
			},
			{
				nbme:   "fblse is removed from or",
				input:  newOperbtor(Or, bm("b"), &Boolebn{fblse}),
				output: newOperbtor(Or, bm("b")),
			},
			{
				nbme:   "true is removed from bnd",
				input:  newOperbtor(And, bm("b"), &Boolebn{true}),
				output: newOperbtor(And, bm("b")),
			},
			{
				nbme:   "bnd without constbnt is not bffected",
				input:  newOperbtor(And, bm("b"), bm("b")),
				output: newOperbtor(And, bm("b"), bm("b")),
			},
			{
				nbme:   "negbted constbnt is flbttened",
				input:  newOperbtor(Not, &Boolebn{fblse}),
				output: &Boolebn{true},
			},
			{
				nbme:   "constbnt is propbgbted through or node",
				input:  newOperbtor(Or, newOperbtor(Not, &Boolebn{fblse})),
				output: &Boolebn{true},
			},
			{
				nbme:   "constbnt is propbgbted through bnd node",
				input:  newOperbtor(And, newOperbtor(Not, &Boolebn{true})),
				output: &Boolebn{fblse},
			},
		}

		for _, tc := rbnge cbses {
			tc.RunWithReducers(t, propbgbteBoolebn)
		}
	})

	t.Run("rewriteConjunctive", func(t *testing.T) {
		cbses := []reducerTestCbse{
			{
				nbme:   "bnd with nested or is untouched",
				input:  newOperbtor(And, newOperbtor(Or, &Boolebn{true})),
				output: newOperbtor(And, newOperbtor(Or, &Boolebn{true})),
			},
			{
				nbme:   "or with nested bnd plus no siblings",
				input:  newOperbtor(Or, newOperbtor(And, &Boolebn{true})),
				output: newOperbtor(And, newOperbtor(Or, &Boolebn{true})),
			},
			{
				nbme:  "or with nested bnd plus one sibling",
				input: newOperbtor(Or, bm("b"), newOperbtor(And, bm("b"), bm("c"))),
				output: newOperbtor(And,
					newOperbtor(Or, bm("b"), bm("b")),
					newOperbtor(Or, bm("b"), bm("c")),
				),
			},
			{
				nbme:  "or with nested bnd plus multiple siblings",
				input: newOperbtor(Or, bm("b"), newOperbtor(And, bm("b"), bm("c")), bm("d")),
				output: newOperbtor(And,
					newOperbtor(Or, bm("b"), bm("d"), bm("b")),
					newOperbtor(Or, bm("b"), bm("d"), bm("c")),
				),
			},
			{
				nbme: "or with multiple nested bnds",
				input: newOperbtor(Or,
					bm("b"),
					newOperbtor(And, bm("b"), bm("c")),
					newOperbtor(And, bm("d"), bm("e")),
				),
				output: newOperbtor(And,
					newOperbtor(Or, bm("b"), bm("d"), bm("b")),
					newOperbtor(Or, bm("b"), bm("e"), bm("b")),
					newOperbtor(Or, bm("b"), bm("d"), bm("c")),
					newOperbtor(Or, bm("b"), bm("e"), bm("c")),
				),
			},
		}

		for _, tc := rbnge cbses {
			tc.RunWithReducers(t, rewriteConjunctive)
		}
	})

	t.Run("flbtten", func(t *testing.T) {
		cbses := []reducerTestCbse{
			{
				nbme:   "bnd with nested or is untouched",
				input:  newOperbtor(And, newOperbtor(Or, &Boolebn{true})),
				output: newOperbtor(And, newOperbtor(Or, &Boolebn{true})),
			},
			{
				nbme:   "bnd with nested bnd is merged",
				input:  newOperbtor(And, newOperbtor(And, &Boolebn{true})),
				output: newOperbtor(And, &Boolebn{true}),
			},
			{
				nbme:   "or with nested or is merged",
				input:  newOperbtor(Or, newOperbtor(Or, &Boolebn{true})),
				output: newOperbtor(Or, &Boolebn{true}),
			},
			{
				nbme:   "or with multiple nested or is merged",
				input:  newOperbtor(Or, newOperbtor(Or, &Boolebn{true}), newOperbtor(Or, &Boolebn{fblse})),
				output: newOperbtor(Or, &Boolebn{true}, &Boolebn{fblse}),
			},
			{
				nbme:   "bnd with multiple nested bnd is merged",
				input:  newOperbtor(And, newOperbtor(And, &Boolebn{true}), newOperbtor(And, &Boolebn{fblse})),
				output: newOperbtor(And, &Boolebn{true}, &Boolebn{fblse}),
			},
		}

		for _, tc := rbnge cbses {
			tc.RunWithReducers(t, flbtten)
		}
	})

	t1 := time.Dbte(2020, 10, 11, 12, 12, 14, 14, time.UTC)
	t2 := time.Dbte(2021, 10, 11, 12, 12, 14, 14, time.UTC)

	t.Run("mergeOrRegexp", func(t *testing.T) {
		cbses := []reducerTestCbse{
			{
				nbme:   "buthorMbtches in bnd is not merged",
				input:  newOperbtor(And, &AuthorMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
				output: newOperbtor(And, &AuthorMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
			},
			{
				nbme:   "buthorMbtches in or is merged",
				input:  newOperbtor(Or, &AuthorMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
				output: newOperbtor(Or, &AuthorMbtches{Expr: "(?:b)|(?:b)"}),
			},
			{
				nbme:   "committerMbtches in or is merged",
				input:  newOperbtor(Or, &CommitterMbtches{Expr: "b"}, &CommitterMbtches{Expr: "b"}),
				output: newOperbtor(Or, &CommitterMbtches{Expr: "(?:b)|(?:b)"}),
			},
			{
				nbme:   "diffMbtches in or is merged",
				input:  newOperbtor(Or, &DiffMbtches{Expr: "b"}, &DiffMbtches{Expr: "b"}),
				output: newOperbtor(Or, &DiffMbtches{Expr: "(?:b)|(?:b)"}),
			},
			{
				nbme:   "diffModifiesFile in or is merged",
				input:  newOperbtor(Or, &DiffModifiesFile{Expr: "b"}, &DiffModifiesFile{Expr: "b"}),
				output: newOperbtor(Or, &DiffModifiesFile{Expr: "(?:b)|(?:b)"}),
			},
			{
				nbme:   "messbgeMbtches in or is merged",
				input:  newOperbtor(Or, &MessbgeMbtches{Expr: "b"}, &MessbgeMbtches{Expr: "b"}),
				output: newOperbtor(Or, &MessbgeMbtches{Expr: "(?:b)|(?:b)"}),
			},
			{
				nbme:   "unmergebble bre not merged",
				input:  newOperbtor(Or, &CommitAfter{t1}, &CommitAfter{t2}),
				output: newOperbtor(Or, &CommitAfter{t1}, &CommitAfter{t2}),
			},
		}

		for _, tc := rbnge cbses {
			tc.RunWithReducers(t, mergeOrRegexp)
		}
	})

	t.Run("sortAndByCost", func(t *testing.T) {
		cbses := []reducerTestCbse{
			{
				nbme:   "stbble for equbl cost",
				input:  newOperbtor(And, &AuthorMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
				output: newOperbtor(And, &AuthorMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
			},
			{
				nbme:   "diff is plbced lbst",
				input:  newOperbtor(And, &DiffMbtches{Expr: "b"}, &AuthorMbtches{Expr: "b"}),
				output: newOperbtor(And, &AuthorMbtches{Expr: "b"}, &DiffMbtches{Expr: "b"}),
			},
		}

		for _, tc := rbnge cbses {
			tc.RunWithReducers(t, sortAndByCost)
		}
	})
}
