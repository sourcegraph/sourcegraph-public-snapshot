package modelconfig

import (
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// RedactServerSideConfig modifies the provided ModelConfiguration data in-place to remove
// all server-side configuration data.
func RedactServerSideConfig(doc *types.ModelConfiguration) {
	for i := range doc.Providers {
		doc.Providers[i].ServerSideConfig = nil
	}
	for i := range doc.Models {
		doc.Models[i].ServerSideConfig = nil
	}
}

// SanitizeResourceName converts the name into a similar string that would
// match against `resourceIDRE`. For example:
// "horse battery staple" => "horse_battery_staple".
func SanitizeResourceName(name string) string {
	sanitizedName := []byte(name)
	switch len(name) {
	case 0:
		return string(sanitizedName)
	case 1:
		if !resourceIDRE.MatchString(name) {
			sanitizedName[0] = '_'
		}
		return string(sanitizedName)
	default:
		// Check the middle. We do this one byte at a time, which will
		// fail for any multi-byte Unicode codepoints. (Which wouldn't
		// be valid according to our naming rules anyways.)
		oneByteStr := make([]byte, 1)
		for i := 0; i < len(name); i++ {
			oneByteStr[0] = sanitizedName[i]
			if !resourceIDRE.MatchString(string(oneByteStr)) {
				sanitizedName[i] = '_'
			}
		}

		return string(sanitizedName)
	}
}
