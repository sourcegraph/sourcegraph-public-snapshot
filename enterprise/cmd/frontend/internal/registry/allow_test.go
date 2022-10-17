package registry

import (
	"reflect"
	"sort"
	"testing"

	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestIsRemoteExtensionAllowed(t *testing.T) {
	defer licensing.TestingSkipFeatureChecks()()
	defer conf.Mock(nil)

	if !frontendregistry.IsRemoteExtensionAllowed("a") {
		t.Errorf("want %q to be allowed", "a")
	}

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{AllowRemoteExtensions: nil}}})
	if !frontendregistry.IsRemoteExtensionAllowed("a") {
		t.Errorf("want %q to be allowed", "a")
	}

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{AllowRemoteExtensions: []string{}}}})
	if frontendregistry.IsRemoteExtensionAllowed("a") {
		t.Errorf("want %q to be disallowed", "a")
	}

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{AllowRemoteExtensions: []string{"a"}}}})
	if !frontendregistry.IsRemoteExtensionAllowed("a") {
		t.Errorf("want %q to be allowed", "a")
	}
}

func sameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))

	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}

func TestFilterRemoteExtensions(t *testing.T) {
	defer licensing.TestingSkipFeatureChecks()()

	run := func(extensionConfig *schema.Extensions, extensions []*registry.Extension, want []string) {
		t.Helper()

		if extensionConfig != nil {
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: extensionConfig}})
			defer conf.Mock(nil)
		}

		var xs []*registry.Extension
		for _, x := range extensions {
			xs = append(xs, x)
		}
		got := []string{}
		for _, x := range frontendregistry.FilterRemoteExtensions(xs) {
			got = append(got, x.ExtensionID)
		}
		if !sameElements(got, want) {
			t.Errorf("want %+v got %+v", want, got)
		}
	}

	run(nil, []*registry.Extension{}, []string{})
	run(nil, []*registry.Extension{{ExtensionID: "a"}}, []string{"a"})
	run(&schema.Extensions{AllowRemoteExtensions: []string{}}, []*registry.Extension{}, []string{})
	run(&schema.Extensions{AllowRemoteExtensions: []string{"a"}}, []*registry.Extension{}, []string{})
	run(&schema.Extensions{AllowRemoteExtensions: []string{}}, []*registry.Extension{{ExtensionID: "a"}}, []string{})
	run(&schema.Extensions{AllowRemoteExtensions: []string{"a"}}, []*registry.Extension{{ExtensionID: "b"}}, []string{})
	run(&schema.Extensions{AllowRemoteExtensions: []string{"a"}}, []*registry.Extension{{ExtensionID: "a"}}, []string{"a"})
	run(&schema.Extensions{AllowRemoteExtensions: []string{"b", "c"}}, []*registry.Extension{{ExtensionID: "a"}, {ExtensionID: "b"}, {ExtensionID: "c"}}, []string{"b", "c"})
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "tobias",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/users/tobias",
				},
			},
			{
				ExtensionID: "c",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
		},
		[]string{"a", "c"},
	)
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
			AllowRemoteExtensions:                  []string{"b", "c"},
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "c",
				Publisher: registry.Publisher{
					Name: "tobias",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/users/tobias",
				},
			},
		},
		[]string{"b", "c"},
	)
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
			AllowRemoteExtensions:                  []string{},
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
		},
		[]string{},
	)
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
			RemoteRegistry:                         "https://sourcegraph.com/.api/registry",
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "tobias",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/users/tobias",
				},
			},
			{
				ExtensionID: "c",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
		},
		[]string{"a", "c"},
	)
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
			RemoteRegistry:                         false,
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "tobias",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/users/tobias",
				},
			},
			{
				ExtensionID: "c",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
		},
		[]string{"a", "b", "c"},
	)
	run(
		&schema.Extensions{
			AllowOnlySourcegraphAuthoredExtensions: true,
			RemoteRegistry:                         "https://some-remote-registry.com/.api/registry",
		},
		[]*registry.Extension{
			{
				ExtensionID: "a",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
			{
				ExtensionID: "b",
				Publisher: registry.Publisher{
					Name: "tobias",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/users/tobias",
				},
			},
			{
				ExtensionID: "c",
				Publisher: registry.Publisher{
					Name: "sourcegraph",
					URL:  "https://sourcegraph.com/extensions/registry/publishers/organizations/sourcegraph",
				},
			},
		},
		[]string{"a", "b", "c"},
	)
}
