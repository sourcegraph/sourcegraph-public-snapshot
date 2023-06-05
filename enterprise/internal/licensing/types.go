package licensing

type LicenseCheckRequestParams struct {
	ClientSiteID string `json:"siteID"`
}

type LicenseCheckResponseData struct {
	IsValid bool    `json:"is_valid"`
	Reason  *string `json:"reason"`
}

type LicenseCheckResponse struct {
	Error *string                   `json:"error"`
	Data  *LicenseCheckResponseData `json:"data"`
}
