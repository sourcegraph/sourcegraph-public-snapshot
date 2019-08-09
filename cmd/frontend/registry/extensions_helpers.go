package registry

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

// FilterRegistryExtensions returns the subset of extensions that match the query. It does not
// modify its arguments.
func FilterRegistryExtensions(extensions []*registry.Extension, query string) []*registry.Extension {
	if query == "" {
		return extensions
	}

	query = strings.ToLower(query)
	var keep []*registry.Extension
	for _, x := range extensions {
		if strings.Contains(strings.ToLower(x.ExtensionID), query) {
			keep = append(keep, x)
		}
	}
	return keep
}

// FindRegistryExtension returns the first (and, hopefully, only, although that's not enforced)
// extension whose field matches the given value, or nil if none match.
func FindRegistryExtension(extensions []*registry.Extension, field, value string) *registry.Extension {
	match := func(x *registry.Extension) bool {
		switch field {
		case "uuid":
			return x.UUID == value
		case "extensionID":
			return x.ExtensionID == value
		default:
			panic("unexpected field: " + field)
		}
	}

	for _, x := range extensions {
		if match(x) {
			return x
		}
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_428(size int) error {
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
