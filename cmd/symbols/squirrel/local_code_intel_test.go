package squirrel

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestLocalCodeIntel(t *testing.T) {
	goSource := `
func main() {
	z := 4
	f(z)
}

func f(x int) {
	fmt.Println(x)
}
`

	symbolNameToRefCount := map[string]int{
		"main": 1,
		"z":    2,
		"f":    1, // TODO allow functions to escape their immediate scope (should be 2)
		"x":    2,
	}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		return []byte(goSource), nil
	}

	squirrel := NewSquirrelService(readFile, nil)
	defer squirrel.Close()

	payload, err := squirrel.localCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: "test.go"})
	fatalIfError(t, err)

	for _, symbol := range payload.Symbols {
		// Check if len(symbol.Refs) is equal to symbolNameToRefCount[symbol.Name]
		if wantRefCount, ok := symbolNameToRefCount[symbol.Name]; ok && len(symbol.Refs) != wantRefCount {
			t.Fatalf("symbol %s has %d refs, want %d", symbol.Name, len(symbol.Refs), wantRefCount)
		}
	}
}
