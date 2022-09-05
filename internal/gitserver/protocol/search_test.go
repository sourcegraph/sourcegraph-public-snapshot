package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryConstruction(t *testing.T) {
	t.Run("zero-length and is reduced", func(t *testing.T) {
		require.Equal(t, &Boolean{true}, NewAnd())
	})

	t.Run("single-element and is unwrapped", func(t *testing.T) {
		require.Equal(t, &Boolean{true}, NewAnd(&Boolean{true}))
	})

	t.Run("zero-length or is reduced", func(t *testing.T) {
		require.Equal(t, &Boolean{false}, NewOr())
	})

	t.Run("single-element or is unwrapped", func(t *testing.T) {
		require.Equal(t, &Boolean{true}, NewOr(&Boolean{true}))
	})

	t.Run("double negation cancels", func(t *testing.T) {
		require.Equal(t, &Boolean{true}, NewNot(NewNot(&Boolean{true})))
	})

	t.Run("nested and operators are flattened", func(t *testing.T) {
		input := NewAnd(
			NewAnd(&Boolean{true}, &Boolean{false}),
		)
		expected := &Operator{
			Kind: And,
			Operands: []Node{
				&Boolean{true},
				&Boolean{false},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("nested or operators are flattened", func(t *testing.T) {
		input := NewOr(
			NewOr(&Boolean{false}, &Boolean{true}),
		)
		expected := &Operator{
			Kind: Or,
			Operands: []Node{
				&Boolean{false},
				&Boolean{true},
			},
		}
		require.Equal(t, expected, input)
	})
}

func TestDistribute(t *testing.T) {
	am := func(expr string) *AuthorMatches {
		return &AuthorMatches{Expr: expr}
	}

	cases := []struct {
		input1 [][]Node
		input2 []Node
		output [][]Node
	}{
		{
			input1: [][]Node{{am("a"), am("b")}},
			input2: []Node{am("c")},
			output: [][]Node{
				{am("c"), am("a")},
				{am("c"), am("b")},
			},
		},
		{
			input1: [][]Node{
				{am("a"), am("b")},
				{am("c"), am("d")},
			},
			input2: []Node{am("e")},
			output: [][]Node{
				{am("e"), am("c"), am("a")},
				{am("e"), am("d"), am("a")},
				{am("e"), am("c"), am("b")},
				{am("e"), am("d"), am("b")},
			},
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.output, distribute(tc.input1, tc.input2))
		})
	}
}

type reducerTestCase struct {
	name   string
	input  Node
	output Node
}

func (tc *reducerTestCase) RunWithReducers(t *testing.T, reducers ...pass) {
	t.Run(tc.name, func(t *testing.T) {
		require.Equal(t, tc.output, ReduceWith(tc.input, reducers...))
	})
}

func TestReducers(t *testing.T) {
	am := func(expr string) *AuthorMatches {
		return &AuthorMatches{Expr: expr}
	}
	t.Run("propagateConst", func(t *testing.T) {
		cases := []reducerTestCase{
			{
				name:   "and with false is false",
				input:  newOperator(And, &Boolean{true}, &Boolean{false}),
				output: &Boolean{false},
			},
			{
				name:   "or with true is true",
				input:  newOperator(Or, &Boolean{true}, &Boolean{false}),
				output: &Boolean{true},
			},
			{
				name:   "false is removed from or",
				input:  newOperator(Or, am("a"), &Boolean{false}),
				output: newOperator(Or, am("a")),
			},
			{
				name:   "true is removed from and",
				input:  newOperator(And, am("a"), &Boolean{true}),
				output: newOperator(And, am("a")),
			},
			{
				name:   "and without constant is not affected",
				input:  newOperator(And, am("a"), am("b")),
				output: newOperator(And, am("a"), am("b")),
			},
			{
				name:   "negated constant is flattened",
				input:  newOperator(Not, &Boolean{false}),
				output: &Boolean{true},
			},
			{
				name:   "constant is propagated through or node",
				input:  newOperator(Or, newOperator(Not, &Boolean{false})),
				output: &Boolean{true},
			},
			{
				name:   "constant is propagated through and node",
				input:  newOperator(And, newOperator(Not, &Boolean{true})),
				output: &Boolean{false},
			},
		}

		for _, tc := range cases {
			tc.RunWithReducers(t, propagateBoolean)
		}
	})

	t.Run("rewriteConjunctive", func(t *testing.T) {
		cases := []reducerTestCase{
			{
				name:   "and with nested or is untouched",
				input:  newOperator(And, newOperator(Or, &Boolean{true})),
				output: newOperator(And, newOperator(Or, &Boolean{true})),
			},
			{
				name:   "or with nested and plus no siblings",
				input:  newOperator(Or, newOperator(And, &Boolean{true})),
				output: newOperator(And, newOperator(Or, &Boolean{true})),
			},
			{
				name:  "or with nested and plus one sibling",
				input: newOperator(Or, am("a"), newOperator(And, am("b"), am("c"))),
				output: newOperator(And,
					newOperator(Or, am("a"), am("b")),
					newOperator(Or, am("a"), am("c")),
				),
			},
			{
				name:  "or with nested and plus multiple siblings",
				input: newOperator(Or, am("a"), newOperator(And, am("b"), am("c")), am("d")),
				output: newOperator(And,
					newOperator(Or, am("a"), am("d"), am("b")),
					newOperator(Or, am("a"), am("d"), am("c")),
				),
			},
			{
				name: "or with multiple nested ands",
				input: newOperator(Or,
					am("a"),
					newOperator(And, am("b"), am("c")),
					newOperator(And, am("d"), am("e")),
				),
				output: newOperator(And,
					newOperator(Or, am("a"), am("d"), am("b")),
					newOperator(Or, am("a"), am("e"), am("b")),
					newOperator(Or, am("a"), am("d"), am("c")),
					newOperator(Or, am("a"), am("e"), am("c")),
				),
			},
		}

		for _, tc := range cases {
			tc.RunWithReducers(t, rewriteConjunctive)
		}
	})

	t.Run("flatten", func(t *testing.T) {
		cases := []reducerTestCase{
			{
				name:   "and with nested or is untouched",
				input:  newOperator(And, newOperator(Or, &Boolean{true})),
				output: newOperator(And, newOperator(Or, &Boolean{true})),
			},
			{
				name:   "and with nested and is merged",
				input:  newOperator(And, newOperator(And, &Boolean{true})),
				output: newOperator(And, &Boolean{true}),
			},
			{
				name:   "or with nested or is merged",
				input:  newOperator(Or, newOperator(Or, &Boolean{true})),
				output: newOperator(Or, &Boolean{true}),
			},
			{
				name:   "or with multiple nested or is merged",
				input:  newOperator(Or, newOperator(Or, &Boolean{true}), newOperator(Or, &Boolean{false})),
				output: newOperator(Or, &Boolean{true}, &Boolean{false}),
			},
			{
				name:   "and with multiple nested and is merged",
				input:  newOperator(And, newOperator(And, &Boolean{true}), newOperator(And, &Boolean{false})),
				output: newOperator(And, &Boolean{true}, &Boolean{false}),
			},
		}

		for _, tc := range cases {
			tc.RunWithReducers(t, flatten)
		}
	})

	t1 := time.Date(2020, 10, 11, 12, 12, 14, 14, time.UTC)
	t2 := time.Date(2021, 10, 11, 12, 12, 14, 14, time.UTC)

	t.Run("mergeOrRegexp", func(t *testing.T) {
		cases := []reducerTestCase{
			{
				name:   "authorMatches in and is not merged",
				input:  newOperator(And, &AuthorMatches{Expr: "a"}, &AuthorMatches{Expr: "b"}),
				output: newOperator(And, &AuthorMatches{Expr: "a"}, &AuthorMatches{Expr: "b"}),
			},
			{
				name:   "authorMatches in or is merged",
				input:  newOperator(Or, &AuthorMatches{Expr: "a"}, &AuthorMatches{Expr: "b"}),
				output: newOperator(Or, &AuthorMatches{Expr: "(?:a)|(?:b)"}),
			},
			{
				name:   "committerMatches in or is merged",
				input:  newOperator(Or, &CommitterMatches{Expr: "a"}, &CommitterMatches{Expr: "b"}),
				output: newOperator(Or, &CommitterMatches{Expr: "(?:a)|(?:b)"}),
			},
			{
				name:   "diffMatches in or is merged",
				input:  newOperator(Or, &DiffMatches{Expr: "a"}, &DiffMatches{Expr: "b"}),
				output: newOperator(Or, &DiffMatches{Expr: "(?:a)|(?:b)"}),
			},
			{
				name:   "diffModifiesFile in or is merged",
				input:  newOperator(Or, &DiffModifiesFile{Expr: "a"}, &DiffModifiesFile{Expr: "b"}),
				output: newOperator(Or, &DiffModifiesFile{Expr: "(?:a)|(?:b)"}),
			},
			{
				name:   "messageMatches in or is merged",
				input:  newOperator(Or, &MessageMatches{Expr: "a"}, &MessageMatches{Expr: "b"}),
				output: newOperator(Or, &MessageMatches{Expr: "(?:a)|(?:b)"}),
			},
			{
				name:   "unmergeable are not merged",
				input:  newOperator(Or, &CommitAfter{t1}, &CommitAfter{t2}),
				output: newOperator(Or, &CommitAfter{t1}, &CommitAfter{t2}),
			},
		}

		for _, tc := range cases {
			tc.RunWithReducers(t, mergeOrRegexp)
		}
	})

	t.Run("sortAndByCost", func(t *testing.T) {
		cases := []reducerTestCase{
			{
				name:   "stable for equal cost",
				input:  newOperator(And, &AuthorMatches{Expr: "a"}, &AuthorMatches{Expr: "b"}),
				output: newOperator(And, &AuthorMatches{Expr: "a"}, &AuthorMatches{Expr: "b"}),
			},
			{
				name:   "diff is placed last",
				input:  newOperator(And, &DiffMatches{Expr: "a"}, &AuthorMatches{Expr: "a"}),
				output: newOperator(And, &AuthorMatches{Expr: "a"}, &DiffMatches{Expr: "a"}),
			},
		}

		for _, tc := range cases {
			tc.RunWithReducers(t, sortAndByCost)
		}
	})
}
