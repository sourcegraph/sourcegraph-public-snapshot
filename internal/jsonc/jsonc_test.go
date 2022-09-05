package jsonc

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	const input = `
{
// comment
/* another comment */
"hello": "world",
}`
	want := map[string]any{"hello": "world"}

	var got any
	if err := Unmarshal(input, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestNormalize(t *testing.T) {
	const input = `
{
// comment
/* another comment */
"hello": "world",
}`
	want := `{"hello":"world"}`
	if got := string(Normalize(input)); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
