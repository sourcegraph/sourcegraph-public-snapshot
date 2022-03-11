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
		"f":    2,
		"x":    2,
	}

	symbolNameToLocal := map[string]bool{
		"z": true,
		"f": false,
	}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		return []byte(goSource), nil
	}

	squirrel := NewSquirrelService(readFile, nil)
	defer squirrel.Close()

	payload, err := squirrel.localCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: "test.go"})
	fatalIfError(t, err)

	for _, symbol := range payload.Symbols {
		if wantRefCount, ok := symbolNameToRefCount[symbol.Name]; ok && len(symbol.Refs) != wantRefCount {
			t.Fatalf("symbol %s has %d refs, want %d", symbol.Name, len(symbol.Refs), wantRefCount)
		}

		if wantLocal, ok := symbolNameToLocal[symbol.Name]; ok && symbol.Local != wantLocal {
			t.Fatalf("symbol %s is %t, want %t", symbol.Name, symbol.Local, wantLocal)
		}
	}
}

func TestLocalCodeIntelExports(t *testing.T) {
	goSource := `
func main() {
	x := 5
	fmt.Println(x)
}
`

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		return []byte(goSource), nil
	}

	squirrel := NewSquirrelService(readFile, nil)
	defer squirrel.Close()

	payload, err := squirrel.localCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: "test.go"})
	fatalIfError(t, err)

	nameToSymbol := map[string]types.Symbol{}
	for _, symbol := range payload.Symbols {
		nameToSymbol[symbol.Name] = symbol
	}

	if symbol, ok := nameToSymbol["x"]; !ok {
		t.Fatalf("symbol x not found")
	} else if symbol.Local == false {
		t.Fatalf("expected symbol x to be local")
	}

	if symbol, ok := nameToSymbol["main"]; !ok {
		t.Fatalf("symbol main not found")
	} else if symbol.Local == true {
		t.Fatalf("expected symbol main to not be local")
	}
}
