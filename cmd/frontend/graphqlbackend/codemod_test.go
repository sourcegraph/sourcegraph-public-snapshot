package graphqlbackend

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
)

func TestCodemod_validateArgsNoRegex(t *testing.T) {
	q, _ := query.ParseAndCheck("re.*gex")
	_, err := validateQuery(q)
	if err == nil {
		t.Fatalf("Expected query %v to fail", q)
	}
	if !strings.HasPrefix(err.Error(), "this looks like a regex search pattern.") {
		t.Fatalf("%v expected complaint about regex pattern. Got %s", q, err)
	}
}

func TestCodemod_validateArgsOk(t *testing.T) {
	q, _ := query.ParseAndCheck(`"not regex"`)
	_, err := validateQuery(q)
	if err != nil {
		t.Fatalf("Expected query %v to to be OK", q)
	}
}

func TestCodemod_resolver(t *testing.T) {
	raw := &rawCodemodResult{
		URI:  "",
		Diff: "Not a valid diff",
	}
	_, err := toMatchResolver("", raw)
	if err == nil {
		t.Fatalf("Expected invalid diff for %v", raw.Diff)
	}
	if !strings.HasPrefix(err.Error(), "Invalid diff") {
		t.Fatalf("Expected error %q", err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_118(size int) error {
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
