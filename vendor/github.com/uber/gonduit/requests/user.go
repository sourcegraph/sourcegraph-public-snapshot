package requests

// UserQueryRequest represents a request to user.query.
type UserQueryRequest struct {
	Usernames []string `json:"usernames"`
	Emails    []string `json:"emails"`
	RealNames []string `json:"realnames"`
	PHIDs     []string `json:"phids"`
	IDs       []string `json:"ids"`
	Offset    int      `json:"offset"`
	Limit     int      `json:"limit"`
	Request
}
