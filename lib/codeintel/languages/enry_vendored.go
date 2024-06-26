package languages

import "strings"

// This file contains functions private functions
// vendored from the go-enry codebase.

// convertToAliasKey is vendored from go-enry to make sure
// we're normalizing strings the same way.
func convertToAliasKey(langName string) string {
	ak := strings.SplitN(langName, `,`, 2)[0]
	ak = strings.Replace(ak, ` `, `_`, -1)
	ak = strings.ToLower(ak)
	return ak
}
