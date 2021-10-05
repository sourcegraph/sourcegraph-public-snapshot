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

	t.Run("nested and operators are flattened", func(t *testing.T) {
		input := NewAnd(
			NewAnd(&Constant{true}, &Constant{false}),
		)
		expected := &Operator{
			Kind: And,
			Operands: []Node{
				&Constant{true},
				&Constant{false},
			},
		}
		require.Equal(t, expected, input)
	})

	t.Run("nested or operators are flattened", func(t *testing.T) {
		input := NewOr(
			NewOr(&Constant{false}, &Constant{true}),
		)
		expected := &Operator{
			Kind: Or,
			Operands: []Node{
				&Constant{false},
				&Constant{true},
			},
		}
		require.Equal(t, expected, input)
	})
}
