package squirrel

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestLocalCodeIntel(t *testing.T) {
	type Wants struct {
		hasDefinition bool
		refCount      int
		isLocal       bool
	}

	type Test struct {
		name        string
		source      string
		nameToWants map[string]Wants
	}

	tests := []Test{
		{
			name: "go",
			source: `
const c = 1
var v = 2
type t1 = int
type t2 int

func f(p int) {
	i := 4
	fmt.Println(p, i)
}

func (t t2) m() {
	f(3)
}
`,
			nameToWants: map[string]Wants{
				"m":   {hasDefinition: true, refCount: 1, isLocal: false},
				"i":   {hasDefinition: true, refCount: 2, isLocal: true},
				"f":   {hasDefinition: true, refCount: 2, isLocal: false},
				"c":   {hasDefinition: true, refCount: 1, isLocal: false},
				"v":   {hasDefinition: true, refCount: 1, isLocal: false},
				"t1":  {hasDefinition: true, refCount: 1, isLocal: false},
				"t2":  {hasDefinition: true, refCount: 2, isLocal: false},
				"t":   {hasDefinition: true, refCount: 1, isLocal: true},
				"p":   {hasDefinition: true, refCount: 2, isLocal: true},
				"fmt": {hasDefinition: false, refCount: 1, isLocal: false},
			},
		},
	}

	for _, test := range tests {
		readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
			return []byte(test.source), nil
		}

		squirrel := NewSquirrelService(readFile, nil)
		defer squirrel.Close()

		payload, err := squirrel.localCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: "test.go"})
		fatalIfError(t, err)

		for _, symbol := range payload.Symbols {
			wants, ok := test.nameToWants[symbol.Name]
			if !ok {
				continue
			}

			if wants.hasDefinition != (symbol.Def != nil) {
				t.Fatalf("symbol %q has definition? got %v, want %v", symbol.Name, symbol.Def != nil, wants.hasDefinition)
			}

			if wants.refCount != len(symbol.Refs) {
				t.Fatalf("symbol %q ref count got %d, want %d", symbol.Name, len(symbol.Refs), wants.refCount)
			}

			if wants.isLocal != symbol.Local {
				t.Fatalf("symbol %q isLocal got %v, want %v", symbol.Name, symbol.Local, wants.isLocal)
			}
		}

		// Make sure we didn't miss any.
	nextName:
		for name := range test.nameToWants {
			for _, symbol := range payload.Symbols {
				if symbol.Name == name {
					continue nextName
				}
			}
			t.Fatalf("expected local code intel to include symbol %s", name)
		}
	}
}
