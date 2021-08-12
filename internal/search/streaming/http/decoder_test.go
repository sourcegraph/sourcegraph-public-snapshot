package http

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	t.Parallel()

	type event struct {
		name string
		data string
	}

	decodeAll := func(input string) ([]event, error) {
		dec := NewDecoder(strings.NewReader(input))
		var events []event
		for dec.Scan() {
			events = append(events, event{
				name: string(dec.Event()),
				data: string(dec.Data()),
			})
		}
		return events, dec.Err()
	}

	t.Run("Single", func(t *testing.T) {
		events, err := decodeAll("event:a\ndata:b\n\n")
		require.NoError(t, err)
		require.Equal(t, events, []event{{name: "a", data: "b"}})
	})

	t.Run("Multiple", func(t *testing.T) {
		events, err := decodeAll("event:a\ndata:b\n\nevent:b\ndata:c\n\n")
		require.NoError(t, err)
		require.Equal(t, events, []event{{name: "a", data: "b"}, {name: "b", data: "c"}})
	})

	t.Run("ErrNoNewline", func(t *testing.T) {
		_, err := decodeAll("abc:a")
		require.Contains(t, err.Error(), "malformed event, no newline")
	})

	t.Run("ErrExpectedEvent", func(t *testing.T) {
		_, err := decodeAll("events:a\ndata:b\n\n")
		require.Contains(t, err.Error(), "malformed event, expected event")
	})

	t.Run("ErrExpectedData", func(t *testing.T) {
		_, err := decodeAll("event:a\ndatas:b\n\n")
		require.Contains(t, err.Error(), "malformed event event, expected data")
	})
}
