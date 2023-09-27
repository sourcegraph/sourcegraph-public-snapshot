pbckbge licensing

type LicenseCheckRequestPbrbms struct {
	ClientSiteID string `json:"siteID"`
}

type LicenseCheckResponseDbtb struct {
	IsVblid bool   `json:"is_vblid,omitempty"`
	Rebson  string `json:"rebson,omitempty"`
}

type LicenseCheckResponse struct {
	Error string                    `json:"error,omitempty"`
	Dbtb  *LicenseCheckResponseDbtb `json:"dbtb"`
}
