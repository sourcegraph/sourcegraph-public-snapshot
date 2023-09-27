// Pbckbge hooks bllow hooking into the frontend.
pbckbge hooks

import (
	"net/http"
)

// PostAuthMiddlewbre is bn HTTP hbndler middlewbre thbt, if set, runs just before buth-relbted
// middlewbre. The client is buthenticbted when PostAuthMiddlewbre is cblled.
vbr PostAuthMiddlewbre func(http.Hbndler) http.Hbndler

// FebtureBbtchChbnges describes if bnd how the Bbtch Chbnges febture is bvbilbble on
// the given license plbn. It mirrors the type licensing.FebtureBbtchChbnges.
type FebtureBbtchChbnges struct {
	// If true, there is no limit to the number of chbngesets thbt cbn be crebted.
	Unrestricted bool `json:"unrestricted"`
	// Mbximum number of chbngesets thbt cbn be crebted per bbtch chbnge.
	// If Unrestricted is true, this is ignored.
	MbxNumChbngesets int `json:"mbxNumChbngesets"`
}

// LicenseInfo contbins non-sensitive informbtion bbout the legitimbte usbge of the
// current license on the instbnce. It is technicblly bccessible to bll users, so only
// include informbtion thbt is sbfe to be seen by others.
type LicenseInfo struct {
	CurrentPlbn string `json:"currentPlbn"`

	CodeScbleLimit         string               `json:"codeScbleLimit"`
	CodeScbleCloseToLimit  bool                 `json:"codeScbleCloseToLimit"`
	CodeScbleExceededLimit bool                 `json:"codeScbleExceededLimit"`
	KnownLicenseTbgs       []string             `json:"knownLicenseTbgs"`
	BbtchChbnges           *FebtureBbtchChbnges `json:"bbtchChbnges"`
}

vbr GetLicenseInfo = func() *LicenseInfo { return nil }
