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
}
