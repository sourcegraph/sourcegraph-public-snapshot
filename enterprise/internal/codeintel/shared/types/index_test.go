package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIndex_UnmarshalJSON(t *testing.T) {
	tts := []struct {
		input string
		want  Index
	}{
		{
			input: `{"requestedEnvVars": ["foobar"]}`,
			want:  Index{RequestedEnvVars: []string{"foobar"}},
		},
		{
			input: `{"requested_env_vars": ["foobar"]}`,
			want:  Index{RequestedEnvVars: []string{"foobar"}},
		},
		{
			input: `{"requestedEnvVars": ["foobar"],"requested_env_vars": ["barfoo"]}`,
			want:  Index{RequestedEnvVars: []string{"barfoo"}},
		},
	}

	for _, tt := range tts {
		var have Index
		if err := json.Unmarshal([]byte(tt.input), &have); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(tt.want, have); diff != "" {
			t.Fatalf("unexpected index from json %s", diff)
		}
	}
}
