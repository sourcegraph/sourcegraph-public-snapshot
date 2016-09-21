package langp

import (
	"path"
	"testing"
)

func TestResolveRepoAlias(t *testing.T) {
	check := func(old, new string) {
		x := ResolveRepoAlias(old)
		if x != new {
			t.Errorf("ResolveRepoAlias(%s) == %s != %s", old, x, new)
		}
		x = UnresolveRepoAlias(new)
		if x != old {
			t.Errorf("UnresolveRepoAlias(%s) == %s != %s", new, x, old)
		}
	}
	for _, r := range repoAliases {
		check(r.OldPrefix, r.NewPrefix)
		check(path.Join(r.OldPrefix, "foo"), path.Join(r.NewPrefix, "foo"))
	}
	cases := []struct{ old, new string }{
		{
			"github.com/kubernetes/kubernetes/pkg/api",
			"k8s.io/kubernetes/pkg/api",
		},
		{
			"github.com/kubernetes/kops/upup/pkg/api",
			"k8s.io/kops/upup/pkg/api",
		},
		{
			"github.com/golang/go",
			"github.com/golang/go",
		},
		{
			"github.com/golang/crypto",
			"golang.org/x/crypto",
		},
	}
	for _, c := range cases {
		check(c.old, c.new)
	}
}
