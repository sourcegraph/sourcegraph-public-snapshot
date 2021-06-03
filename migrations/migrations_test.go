package migrations

import (
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestIDConstraints(t *testing.T) {
	cases := []struct {
		Name string
		FS   fs.FS
	}{
		{Name: "frontend", FS: Frontend},
		{Name: "codeintel", FS: CodeIntel},
		{Name: "codeinsights", FS: CodeInsights},
	}

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

			var ids []int
			for id, names := range byID {
				if len(names) > 1 {
					t.Errorf("multiple migrations with ID %d: %s", id, strings.Join(names, " "))
				}

				ids = append(ids, id)
			}
			sort.Ints(ids)

			for i, id := range ids {
				if i != 0 && ids[i-1]+1 != id {
					// Check if we are using sequential migrations.
					t.Errorf("gap in migrations between %s and %s", byID[ids[i-1]][0], byID[id][0])
				}
			}
		})
	}
}
