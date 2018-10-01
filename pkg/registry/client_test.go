package registry

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

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

func TestEnforceWhitelist(t *testing.T) {
	run := func(whitelist *[]string, extensions []string, want []string) {
		t.Helper()
		if whitelist != nil {
			conf.Mock(&schema.SiteConfiguration{Extensions: &schema.Extensions{RemoteWhitelist: *whitelist}})
			defer conf.Mock(nil)
		}
		var xs []*Extension
		for _, id := range extensions {
			xs = append(xs, &Extension{ExtensionID: id})
		}
		got := []string{}
		for _, x := range Whitelist(xs) {
			got = append(got, x.ExtensionID)
		}
		if !sameElements(got, want) {
			t.Errorf("want %+v got %+v", want, got)
		}
	}

	run(nil, []string{}, []string{})
	run(nil, []string{"a"}, []string{"a"})
	run(&[]string{}, []string{}, []string{})
	run(&[]string{"a"}, []string{}, []string{})
	run(&[]string{}, []string{"a"}, []string{})
	run(&[]string{"a"}, []string{"b"}, []string{})
	run(&[]string{"a"}, []string{"a"}, []string{"a"})
	run(&[]string{"b", "c"}, []string{"a", "b", "c"}, []string{"b", "c"})
}

func TestIsWhitelisted(t *testing.T) {
	defer conf.Mock(nil)

	if !IsWhitelisted(&Extension{ExtensionID: "a"}) {
		t.Errorf("expected %q to be whitelisted", "a")
	}

	conf.Mock(&schema.SiteConfiguration{Extensions: &schema.Extensions{RemoteWhitelist: []string{}}})
	if IsWhitelisted(&Extension{ExtensionID: "a"}) {
		t.Errorf("expected %q to not be whitelisted", "a")
	}

	conf.Mock(&schema.SiteConfiguration{Extensions: &schema.Extensions{RemoteWhitelist: []string{"a"}}})
	if !IsWhitelisted(&Extension{ExtensionID: "a"}) {
		t.Errorf("expected %q to be whitelisted", "a")
	}
}
