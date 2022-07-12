package lockfiles

import (
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

func TestDependencyGraph(t *testing.T) {
	a := &testPkg{name: "a", version: "1.0.0"}
	b := &testPkg{name: "b", version: "1.0.0"}
	c := &testPkg{name: "c", version: "1.0.0"}
	d := &testPkg{name: "d", version: "1.0.0"}
	e := &testPkg{name: "e", version: "1.0.0"}
	f := &testPkg{name: "f", version: "1.0.0"}
	g := &testPkg{name: "g", version: "1.0.0"}

	gold := goldie.New(t, goldie.WithFixtureDir("testdata/dependencygraph"))

	t.Run("normal", func(t *testing.T) {
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

		gold.AssertJson(t, "normal", dg.AsMap())
	})

	t.Run("circular", func(t *testing.T) {
		dg := newDependencyGraph()
		dg.addPackage(a)
		dg.addPackage(b)
		dg.addPackage(c)
		dg.addPackage(d)
		dg.addPackage(e)
		dg.addPackage(f)
		dg.addPackage(f)

		// a -> b -> d
		//   -> c -> e -> f -> c     <-- f depends on c
		//   -> g
		dg.addDependency(a, b)
		dg.addDependency(a, c)
		dg.addDependency(a, g)

		dg.addDependency(b, d)

		dg.addDependency(c, e)
		dg.addDependency(e, f)
		dg.addDependency(f, c)

		gold.AssertJson(t, "circular", dg.AsMap())
	})

	t.Run("circular root", func(t *testing.T) {
		dg := newDependencyGraph()
		dg.addPackage(a)
		dg.addPackage(b)
		dg.addPackage(c)
		dg.addPackage(d)
		dg.addPackage(e)
		dg.addPackage(f)
		dg.addPackage(f)

		// a -> b -> d
		//   -> c -> e -> f -> a     <-- f depends on a
		//   -> g
		dg.addDependency(a, b)
		dg.addDependency(a, c)
		dg.addDependency(a, g)

		dg.addDependency(b, d)

		dg.addDependency(c, e)
		dg.addDependency(e, f)
		dg.addDependency(f, a)

		gold.AssertJson(t, "circular-root", dg.AsMap())
	})
}

var _ reposource.VersionedPackage = &testPkg{}

type testPkg struct {
	name    string
	version string
}

func (t *testPkg) VersionedPackageSyntax() string { return t.name }
func (t *testPkg) PackageSyntax() string          { return t.name }
func (t *testPkg) RepoName() api.RepoName         { return api.RepoName("test/" + t.name) }
func (t *testPkg) PackageVersion() string         { return t.version }
func (t *testPkg) Scheme() string                 { return "test" }
func (t *testPkg) Description() string            { return "" }
func (t *testPkg) GitTagFromVersion() string      { return t.version }
func (t *testPkg) Less(other reposource.VersionedPackage) bool {
	return t.PackageSyntax() < other.VersionedPackageSyntax()
}
