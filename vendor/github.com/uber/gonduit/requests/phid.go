package requests

// PHIDLookupRequest represents a request to phid.lookup.
type PHIDLookupRequest struct {
	Names []string `json:"names"`
	Request
}

// PHIDQueryRequest represents a request to phid.query.
type PHIDQueryRequest struct {
	PHIDs []string `json:"phids"`
	Request
}
