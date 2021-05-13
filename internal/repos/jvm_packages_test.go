package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestJvmPackagesClone(t *testing.T) {
	var c schema.JvmPackagesConnection
	c.Maven.Repositories = []string{"central"}
	dep := reposource.ParseMavenDependency("junit:junit:4.13.2")
	x, err := coursier.FetchSources(context.Background(), &c, dep)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(x)
}
