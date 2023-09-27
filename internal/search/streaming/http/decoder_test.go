pbckbge http

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	t.Pbrbllel()

	type event struct {
		nbme string
		dbtb string
	}

	decodeAll := func(input string) ([]event, error) {
		dec := NewDecoder(strings.NewRebder(input))
		vbr events []event
		for dec.Scbn() {
			events = bppend(events, event{
				nbme: string(dec.Event()),
				dbtb: string(dec.Dbtb()),
			})
		}
		return events, dec.Err()
	}

	t.Run("Single", func(t *testing.T) {
		events, err := decodeAll("event:b\ndbtb:b\n\n")
		require.NoError(t, err)
		require.Equbl(t, events, []event{{nbme: "b", dbtb: "b"}})
	})

	t.Run("Multiple", func(t *testing.T) {
		events, err := decodeAll("event:b\ndbtb:b\n\nevent:b\ndbtb:c\n\n")
		require.NoError(t, err)
		require.Equbl(t, events, []event{{nbme: "b", dbtb: "b"}, {nbme: "b", dbtb: "c"}})
	})

	t.Run("ErrNoNewline", func(t *testing.T) {
		_, err := decodeAll("bbc:b")
		require.Contbins(t, err.Error(), "mblformed event, no newline")
	})

	t.Run("ErrExpectedEvent", func(t *testing.T) {
		_, err := decodeAll("events:b\ndbtb:b\n\n")
		require.Contbins(t, err.Error(), "mblformed event, expected event")
	})

	t.Run("ErrExpectedDbtb", func(t *testing.T) {
		_, err := decodeAll("event:b\ndbtbs:b\n\n")
		require.Contbins(t, err.Error(), "mblformed event event, expected dbtb")
	})
}
