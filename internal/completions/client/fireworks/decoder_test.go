package fireworks

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	t.Parallel()

	type event struct {
		data string
	}

	decodeAll := func(input string) ([]event, error) {
		dec := NewDecoder(strings.NewReader(input))
		var events []event
		for dec.Scan() {
			events = append(events, event{
				data: string(dec.Data()),
			})
		}
		return events, dec.Err()
	}

	t.Run("Single", func(t *testing.T) {
		events, err := decodeAll("data:b\n\n")
		require.NoError(t, err)
		require.Equal(t, events, []event{{data: "b"}})
	})

	t.Run("Multiple", func(t *testing.T) {
		events, err := decodeAll("data:b\n\ndata:c\n\ndata: [DONE]\n\n")
		require.NoError(t, err)
		require.Equal(t, events, []event{{data: "b"}, {data: "c"}})
	})

	t.Run("ErrExpectedData", func(t *testing.T) {
		_, err := decodeAll("datas:b\n\n")
		require.Contains(t, err.Error(), "malformed data, expected data")
	})

	t.Run("Ends after done", func(t *testing.T) {
		events, err := decodeAll("data:b\n\ndata:c\n\ndata: [DONE]\n\ndata:d\n\n")
		require.NoError(t, err)
		require.Equal(t, events, []event{{data: "b"}, {data: "c"}})
	})
}
