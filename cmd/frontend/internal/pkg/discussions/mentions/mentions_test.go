package mentions

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name, input string
		want        []string
	}{
		{
			name:  "basic",
			input: "@bob",
			want:  []string{"bob"},
		},
		{
			name:  "complex",
			input: "hello @sally world\t@bob-1233%#!%#$1\n@kim",
			want:  []string{"sally", "bob-1233%#!%#$1", "kim"},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := Parse(tst.input)
			if !reflect.DeepEqual(got, tst.want) {
				t.Fatalf("got %q want %q", got, tst.want)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_377(size int) error {
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
