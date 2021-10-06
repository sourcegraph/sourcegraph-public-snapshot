package protocol

import (
	"testing"

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
