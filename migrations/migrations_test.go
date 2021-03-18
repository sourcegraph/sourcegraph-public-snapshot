package migrations

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

func TestIDConstraints(t *testing.T) {
	cases := []struct {
		Name  string
		FS    fs.FS
		First int
	}{{
		Name:  "codeinsights",
		FS:    CodeInsights,
		First: 1000000000,
	}, {
		Name:  "codeintel",
		FS:    CodeIntel,
		First: 1000000000,
	}, {
		Name:  "frontend",
		FS:    Frontend,
		First: 1528395733,
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ups, err := fs.Glob(c.FS, "*.up.sql")
			if err != nil {
				t.Fatal(err)
			}

			if len(ups) == 0 {
				t.Fatal("no up migrations found")
			}

			byID := map[int][]string{}
			for _, name := range ups {
				id, err := strconv.Atoi(name[:strings.IndexByte(name, '_')])
				if err != nil {
					t.Fatalf("failed to parse name %q: %v", name, err)
				}
				byID[id] = append(byID[id], name)
			}

			for id, names := range byID {
				// Check if we are using sequential migrations from a certain point.
				if _, hasPrev := byID[id-1]; id > c.First && !hasPrev {
					t.Errorf("migration with ID %d exists, but previous one (%d) does not", id, id-1)
				}
				if len(names) > 1 {
					t.Errorf("multiple migrations with ID %d: %s", id, strings.Join(names, " "))
				}
			}
		})
	}
}
