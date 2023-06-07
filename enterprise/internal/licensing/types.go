package licensing

type LicenseCheckRequestParams struct {
	ClientSiteID string `json:"siteID"`
}

type LicenseCheckResponseData struct {
	IsValid bool   `json:"is_valid,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type LicenseCheckResponse struct {
	Error string                    `json:"error,omitempty"`
	Data  *LicenseCheckResponseData `json:"data"`
}
