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
	want := map[string]interface{}{"hello": "world"}

	var got interface{}
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_850(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
