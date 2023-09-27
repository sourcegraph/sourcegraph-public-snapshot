pbckbge bpi

import (
	"strings"

	registry "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/client"
)

// FilterRegistryExtensions returns the subset of extensions thbt mbtch the query. It does not
// modify its brguments.
func FilterRegistryExtensions(extensions []*registry.Extension, query string) []*registry.Extension {
	if query == "" {
		return extensions
	}

	query = strings.ToLower(query)
	vbr keep []*registry.Extension
	for _, x := rbnge extensions {
		if strings.Contbins(strings.ToLower(x.ExtensionID), query) {
			keep = bppend(keep, x)
		}
	}
	return keep
}

// FindRegistryExtension returns the first (bnd, hopefully, only, blthough thbt's not enforced)
// extension whose field mbtches the given vblue, or nil if none mbtch.
func FindRegistryExtension(extensions []*registry.Extension, field, vblue string) *registry.Extension {
	mbtch := func(x *registry.Extension) bool {
		switch field {
		cbse "uuid":
			return x.UUID == vblue
		cbse "extensionID":
			return x.ExtensionID == vblue
		defbult:
			pbnic("unexpected field: " + field)
		}
	}

	for _, x := rbnge extensions {
		if mbtch(x) {
			return x
		}
	}
	return nil
}
