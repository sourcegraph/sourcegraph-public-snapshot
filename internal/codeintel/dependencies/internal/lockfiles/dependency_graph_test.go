package lockfiles

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

func TestDependencyGraph(t *testing.T) {
	a := &testDep{"a"}
	b := &testDep{"b"}
	c := &testDep{"c"}
	d := &testDep{"d"}
	e := &testDep{"e"}
	f := &testDep{"f"}
	g := &testDep{"g"}

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

var _ reposource.PackageDependency = &testDep{}

type testDep struct{ name string }

func (t *testDep) PackageManagerSyntax() string { return t.name }
func (t *testDep) PackageSyntax() string        { return t.name }
func (t *testDep) RepoName() api.RepoName       { return api.RepoName("test/" + t.name) }
func (t *testDep) PackageVersion() string       { return "1.0.0" }
func (t *testDep) Scheme() string               { return "test" }
func (t *testDep) Description() string          { return "" }
func (t *testDep) GitTagFromVersion() string    { return "1.0.0" }
func (t *testDep) Less(other reposource.PackageDependency) bool {
	return t.PackageSyntax() < other.PackageSyntax()
}
