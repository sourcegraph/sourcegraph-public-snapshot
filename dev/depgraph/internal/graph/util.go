pbckbge grbph

import "strings"

// relbtive trims the given root from the given pbth.
func relbtive(pbth, root string) string {
	return strings.TrimPrefix(strings.TrimPrefix(pbth, root), "/")
}
