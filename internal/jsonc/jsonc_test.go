package jsonc

import (
	"reflect"
	"strconv"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{
			input: `{
// comment
/* another comment */
"hello": "world",
}`,
			want: map[string]any{"hello": "world"},
		},
		{
			input: `// just
		// comments
		// here`,
			want: nil,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var got any
			if err := Unmarshal(test.input, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{
			input: `{
// comment
/* another comment */
"hello": "world",
}`,
			want: `{"hello":"world"}`,
		},
		{
			input: `// just
		// comments
		// here`,
			want: `null`,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := Parse(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}
