package lockfiles

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

func TestDependencyGraph(t *testing.T) {
	a := &testPackageVersion{name: "a", version: "1.0.0"}
	b := &testPackageVersion{name: "b", version: "1.0.0"}
	c := &testPackageVersion{name: "c", version: "1.0.0"}
	d := &testPackageVersion{name: "d", version: "1.0.0"}
	e := &testPackageVersion{name: "e", version: "1.0.0"}
	f := &testPackageVersion{name: "f", version: "1.0.0"}
	g := &testPackageVersion{name: "g", version: "1.0.0"}

	dg := newDependencyGraph()
	dg.addPackage(a)
	dg.addPackage(b)
	dg.addPackage(c)
	dg.addPackage(d)
	dg.addPackage(e)
	dg.addPackage(f)
	dg.addPackage(f)

	// a -> b -> d
	//   -> c -> e -> f
	//   -> g
	dg.addDependency(a, b)
	dg.addDependency(a, c)
	dg.addDependency(a, g)

	dg.addDependency(b, d)

	dg.addDependency(c, e)
	dg.addDependency(e, f)

	want := `` +
		`test/a:
	test/b:
		test/d
	test/c:
		test/e:
			test/f
	test/g
`

	got := dg.String()
	fmt.Println(got)

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("+want,-got\n%s", d)
	}

}

var _ reposource.PackageVersion = &testPackageVersion{}

type testPackageVersion struct {
	name    string
	version string
}

func (t *testPackageVersion) PackageVersionSyntax() string { return t.name }
func (t *testPackageVersion) PackageSyntax() string        { return t.name }
func (t *testPackageVersion) RepoName() api.RepoName       { return api.RepoName("test/" + t.name) }
func (t *testPackageVersion) PackageVersion() string       { return t.version }
func (t *testPackageVersion) Scheme() string               { return "test" }
func (t *testPackageVersion) Description() string          { return "" }
func (t *testPackageVersion) GitTagFromVersion() string    { return t.version }
func (t *testPackageVersion) Less(other reposource.PackageVersion) bool {
	return t.PackageSyntax() < other.PackageSyntax()
}
