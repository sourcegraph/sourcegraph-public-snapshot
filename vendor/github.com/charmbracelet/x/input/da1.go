package input

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesEvent represents a primary device attributes event.
type PrimaryDeviceAttributesEvent []uint

func parsePrimaryDevAttrs(csi *ansi.CsiSequence) Event {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesEvent, len(csi.Params))
	csi.Range(func(i int, p int, hasMore bool) bool {
		if !hasMore {
			da1[i] = uint(p)
		}
		return true
	})
	return da1
}
