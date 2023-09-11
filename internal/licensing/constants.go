package licensing

import "time"

const (
	LicenseCheckInterval    = 12 * time.Hour
	LicenseValidityStoreKey = "licensing:is_license_valid"
	LicenseInvalidReason    = "licensing:license_invalid_reason"
)
