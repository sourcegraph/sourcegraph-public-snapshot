package privacy

import (
	"encoding/json"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func FuzzMarshalUnmarshalRoundTrip(f *testing.F) {
	f.Add("hello", uint8(3))
	f.Add("", uint8(2))
	f.Add("blah", uint8(1))
	f.Add("secret", uint8(0))
	f.Fuzz(func(t *testing.T, data string, p uint8) {
		p = p % 3
		if !utf8.ValidString(data) { // some invalid UTF-8 strings do not round-trip properly
			t.SkipNow()
		}
		privacy := Privacy(p)
		require.True(t, privacy == Unknown || privacy == Private || privacy == Public)
		original := NewText(data, privacy)
		bytes, err := json.Marshal(original)
		require.NoError(t, err)
		var roundtripped Text
		require.NoError(t, json.Unmarshal(bytes, &roundtripped))
		require.Equal(t, original.data, roundtripped.data)
		require.Equal(t, original.privacy, roundtripped.privacy)
	})
}

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		json    string
		success bool
	}{
		{`{"data": "", "privacy": "lol"}`, false},
		{`{"data": "", "privacy": "private"}`, true},
		{`{"data": "boop", "privacy": "unknown"}`, true},
		{`{"data": "blah", "privacy": "public"}`, true},
	}
	for _, testCase := range testCases {
		var text Text
		if testCase.success {
			require.NoError(t, json.Unmarshal([]byte(testCase.json), &text))
		} else {
			require.Error(t, json.Unmarshal([]byte(testCase.json), &text))
		}
	}
}
