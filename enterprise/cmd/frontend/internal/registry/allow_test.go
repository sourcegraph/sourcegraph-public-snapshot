package registry

import (
	"reflect"
	"sort"
	"testing"

	frontendregistry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestIsRemoteExtensionAllowed(t *testing.T) {
	licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
		return &license.Info{Tags: licensing.EnterpriseTags}, "test-signature", nil
	}
	defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
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
	licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
		return &license.Info{Tags: licensing.EnterpriseTags}, "test-signature", nil
	}
	defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

	run := func(allowRemoteExtensions *[]string, extensions []string, want []string) {
		t.Helper()
		if allowRemoteExtensions != nil {
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{AllowRemoteExtensions: *allowRemoteExtensions}}})
			defer conf.Mock(nil)
		}
		var xs []*registry.Extension
		for _, id := range extensions {
			xs = append(xs, &registry.Extension{ExtensionID: id})
		}
		got := []string{}
		for _, x := range frontendregistry.FilterRemoteExtensions(xs) {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_673(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
