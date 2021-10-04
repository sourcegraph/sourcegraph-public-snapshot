package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryConstruction(t *testing.T) {
	t.Run("zero-length and is reduced", func(t *testing.T) {
		require.Equal(t, &Constant{true}, NewAnd())
	})

	t.Run("single-element and is unwrapped", func(t *testing.T) {
		require.Equal(t, &Constant{true}, NewAnd(&Constant{true}))
	})

	t.Run("zero-length or is reduced", func(t *testing.T) {
		require.Equal(t, &Constant{false}, NewOr())
	})

	t.Run("single-element or is unwrapped", func(t *testing.T) {
		require.Equal(t, &Constant{true}, NewOr(&Constant{true}))
	})

	t.Run("double negation cancels", func(t *testing.T) {
		require.Equal(t, &Constant{true}, NewNot(NewNot(&Constant{true})))
	})

	t.Run("negation of and is pushed to atoms", func(t *testing.T) {
		input := NewNot(NewAnd(&Constant{true}, &Constant{false}))
		expected := &Operator{
			Kind: Or,
			Operands: []Node{
				&Operator{
					Kind:     Not,
					Operands: []Node{&Constant{true}},
				},
				&Operator{
					Kind:     Not,
					Operands: []Node{&Constant{false}},
				},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("negation of or is pushed to atoms", func(t *testing.T) {
		input := NewNot(NewOr(&Constant{true}, &Constant{false}))
		expected := &Operator{
			Kind: And,
			Operands: []Node{
				&Operator{
					Kind:     Not,
					Operands: []Node{&Constant{true}},
				},
				&Operator{
					Kind:     Not,
					Operands: []Node{&Constant{false}},
				},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("sibling and operators are merged", func(t *testing.T) {
		input := NewOr(
			NewAnd(&Constant{true}, &Constant{false}),
			&AuthorMatches{},
			NewAnd(&Constant{false}, &Constant{true}),
		)
		expected := &Operator{
			Kind: Or,
			Operands: []Node{
				&AuthorMatches{},
				&Operator{
					Kind: And,
					Operands: []Node{
						&Constant{true},
						&Constant{false},
						&Constant{false},
						&Constant{true},
					},
				},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("sibling or operators are merged", func(t *testing.T) {
		input := NewAnd(
			NewOr(&Constant{true}, &Constant{false}),
			&AuthorMatches{},
			NewOr(&Constant{false}, &Constant{true}),
		)
		expected := &Operator{
			Kind: And,
			Operands: []Node{
				&AuthorMatches{},
				&Operator{
					Kind: Or,
					Operands: []Node{
						&Constant{true},
						&Constant{false},
						&Constant{false},
						&Constant{true},
					},
				},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("nested and operators are flattened", func(t *testing.T) {
		input := NewAnd(
			NewOr(&Constant{false}, &Constant{true}),
			NewAnd(&Constant{true}, &Constant{false}),
			&AuthorMatches{},
		)
		expected := &Operator{
			Kind: And,
			Operands: []Node{
				&Constant{true},
				&Constant{false},
				&AuthorMatches{},
				&Operator{
					Kind: Or,
					Operands: []Node{
						&Constant{false},
						&Constant{true},
					},
				},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("nested or operators are flattened", func(t *testing.T) {
		input := NewOr(
			NewAnd(&Constant{true}, &Constant{false}),
			NewOr(&Constant{false}, &Constant{true}),
			&AuthorMatches{},
		)
		expected := &Operator{
			Kind: Or,
			Operands: []Node{
				&Constant{false},
				&Constant{true},
				&AuthorMatches{},
				&Operator{
					Kind: And,
					Operands: []Node{
						&Constant{true},
						&Constant{false},
					},
				},
			},
		}
		require.Equal(t, expected, input)
	})
}
