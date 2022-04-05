package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update generated queries")

func TestLoadQueries(t *testing.T) {
	for _, env := range []string{"", "cloud", "dogfood"} {
		t.Run(env, func(t *testing.T) {
			c, err := loadQueries("")
			if err != nil {
				t.Fatal(err)
			}

			if len(c.Groups) < 1 {
				t.Fatal("expected atleast 1 group")
			}

			if len(c.Groups[0].Queries) < 2 {
				t.Fatal("expected atleast 2 queries")
			}

			names := map[string]bool{}
			for _, q := range c.Groups[0].Queries {
				if names[q.Name] {
					t.Fatalf("name %q is not unique", q.Name)
				}
				names[q.Name] = true
			}

			if testing.Verbose() {
				for _, q := range c.Groups[0].Queries {
					t.Logf("% -25s %s", q.Name, q.Query)
				}
			}
		})
	}
}

func TestDogfoodQueries(t *testing.T) {
	got := generateQueries(t, "dogfood",
		// we want to target gigarepo on dogfood
		`repo:^github\.com/sgtest/megarepo$`, `repo:^gigarepo$`,
	)

	// Just confirm we don't mention megarepo since we want gigarepo
	if strings.Contains(got, "megarepo") {
		t.Fatal("megarepo should not appear in dogfood queries")
	}
}

// generateQueries will generate a copy of queries_cloud.txt with the
// replacements applied. It returns the document without the header explaining
// the document.
func generateQueries(t *testing.T, name string, replacements ...string) string {
	from := "queries_cloud.txt"
	to := strings.ReplaceAll(from, "cloud", name)
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	_, _ = fmt.Fprintf(&out, "## @generated from %s for %s\n", from, name)

	for i := 0; i < len(replacements); i += 2 {
		old, new := replacements[i], replacements[i+1]
		_, _ = fmt.Fprintf(&out, "## replace %s -> %s\n", old, new)
		b = bytes.ReplaceAll(b, []byte(old), []byte(new))
	}

	_, _ = out.WriteString("\n\n")
	_, _ = out.Write(b)

	want := out.Bytes()
	if *update {
		err := os.WriteFile(to, want, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}

	got, err := os.ReadFile(to)
	if err != nil {
		t.Fatal(err)
	}

	if d := cmp.Diff(string(want), string(got)); d != "" {
		t.Fatalf("mismatch. To update run 'go test -update' (-want +got):\n%s", d)
	}

	return string(b)
}
