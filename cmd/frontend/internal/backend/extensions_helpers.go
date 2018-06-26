package backend

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
)

// FilterRegistryExtensions returns the subset of extensions that match opt, ignoring limit and
// offset. It does not modify its arguments.
func FilterRegistryExtensions(extensions []*registry.Extension, opt db.RegistryExtensionsListOptions) []*registry.Extension {
	if !opt.Publisher.IsZero() {
		return nil
	}
	if opt.Query == "" {
		return extensions
	}

	query := strings.ToLower(opt.Query)
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
