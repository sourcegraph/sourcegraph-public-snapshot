package responses

// ResponseObject holds fields which are common for all objects returned from
// *.search API methods.
type ResponseObject struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	PHID string `json:"phid"`
}

// SearchCursor holds paging information on responses from *.search API methods.
type SearchCursor struct {
	Limit  uint64 `json:"limit"`
	After  string `json:"after"`
	Before string `json:"before"`
}
