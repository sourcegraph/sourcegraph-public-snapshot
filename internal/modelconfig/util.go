package modelconfig

import (
	"strings"

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
// "${x2}/bar;  baz!" => "__x2__bar___baz_"
func SanitizeResourceName(name string) string {
	// Start by just converting everything to lower-case, so that
	// the result doesn't look too awkward.
	name = strings.ToLower(name)

	sanitizedName := []byte(name)
	switch len(name) {
	case 0:
		return string(sanitizedName)
	case 1:
		if !resourceIDFirstLastRE.MatchString(name) {
			sanitizedName[0] = '_'
		}
		return string(sanitizedName)
	default:
		l := len(name)
		// Check the first and last characters.
		firstOK := resourceIDFirstLastRE.MatchString(name[:1])
		lastOK := resourceIDFirstLastRE.MatchString(name[l-1:])
		if !firstOK {
			sanitizedName[0] = '_'
		}
		if !lastOK {
			sanitizedName[l-1] = '_'
		}

		// Check the middle. We do this one byte at a time, which will
		// fail for any UTF codepoints requiring more than one byte.
		oneByteStr := make([]byte, 1)
		for i := 1; i < l-1; i++ {
			oneByteStr[0] = sanitizedName[i]
			if !resourceIDMiddleRE.MatchString(string(oneByteStr)) {
				sanitizedName[i] = '_'
			}
		}

		return string(sanitizedName)
	}
}
