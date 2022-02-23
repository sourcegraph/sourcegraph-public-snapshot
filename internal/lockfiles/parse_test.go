package lockfiles

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

func TestParse(t *testing.T) {
	tests := []struct {
		file string
		data string
		want []reposource.PackageDependency
	}{
		{
			file: "package-lock.json",
			data: `{"dependencies": {
        "nan": {"version": "2.15.0"},
        "@octokit/request": {"version": "5.6.2"}
      }}`,
			want: []reposource.PackageDependency{
				npmDependency(t, "@octokit/request@5.6.2"),
				npmDependency(t, "nan@2.15.0"),
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			got, err := Parse(test.file, []byte(test.data))
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].PackageManagerSyntax() < got[j].PackageManagerSyntax()
			})

			comparer := cmp.Comparer(func(a, b reposource.PackageDependency) bool {
				return a.PackageManagerSyntax() == b.PackageManagerSyntax()
			})

			if diff := cmp.Diff(test.want, got, comparer); diff != "" {
				t.Fatalf("dependency mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func npmDependency(t testing.TB, dep string) *reposource.NPMDependency {
	t.Helper()

	d, err := reposource.ParseNPMDependency(dep)
	if err != nil {
		t.Fatal(err)
	}

	return d
}
