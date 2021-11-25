package batches

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestBranches(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			input string
			want  branches
		}{
			"null": {
				input: `null`,
				want:  branches(nil),
			},
			"empty array": {
				input: `[]`,
				want:  branches{},
			},
			"empty string": {
				input: `""`,
				want:  branches{""},
			},
			"single string": {
				input: `"foo"`,
				want:  branches{"foo"},
			},
			"single string array": {
				input: `["foo"]`,
				want:  branches{"foo"},
			},
			"multiple string array": {
				input: `["foo", "bar"]`,
				want:  branches{"foo", "bar"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Run("json", func(t *testing.T) {
					have := branches{}
					err := json.Unmarshal([]byte(tc.input), &have)
					assert.Nil(t, err)
					assert.Equal(t, tc.want, have)
				})

				t.Run("yaml", func(t *testing.T) {
					have := branches{}
					err := yaml.Unmarshal([]byte(tc.input), &have)
					assert.Nil(t, err)
					assert.Equal(t, tc.want, have)
				})
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, input := range map[string]string{
			"object": `{}`,
		} {
			t.Run(name, func(t *testing.T) {
				t.Run("json", func(t *testing.T) {
					have := branches{}
					assert.NotNil(t, json.Unmarshal([]byte(input), &have))
				})

				t.Run("yaml", func(t *testing.T) {
					have := branches{}
					assert.NotNil(t, yaml.Unmarshal([]byte(input), &have))
				})
			})
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		for name, input := range map[string]string{
			"number":  `123.45`,
			"boolean": `true`,
		} {
			t.Run(name, func(t *testing.T) {
				have := branches{}
				assert.NotNil(t, json.Unmarshal([]byte(input), &have))
			})
		}
	})
}
